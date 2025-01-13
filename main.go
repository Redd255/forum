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
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username 
	);`
	_, err = db.Exec(createTableQuery)
	if err != nil {
		log.Fatal("Failed to create table:", err)
	}
}

func main() {
	http.HandleFunc("/", myserver.HomePage)

	fmt.Println("Server running at http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
