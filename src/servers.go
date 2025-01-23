package myserver

import (
    "database/sql"
    "html/template"
    "log"
    "net/http"

    _ "github.com/mattn/go-sqlite3"
)

var db *sql.DB
var templates = template.Must(template.ParseFiles("signin.html", "signup.html"))

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

// SignUp handles user registration
func SignUp(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        templates.ExecuteTemplate(w, "signup.html", nil)
        return
    }
	
    if err := r.ParseForm(); err != nil {
        http.Error(w, "Failed to parse form", http.StatusBadRequest)
        return
    }

    username := r.FormValue("username")
    email := r.FormValue("email")
    password := r.FormValue("password")

    if username == "" || email == "" || password == "" {
        errorPage(w, "All fields are required")
        return
    }

    // Check if email exists
    var existingEmail string
    err := db.QueryRow("SELECT email FROM users WHERE email = ?", email).Scan(&existingEmail)
    if err == nil {
        errorPage(w, "Email already in use")
        return
    }
    if err != sql.ErrNoRows {
        log.Println("Database error:", err)
        http.Error(w, "Internal Server Error", http.StatusInternalServerError)
        return
    }

    _, err = db.Exec("INSERT INTO users (username, email, password) VALUES (?, ?, ?)",
        username, email, password)
    if err != nil {
        log.Println("Failed to insert user:", err)
        http.Error(w, "Internal Server Error", http.StatusInternalServerError)
        return
    }

    http.Redirect(w, r, "/signin", http.StatusSeeOther)
}

// SignIn handles user authentication
func SignIn(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        templates.ExecuteTemplate(w, "signin.html", nil)
        return
    }

    if err := r.ParseForm(); err != nil {
        http.Error(w, "Failed to parse form", http.StatusBadRequest)
        return
    }

    username := r.FormValue("username")
    password := r.FormValue("password")

    if username == "" || password == "" {
        errorPage(w, "Username and password are required")
        return
    }

    var dbPassword string
    err := db.QueryRow("SELECT password FROM users WHERE username = ?", username).Scan(&dbPassword)
    if err == sql.ErrNoRows {
        errorPage(w, "Invalid username or password")
        return
    }
    if err != nil {
        log.Printf("Database error: %v", err)
        http.Error(w, "Internal Server Error", http.StatusInternalServerError)
        return
    }

    if password != dbPassword {
        errorPage(w, "Invalid username or password")
        return
    }

    http.Redirect(w, r, "/homepage", http.StatusSeeOther)
}

func HomePage(w http.ResponseWriter, r *http.Request) {
    http.ServeFile(w, r, "homepage.html")
}

func errorPage(w http.ResponseWriter, message string) {
    data := map[string]string{
        "ErrorMessage": message,
    }
    err := templates.ExecuteTemplate(w, "signin.html", data)
    if err != nil {
        http.Error(w, "Failed to execute template: "+err.Error(), http.StatusInternalServerError)
    }
}