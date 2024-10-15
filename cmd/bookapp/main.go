package main

import (
    "database/sql"
	"text/template"
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

    // Home and Library handlers
    http.HandleFunc("/", homeHandler) // Handle GET requests to the home page
    http.HandleFunc("/library", libraryHandler) // Handle GET requests to the library page
    http.HandleFunc("/books", func(w http.ResponseWriter, r *http.Request) {
        book.HandleBooks(w, r, database) // Delegate to the book handler
    })
	http.HandleFunc("/books/", func(w http.ResponseWriter, r *http.Request) {
        book.HandleBooks(w, r, database) // Delegate to the book handler
    })

    log.Println("Starting server on :8080")
    log.Fatal(http.ListenAndServe(":8080", nil))
}

// homeHandler serves the addBook.html template without fetching books
func homeHandler(w http.ResponseWriter, r *http.Request) {
    tmpl, err := template.ParseFiles("frontend/templates/addBook.html") // Adjust path as necessary
    if err != nil {
        http.Error(w, "Failed to load template", http.StatusInternalServerError)
        return
    }
    err = tmpl.Execute(w, nil) // No data is passed to the template
    if err != nil {
        http.Error(w, "Failed to execute template", http.StatusInternalServerError)
    }
}

// libraryHandler serves the library.html template and fetches books from the database
func libraryHandler(w http.ResponseWriter, r *http.Request) {
    books, err := book.GetBooks(database) // Fetch all books from the database
    if err != nil {
        http.Error(w, "Unable to retrieve books", http.StatusInternalServerError)
        return
    }

    tmpl := template.Must(template.ParseFiles("frontend/templates/library.html"))
    tmpl.Execute(w, books) // Pass the list of books to the template
}
