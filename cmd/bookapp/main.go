package main

import (
    "database/sql"
    "html/template"
    "log"
    "net/http"

    _ "github.com/go-sql-driver/mysql" // Import the MySQL driver
    "bookstore/internal/db"
    "bookstore/pkg/book"
)

var database *sql.DB // Global variable to hold the database connection

func main() {
    var err error
    database = db.InitDB() // Initialize MySQL connection
    if database == nil {
        log.Fatalf("Could not connect to the database: %v", err)
    }

    // Serve static files from the frontend directory
    http.Handle("/css/", http.StripPrefix("/css/", http.FileServer(http.Dir("frontend/css"))))

    http.HandleFunc("/", homeHandler)               // Handle GET requests to the home page
    http.HandleFunc("/books", func(w http.ResponseWriter, r *http.Request) {
        book.HandleBooks(w, r, database) // Pass the database connection
    })     // Handle book-related requests
	http.HandleFunc("/books/", func(w http.ResponseWriter, r *http.Request) {
        book.HandleBooks(w, r, database) // Handle delete requests here too
    })

    log.Println("Starting server on :8080")
    log.Fatal(http.ListenAndServe(":8080", nil))
}

// homeHandler serves the index.html template and fetches books from the database
func homeHandler(w http.ResponseWriter, r *http.Request) {
    books, err := book.GetBooks(database) // Fetch books from the database
    if err != nil {
        http.Error(w, "Failed to get books", http.StatusInternalServerError)
        return
    }

    tmpl, err := template.ParseFiles("frontend/templates/index.html") // Adjust path as necessary
    if err != nil {
        http.Error(w, "Failed to load template", http.StatusInternalServerError)
        return
    }

    err = tmpl.Execute(w, books) // Pass the list of books to the template
    if err != nil {
        http.Error(w, "Failed to execute template", http.StatusInternalServerError)
    }
}
