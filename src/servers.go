package myserver

import (
	"database/sql"
	"fmt"
	"net/http"
)

var db *sql.DB

func InitDatabase(database *sql.DB) {
	db = database
}

func SingUp(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "singup.html")
}

func SingIn(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	username := r.FormValue("fname")
	gmail := r.FormValue("gmail")
	age := r.FormValue("age")
	pass := r.FormValue("pass")
	insertQuery := `INSERT INTO users (name, email, age,pass) VALUES (?, ?, ?, ?);`
	_, err1 := db.Exec(insertQuery, username, gmail, age, pass)
	if err1 != nil {
		// http.Error(w, "Unable to insert data: "+err.Error(), http.StatusInternalServerError)
		return
	}
	// var test *sql.Row
	// err := db.QueryRow("SELECT pass FROM users WHERE username = ?", username).Scan(&test)
	// if err != nil {
	// 	fmt.Println("err")
	// }
	http.ServeFile(w, r, "singin.html")
}

func SingIn2(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	username := r.FormValue("fname")
	// pass := r.FormValue("pass")
	var test string
	err := db.QueryRow("SELECT pass FROM users WHERE name = ?", username).Scan(&test)
	if err != nil {
		fmt.Println("err")
	}
	fmt.Println(test)
	http.ServeFile(w, r, "singin2.html")
}

func HomePage(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "homepage.html")
}
