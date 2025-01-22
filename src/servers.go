package myserver

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

// Initialize the database connection and create the table
func init() {
	var err error
	db, err = sql.Open("sqlite3", "./my.db")
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
}

func SingUp(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		http.ServeFile(w, r, "singup.html")
		return
	}

	r.ParseForm()
	username := r.FormValue("username")
	email := r.FormValue("email")
	password := r.FormValue("password")

	insertQuery := `INSERT INTO users (username, email, password) VALUES (?, ?, ?);`
	_, err := db.Exec(insertQuery, username, email, password)
	if err != nil {
		log.Println("Failed to insert user:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/singin", http.StatusSeeOther)
}

func SingIn(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		http.ServeFile(w, r, "singin.html")
		return
	}

	r.ParseForm()
	username := r.FormValue("username")
	password := r.FormValue("password")

	var dbPassword string
	err := db.QueryRow("SELECT password FROM users WHERE username = ?", username).Scan(&dbPassword)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Redirect(w, r, "/singin", http.StatusSeeOther)
			return
		}
		log.Println("Database error:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	fmt.Println(password + "Password does not match" + dbPassword)

	if password != dbPassword {
		http.Redirect(w, r, "/singin", http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, "/homepage", http.StatusSeeOther)
}

func HomePage(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "homepage.html")
}
