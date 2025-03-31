package db

import (
	"database/sql"
)

type DB struct {
	DB *sql.DB
}

func (db *DB) OpenConnection() {
	db.DB, _ = sql.Open("postgres", "postgres://app:gkGK7GJl5XSKiSlux57FJD45Fj@localhost:5432/trash_archive?sslmode=disable")
}

func (db *DB) CloseConnection() {
	db.DB = nil
}
