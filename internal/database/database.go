package database

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

type DataBase struct {
	DB *sql.DB
}

func InitDB(dsn string) (*DataBase, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to connect database: %w", err)
	}

	return &DataBase{DB: db}, nil
}
