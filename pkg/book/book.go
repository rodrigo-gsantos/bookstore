package book

import (
    "database/sql"
    "encoding/json"
    "fmt"
    "net/http"
	"strconv"
	"log"
	"text/template"
)

// Book struct to represent a book entity
type Book struct {
    ID     int    `json:"id"`
    Title  string `json:"title"`
    Author string `json:"author"`
    Pages  int    `json:"pages"`
}

// AddBook inserts a new book into the database
func AddBook(db *sql.DB, book Book) error {
    query := "INSERT INTO books (title, author, pages) VALUES (?, ?, ?)"
    _, err := db.Exec(query, book.Title, book.Author, book.Pages)
    if err != nil {
        return fmt.Errorf("addBook: %v", err)
    }
    return nil
}

func GetBooks(db *sql.DB, search string, sortBy string) ([]Book, error) {
    var query string
    var rows *sql.Rows
    var err error

    // If a search query is provided, filter the results
    if search != "" {
        query = "SELECT id, title, author, pages FROM books WHERE title LIKE ? OR author LIKE ?"
        search = "%" + search + "%" // Prepare the search pattern
        rows, err = db.Query(query, search, search)
    } else {
        // Construct the base query
        query = "SELECT id, title, author, pages FROM books"

        // Add sorting based on the sortBy parameter
        switch sortBy {
        case "title":
            query += " ORDER BY title"
        case "author":
            query += " ORDER BY author"
        case "pages":
            query += " ORDER BY pages"
        default:
            query += " ORDER BY title" // Default sorting by title
        }

        rows, err = db.Query(query)
    }

    if err != nil {
        return nil, fmt.Errorf("getBooks: %v", err)
    }
    defer rows.Close()

    var books []Book
    for rows.Next() {
        var book Book
        if err := rows.Scan(&book.ID, &book.Title, &book.Author, &book.Pages); err != nil {
            return nil, fmt.Errorf("getBooks: %v", err)
        }
        books = append(books, book)
    }

    return books, nil
}


// UpdateBook updates an existing book in the database
func UpdateBook(db *sql.DB, book Book) error {
    query := "UPDATE books SET title = ?, author = ?, pages = ? WHERE id = ?"
    _, err := db.Exec(query, book.Title, book.Author, book.Pages, book.ID)
    if err != nil {
        return fmt.Errorf("updateBook: %v", err)
    }
    return nil
}


// DeleteBook removes a book from the database
func DeleteBook(db *sql.DB, id int) error {
	log.Printf("Attempting to delete book with ID: %d", id) // Log the ID being deleted
    query := "DELETE FROM books WHERE id = ?"
    _, err := db.Exec(query, id)
    if err != nil {
        return fmt.Errorf("deleteBook: %v", err)
    }
    return nil
}

func HandleBooks(w http.ResponseWriter, r *http.Request, db *sql.DB) {
    switch r.Method {
    case "GET":
		log.Println("Received GET request")
        sortBy := r.URL.Query().Get("sortBy")
        search := r.URL.Query().Get("search")

        log.Printf("SortBy: %s, Search: %s", sortBy, search)  // Log to verify the request parameters

        books, err := GetBooks(db, search, sortBy)
        if err != nil {
            // Log the error before returning it
            log.Printf("Error fetching books: %v", err)
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }

        if r.Header.Get("X-Requested-With") == "XMLHttpRequest" {
			// Handle AJAX request
			w.Header().Set("Content-Type", "application/json")
			err := json.NewEncoder(w).Encode(books)  // Return the books as JSON
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}
		

        // Otherwise, render the books as HTML for normal page load
        tmpl, err := template.ParseFiles("frontend/templates/library.html")
        if err != nil {
            log.Printf("Error parsing template: %v", err)
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }
        tmpl.Execute(w, books)	


    case "POST":
		log.Println("Received POST request")
        // Handle form submission
        title := r.FormValue("Title")
        author := r.FormValue("Author")
        pages, err := strconv.Atoi(r.FormValue("Pages"))
        if err != nil {
            http.Error(w, "Invalid number of pages", http.StatusBadRequest)
            return
        }

        book := Book{Title: title, Author: author, Pages: pages}
        err = AddBook(db, book) // Add book to the database
        if err != nil {
            http.Error(w, "Failed to add book", http.StatusInternalServerError)
            return
        }

        // Redirect back to the home page
        http.Redirect(w, r, "/", http.StatusSeeOther)

    case "PUT":
        log.Println("Received PUT request")
    var book Book
    
    // Check for decoding errors
    if err := json.NewDecoder(r.Body).Decode(&book); err != nil {
        log.Printf("Error decoding book: %v", err) // Log the error for debugging
        http.Error(w, "Bad request: "+err.Error(), http.StatusBadRequest)
        return
    }

    log.Printf("Updating Book with ID: %d, Title: %s, Author: %s, Pages: %d", book.ID, book.Title, book.Author, book.Pages)

    // Check if the book ID is valid
    if book.ID <= 0 {
        http.Error(w, "Invalid book ID", http.StatusBadRequest)
        return
    }

    if err := UpdateBook(db, book); err != nil {
        http.Error(w, "Failed to update book", http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusOK)

    case "DELETE":
		idStr := r.URL.Path[len("/books/"):] // Extract the ID from the URL
        id, err := strconv.Atoi(idStr)
        if err != nil {
            http.Error(w, "Invalid book ID", http.StatusBadRequest)
            return
        }

        log.Printf("Received DELETE request for book ID: %d", id) // Log the received ID

        // Call DeleteBook function
        err = DeleteBook(db, id)
        if err != nil {
            http.Error(w, "Failed to delete book", http.StatusInternalServerError)
            return
        }

        log.Printf("Successfully deleted book with ID: %d", id) // Log successful deletion
        w.WriteHeader(http.StatusNoContent) // Send a 204 No Content response
        return

    default:
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
    }
}