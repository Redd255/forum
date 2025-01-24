package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	myserver "test/src"

	_ "github.com/mattn/go-sqlite3"
)


func init() {
	var err error
	db, err := sql.Open("sqlite3", "../database/my.db")
	if err != nil {
		log.Fatal(err)
	}

	createTableQuery := `
    CREATE TABLE IF NOT EXISTS users (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        username TEXT NOT NULL,
        email TEXT NOT NULL UNIQUE,
        password TEXT NOT NULL
    );`
	_, err = db.Exec(createTableQuery)
	if err != nil {
		log.Fatal("Failed to create table:", err)
	}
	createPostsTableQuery := `
    CREATE TABLE IF NOT EXISTS posts (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        user_id INTEGER NOT NULL,
        content TEXT NOT NULL,
        FOREIGN KEY(user_id) REFERENCES users(id)
    );`
	_, err = db.Exec(createPostsTableQuery)
	if err != nil {
		log.Fatal("Failed to create posts table:", err)
	}
	myserver.InitHandlers(db)
}

func main() {
	http.HandleFunc("/", myserver.SignUp)
	http.HandleFunc("/signin", myserver.SignIn)
	http.HandleFunc("/homepage", myserver.HomePage)

	fmt.Println("Server running at http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
