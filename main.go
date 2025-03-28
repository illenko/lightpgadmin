package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	_ "github.com/lib/pq"
)

type DatabaseSource struct {
	DB *sql.DB
}

type DatabaseMetadata struct {
	Tables []TableMetadata `json:"tables"`
}

type TableMetadata struct {
	Name       string   `json:"name"`
	Columns    []Column `json:"columns"`
	PrimaryKey []string `json:"primaryKey"`
}

type Column struct {
	Name     string `json:"name"`
	DataType string `json:"dataType"`
	Nullable string `json:"nullable"`
}

func NewDatabaseSource(connStr string) (*DatabaseSource, error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	if err = db.Ping(); err != nil {
		return nil, err
	}

	fmt.Println("Successfully connected")

	return &DatabaseSource{DB: db}, nil
}

func (ds *DatabaseSource) Close() error {
	return ds.DB.Close()
}

func (ds *DatabaseSource) GetDatabaseMetadata() (*DatabaseMetadata, error) {
	tableNames, err := ds.getTableNames()
	if err != nil {
		return nil, err
	}

	metadata := &DatabaseMetadata{Tables: []TableMetadata{}}

	for _, tableName := range tableNames {
		tableMetadata, err := ds.GetTableMetadata(tableName)
		if err != nil {
			return nil, err
		}
		metadata.Tables = append(metadata.Tables, *tableMetadata)
	}

	return metadata, nil
}

func (ds *DatabaseSource) getTableNames() ([]string, error) {
	rows, err := ds.DB.Query("SELECT table_name FROM information_schema.tables WHERE table_schema = 'public' AND table_type = 'BASE TABLE'")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tableNames []string
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			return nil, err
		}
		tableNames = append(tableNames, tableName)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return tableNames, nil
}

func (ds *DatabaseSource) getTableColumns(tableName string) ([]Column, error) {
	rows, err := ds.DB.Query("SELECT column_name, data_type, is_nullable FROM information_schema.columns WHERE table_name = $1", tableName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var columns []Column
	for rows.Next() {
		var col Column
		if err := rows.Scan(&col.Name, &col.DataType, &col.Nullable); err != nil {
			return nil, err
		}
		columns = append(columns, col)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return columns, nil
}

func (ds *DatabaseSource) getTablePrimaryKeys(tableName string) ([]string, error) {
	rows, err := ds.DB.Query(`
                SELECT kcu.column_name
                FROM information_schema.table_constraints AS tc
                JOIN information_schema.key_column_usage AS kcu
                        ON tc.constraint_name = kcu.constraint_name
                        AND tc.table_schema = kcu.table_schema
                WHERE tc.constraint_type = 'PRIMARY KEY' AND tc.table_name = $1;`, tableName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var primaryKeys []string
	for rows.Next() {
		var pk string
		if err := rows.Scan(&pk); err != nil {
			return nil, err
		}
		primaryKeys = append(primaryKeys, pk)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return primaryKeys, nil
}

func (ds *DatabaseSource) GetTableMetadata(tableName string) (*TableMetadata, error) {
	tableInfo := &TableMetadata{Name: tableName}

	columns, err := ds.getTableColumns(tableName)
	if err != nil {
		return nil, err
	}
	tableInfo.Columns = columns

	primaryKey, err := ds.getTablePrimaryKeys(tableName)
	if err != nil {
		return nil, err
	}
	tableInfo.PrimaryKey = primaryKey

	return tableInfo, nil
}

func (ds *DatabaseSource) GetData(tableName string) ([]map[string]any, error) {
	rows, err := ds.DB.Query(fmt.Sprintf("SELECT * FROM %s", tableName))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	var data []map[string]any
	for rows.Next() {
		columnPointers := make([]any, len(columns))
		for i := range columnPointers {
			columnPointers[i] = new(any)
		}

		if err := rows.Scan(columnPointers...); err != nil {
			return nil, err
		}

		rowData := make(map[string]any)
		for i, colName := range columns {
			val := columnPointers[i].(*any)
			switch v := (*val).(type) {
			case []byte:
				rowData[colName] = string(v)
			case string:
				rowData[colName] = v
			default:
				rowData[colName] = *val
			}
		}
		data = append(data, rowData)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return data, nil
}

func main() {
	connStr := "user=postgres password=postgres dbname=postgres sslmode=disable"
	source, err := NewDatabaseSource(connStr)
	if err != nil {
		log.Fatal(err)
	}
	defer source.Close()

	m, err := source.GetDatabaseMetadata()

	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("GET /tables", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(m.Tables)
	})

	http.HandleFunc("GET /tables/{name}/data", func(w http.ResponseWriter, r *http.Request) {
		tableName := r.PathValue("name")

		w.Header().Set("Content-Type", "application/json")

		data, err := source.GetData(tableName)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		json.NewEncoder(w).Encode(data)
	})

	http.ListenAndServe(":8080", nil)

}
