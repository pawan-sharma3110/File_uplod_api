package main

import (
	"database/sql"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"

	_ "github.com/lib/pq"
)

// Database connection function
func connectDB() (*sql.DB, error) {
	connStr := "host=localhost port=5432 user=postgres dbname=To_Do password=Pawan@2003 sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}
	return db, nil
}

// File upload handler
func uploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse the form data
	err := r.ParseMultipartForm(10 << 20) // 10 MB limit
	if err != nil {
		http.Error(w, "Error parsing form data", http.StatusBadRequest)
		return
	}

	// Get the file from the form
	file, handler, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Error retrieving the file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Read the file content into a byte slice
	fileBytes, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, "Error reading the file", http.StatusInternalServerError)
		return
	}

	// Connect to the database
	db, err := connectDB()
	if err != nil {
		http.Error(w, "Error connecting to the database", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	// Insert the file into the database
	_, err = db.Exec("INSERT INTO images (filename, data) VALUES ($1, $2)", handler.Filename, fileBytes)
	if err != nil {
		http.Error(w, "Error saving the file to the database", http.StatusInternalServerError)
		return
	}

	// Send a success response
	fmt.Fprintf(w, "File uploaded successfully!")
}

// Get image by ID handler
func getImageHandler(w http.ResponseWriter, r *http.Request) {
	// Parse the image ID from the URL path
	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	// Connect to the database
	db, err := connectDB()
	if err != nil {
		http.Error(w, "Error connecting to the database", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	// Query the database for the image data
	var filename string
	var data []byte
	err = db.QueryRow("SELECT filename, data FROM images WHERE id = $1", id).Scan(&filename, &data)
	if err == sql.ErrNoRows {
		http.Error(w, "Image not found", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, "Error fetching image from the database", http.StatusInternalServerError)
		return
	}

	// Set the content type and serve the image data
	w.Header().Set("Content-Type", "image/jpeg")
	w.Header().Set("Content-Disposition", fmt.Sprintf("inline; filename=\"%s\"", filename))
	w.Write(data)
}

func main() {
	// Set up the file upload route
	http.HandleFunc("/upload", uploadHandler)
	http.HandleFunc("/get-image/", getImageHandler)

	// Start the server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	fmt.Printf("Server is running on port %s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
