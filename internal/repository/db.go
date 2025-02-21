package repository

import (
    "database/sql"
    _ "github.com/mattn/go-sqlite3"
)

func NewDB() (*sql.DB, error) {
    db, err := sql.Open("sqlite3", "./langportal.db")
    if err != nil {
        return nil, err
    }
    
    if err = db.Ping(); err != nil {
        return nil, err
    }
    
    return db, nil
}