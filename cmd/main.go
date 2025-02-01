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
			username TEXT NOT NULL UNIQUE,
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
	_, err = db.Exec(`
    CREATE TABLE IF NOT EXISTS comments (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        post_id INTEGER NOT NULL,
        user_id INTEGER NOT NULL,
        content TEXT NOT NULL,
        created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
        FOREIGN KEY(post_id) REFERENCES posts(id),
        FOREIGN KEY(user_id) REFERENCES users(id)
    )`)
	if err != nil {
		log.Fatal("Failed to create comments table:", err)
	}

	_, err = db.Exec(`
    CREATE TABLE IF NOT EXISTS likes (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        post_id INTEGER NOT NULL,
        user_id INTEGER NOT NULL,
        created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
        UNIQUE(post_id, user_id),
        FOREIGN KEY(post_id) REFERENCES posts(id),
        FOREIGN KEY(user_id) REFERENCES users(id)
    )`)
	if err != nil {
		log.Fatal("Failed to create likes table:", err)
	}
	myserver.InitHandlers(db)
}

func main() {
	staticDir := filepath.Join("..", "static")
	fmt.Println(staticDir)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir(staticDir))))
	http.HandleFunc("/", myserver.SignUp)
	http.HandleFunc("/signin", myserver.SignIn)
	http.HandleFunc("/homepage", myserver.HomePage)
	http.HandleFunc("/comment", myserver.AddComment)
    http.HandleFunc("/like", myserver.AddLike)

	fmt.Println("Server running at http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
