package main

import (
	"database/sql"
	"fmt"
	"github.com/lib/pq"
	"log"
)

func main() {
	connStr := "user=postgres password=postgres dbname=postgres sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if err = db.Ping(); err != nil {
		log.Fatal(err)
	}

	fmt.Println("Successfully connected to PostgreSQL!")

	rows, err := db.Query("SELECT table_name FROM information_schema.tables WHERE table_schema = 'public' AND table_type = 'BASE TABLE'")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	var tableNames []string
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			log.Fatal(err)
		}
		tableNames = append(tableNames, tableName)
	}

	if err := rows.Err(); err != nil {
		log.Fatal(err)
	}

	fmt.Println("Tables:", tableNames)

	for _, tableName := range tableNames {
		rows, err = db.Query("SELECT column_name, data_type, is_nullable FROM information_schema.columns WHERE table_name = $1", tableName)
		if err != nil {
			log.Fatal(err)
		}
		defer rows.Close()

		type Column struct {
			Name       string
			DataType   string
			IsNullable string
		}

		var columns []Column
		for rows.Next() {
			var col Column
			if err := rows.Scan(&col.Name, &col.DataType, &col.IsNullable); err != nil {
				log.Fatal(err)
			}
			columns = append(columns, col)
		}

		if err := rows.Err(); err != nil {
			log.Fatal(err)
		}

		fmt.Println("Columns:", columns)
	}

	for _, tableName := range tableNames {
		rows, err = db.Query(`
    	SELECT kcu.column_name
    	FROM information_schema.table_constraints AS tc
    	JOIN information_schema.key_column_usage AS kcu
      	ON tc.constraint_name = kcu.constraint_name
      	AND tc.table_schema = kcu.table_schema
    	WHERE tc.constraint_type = 'PRIMARY KEY' AND tc.table_name = $1;`, tableName)

		if err != nil {
			log.Fatal(err)
		}
		defer rows.Close()

		var primaryKeys []string
		for rows.Next() {
			var pk string
			if err := rows.Scan(&pk); err != nil {
				log.Fatal(err)
			}
			primaryKeys = append(primaryKeys, pk)
		}

		if err := rows.Err(); err != nil {
			log.Fatal(err)
		}

		fmt.Println("Primary Keys:", primaryKeys)

	}

	for _, tableName := range tableNames {
		rows, err = db.Query("SELECT * FROM " + pq.QuoteIdentifier(tableName))
		if err != nil {
			log.Fatal(err)
		}
		defer rows.Close()

		columnTypes, err := rows.ColumnTypes()
		if err != nil {
			log.Fatal(err)
		}

		columnCount := len(columnTypes)
		values := make([]interface{}, columnCount)
		valuePtrs := make([]interface{}, columnCount)

		for rows.Next() {
			for i := 0; i < columnCount; i++ {
				valuePtrs[i] = &values[i]
			}
			if err := rows.Scan(valuePtrs...); err != nil {
				log.Fatal(err)
			}

			row := make(map[string]interface{})
			for i, col := range columnTypes {
				val := values[i]
				if val == nil {
					row[col.Name()] = nil
				} else {
					row[col.Name()] = val
				}
			}
			fmt.Println(row)
		}

		if err := rows.Err(); err != nil {
			log.Fatal(err)
		}

	}

}
