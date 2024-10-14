package db

import (
    "database/sql"
    "log"
    _ "github.com/go-sql-driver/mysql" // Import MySQL driver
)

var database *sql.DB

// InitDB initializes the database connection and returns the *sql.DB instance
func InitDB() *sql.DB {
    var err error
    // Update with your actual connection string
    database, err = sql.Open("mysql", "root:password@tcp(127.0.0.1:3306)/bookstore")
    if err != nil {
        log.Fatal(err)
    }
    
    // Check if the connection is established
    if err = database.Ping(); err != nil {
        log.Fatal(err)
    }
    
    log.Println("Database connected successfully")
    return database // Return the *sql.DB instance
}
