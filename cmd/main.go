package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"path/filepath"

	myserver "test/src"

	_ "github.com/mattn/go-sqlite3"
)

func init() {
	db, err := sql.Open("sqlite3", "../database/my.db")
	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			username TEXT NOT NULL,
			email TEXT NOT NULL UNIQUE,
			password TEXT NOT NULL
		)`)
	if err != nil {
		log.Fatal("Failed to create users table:", err)
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS posts (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			content TEXT NOT NULL,
			FOREIGN KEY(user_id) REFERENCES users(id)
		)`)
	if err != nil {
		log.Fatal("Failed to create posts table:", err)
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS sessions (
			session_id TEXT PRIMARY KEY,
			user_id INTEGER NOT NULL,
			expiry DATETIME NOT NULL,
			FOREIGN KEY(user_id) REFERENCES users(id)
		)`)
	if err != nil {
		log.Fatal("Failed to create sessions table:", err)
	}

	myserver.InitHandlers(db)
}

func main() {
	staticDir := filepath.Join("..", "static")
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir(staticDir))))
	http.HandleFunc("/", myserver.SignUp)
	http.HandleFunc("/signup", myserver.SignUp)
	http.HandleFunc("/signin", myserver.SignIn)
	http.HandleFunc("/homepage", myserver.HomePage)

	fmt.Println("Server running at http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
