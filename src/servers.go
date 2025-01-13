package myserver

import (
	"database/sql"
	"net/http"
)

var db *sql.DB

func InitDatabase(database *sql.DB) {
	db = database
}

func HomePage(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("fname")
	insertQuery := "INSERT INTO users (username) VALUES (?)"
	_, err := db.Exec(insertQuery, username)
	if err != nil {
		http.Error(w, "Unable to insert data", http.StatusInternalServerError)
		return
	}
	http.ServeFile(w, r, "index.html")
}
