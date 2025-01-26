package myserver

import (
	"database/sql"
	"html/template"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	_ "github.com/mattn/go-sqlite3"
)

var (
	db        *sql.DB
	templates = template.Must(template.ParseFiles(
		"../templates/signin.html",
		"../templates/signup.html",
		"../templates/homepage.html",
	))
)

func InitHandlers(database *sql.DB) {
	db = database
}

// SignUp handles user registration
func SignUp(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		templates.ExecuteTemplate(w, "signup.html", nil)
		return
	}

	username := r.FormValue("username")
	email := r.FormValue("email")
	password := r.FormValue("password")

	if username == "" || email == "" || password == "" {
		errorPage(w, "All fields are required", "signup.html")
		return
	}

	// Check if email exists
	var existingEmail string
	err := db.QueryRow("SELECT email FROM users WHERE email = ?", email).Scan(&existingEmail)
	if err == nil {
		errorPage(w, "Email already in use", "signup.html")
		return
	}
	if err != sql.ErrNoRows {
		log.Println("Database error:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Println("Failed to hash password:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	_, err = db.Exec("INSERT INTO users (username, email, password) VALUES (?, ?, ?)",
		username, email, hashedPassword)
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
		errorPage(w, "Username and password are required", "signin.html")
		return
	}

	var userID int
	var hashedPassword string
	err := db.QueryRow("SELECT id, password FROM users WHERE username = ?", username).Scan(&userID, &hashedPassword)
	if err == sql.ErrNoRows {
		errorPage(w, "Invalid username or password", "signin.html")
		return
	}
	if err != nil {
		log.Printf("Database error: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password)); err != nil {
		errorPage(w, "Invalid username or password", "signin.html")
		return
	}

	// Create session
	sessionID := uuid.New().String()
	expiry := time.Now().Add(24 * time.Hour)
	_, err = db.Exec("INSERT INTO sessions (session_id, user_id, expiry) VALUES (?, ?, ?)",
		sessionID, userID, expiry)
	if err != nil {
		log.Printf("Failed to create session: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Set session cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    sessionID,
		Expires:  expiry,
		Path:     "/",
		HttpOnly: true,
	})

	http.Redirect(w, r, "/homepage", http.StatusSeeOther)
}

// HomePage handles post creation and display
func HomePage(w http.ResponseWriter, r *http.Request) {
	// Verify session
	cookie, err := r.Cookie("session")
	if err != nil {
		http.Redirect(w, r, "/signin", http.StatusSeeOther)
		return
	}

	var userID int
	var expiry time.Time
	err = db.QueryRow(`
		SELECT user_id, expiry 
		FROM sessions 
		WHERE session_id = ?`,
		cookie.Value).Scan(&userID, &expiry)
	if err != nil {
		http.Redirect(w, r, "/signin", http.StatusSeeOther)
		return
	}

	// Check session expiration
	if time.Now().After(expiry) {
		db.Exec("DELETE FROM sessions WHERE session_id = ?", cookie.Value)
		http.Redirect(w, r, "/signin", http.StatusSeeOther)
		return
	}

	// Handle post creation
	if r.Method == http.MethodPost {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Failed to parse form", http.StatusBadRequest)
			return
		}

		content := r.FormValue("content")
		if content == "" {
			errorPage(w, "Post content cannot be empty", "homepage.html")
			return
		}

		_, err := db.Exec("INSERT INTO posts (user_id, content) VALUES (?, ?)", userID, content)
		if err != nil {
			log.Println("Failed to create post:", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/homepage", http.StatusSeeOther)
		return
	}

	// Get current user's username
	var username string
	err = db.QueryRow("SELECT username FROM users WHERE id = ?", userID).Scan(&username)
	if err != nil {
		username = "Unknown"
	}

	// Get all posts
	type Post struct {
		Username string
		Content  string
	}
	var posts []Post

	rows, err := db.Query(`
		SELECT users.username, posts.content 
		FROM posts 
		JOIN users ON posts.user_id = users.id 
		ORDER BY posts.id DESC`)
	if err != nil {
		log.Println("Failed to fetch posts:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var post Post
		if err := rows.Scan(&post.Username, &post.Content); err != nil {
			log.Printf("Error scanning post: %v", err)
			continue
		}
		posts = append(posts, post)
	}

	// Render homepage template
	data := struct {
		Username string
		Posts    []Post
	}{
		Username: username,
		Posts:    posts,
	}

	templates.ExecuteTemplate(w, "homepage.html", data)
}
