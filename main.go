package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

type DatabaseSource struct {
	DB *sql.DB
}

type DatabaseMetadata struct {
	Tables []TableMetadata
}

type TableMetadata struct {
	Name       string
	Columns    []Column
	PrimaryKey []string
}

type Column struct {
	Name     string
	DataType string
	Nullable string
}

func NewDatabaseSource(connStr string) (*DatabaseSource, error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	if err = db.Ping(); err != nil {
		return nil, err
	}

	fmt.Println("Successfully connected to PostgreSQL!")

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

	fmt.Println("Database Metadata:", m)
}
