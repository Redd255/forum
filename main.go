package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	myserver "test/src"

	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

func init() {
	var err error
	db, err = sql.Open("sqlite3", "./my.db")
	if err != nil {
		log.Fatal(err)
		return
	}

	myserver.InitDatabase(db)

	createTableQuery := `
	DROP TABLE IF EXISTS users;
	CREATE TABLE users (
		id INTEGER PRIMARY KEY AUTOINCREMENT, -- Unique ID for each user
		name TEXT NOT NULL,                   -- User's name
		email TEXT NOT NULL UNIQUE,           -- User's email (must be unique)
		age INTEGER,                          -- User's age
		pass TEXT NOT NULL
	);`

	_, err = db.Exec(createTableQuery)
	if err != nil {
		log.Fatal("Failed to recreate table:", err)
	}
}

func main() {
	http.HandleFunc("/", myserver.SingUp)
	http.HandleFunc("/HomePage", myserver.HomePage)
	http.HandleFunc("/SingIn", myserver.SingIn)
	http.HandleFunc("/SingIn2", myserver.SingIn2)


	fmt.Println("Server running at http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
