// Package dbsqlite
package dbsqlite

import (
	"database/sql"
	"fmt"

	"github.com/adrg/xdg"
	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

func Init() error {
	dbFile, err := FileDB()
	if err != nil {
		return fmt.Errorf("could not get database file path: %w", err)
	}

	conn, err := sql.Open("sqlite3", dbFile)
	if err != nil {
		return fmt.Errorf("could not open database: %w", err)
	}
	db = conn

	return nil
}

func DB() *sql.DB {
	return db
}

func FileDB() (string, error) {
	configFilePath, err := xdg.ConfigFile("selene466-go-whatsapp-server/database.db")
	if err != nil {
		return "", fmt.Errorf("could not resolve path for database file: %w", err)
	}
	return configFilePath, nil
}

func Close() {
	if db != nil {
		db.Close()
	}
}
