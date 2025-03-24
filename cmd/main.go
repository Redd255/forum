package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	myserver "test/src"

	_ "github.com/mattn/go-sqlite3"
)

func init() {
	os.MkdirAll("../uploads", os.ModePerm)
	db, err := sql.Open("sqlite3", "../database/my.db")
	if err != nil {
		log.Fatal(err)
	}

	sqlfile, err := os.ReadFile("../database/my.sql")
	if err != nil {
		log.Fatal("Failed to read SQL file:", err)
	}

	_, err = db.Exec(string(sqlfile))
	if err != nil {
		log.Fatal("Failed to create users table:", err)
	}

	defaultTags := []string{"Music", "Sports", "Technology", "Art", "Food", "Travel", "Fashion", "Health", "Education", "Gaming"}
	for _, tagName := range defaultTags {
		_, err = db.Exec("INSERT OR IGNORE INTO tags (name) VALUES (?)", tagName)
		if err != nil {
			log.Printf("Warning: Failed to insert default tag '%s': %v", tagName, err)
		}
	}

	myserver.InitHandlers(db)
}

func main() {
	staticDir := filepath.Join("..", "static")
	uploadsDir := filepath.Join("..", "uploads")

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir(staticDir))))
	http.Handle("/uploads/", http.StripPrefix("/uploads/", http.FileServer(http.Dir(uploadsDir))))

	http.HandleFunc("/", myserver.SignUp)
	http.HandleFunc("/signin", myserver.SignIn)
	http.HandleFunc("/logout", myserver.Logout)
	http.HandleFunc("/homepage", myserver.HomePage)
	http.HandleFunc("/comment", myserver.AddComment)
	http.HandleFunc("/like", myserver.AddLike)
	http.HandleFunc("/like-comment", myserver.LikeComment)
	http.HandleFunc("/tag", myserver.FilterByTag)

	fmt.Println("Server running at http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
