package myserver

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
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

type Post struct {
	ID       int
	Username string
	Content  string
	Likes    int
	Comments []Comment
}

type Comment struct {
	Username string
	Content  string
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

	var existingUserName string
	err = db.QueryRow("SELECT username FROM users WHERE username = ?", username).Scan(&existingUserName)
	if err == nil {
		errorPage(w, "UserName already in use", "signup.html")
		return
	}
	if err != sql.ErrNoRows {
		log.Println("Database error:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

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

	if err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password)); err != nil {
		errorPage(w, "Invalid username or password", "signin.html")
		return
	}

	sessionID := uuid.New().String()
	expiry := time.Now().Add(24 * time.Hour)
	_, err = db.Exec("INSERT INTO sessions (session_id, user_id, expiry) VALUES (?, ?, ?)",
		sessionID, userID, expiry)
	if err != nil {
		log.Printf("Failed to create session: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

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
	cookie, err := r.Cookie("session")
	if err != nil {
		http.Redirect(w, r, "/signin", http.StatusSeeOther)
		return
	}

	var userID int
	var expiry time.Time
	err = db.QueryRow(`SELECT user_id, expiry FROM sessions WHERE session_id = ?`, cookie.Value).Scan(&userID, &expiry)
	if err != nil {
		http.Redirect(w, r, "/signin", http.StatusSeeOther)
		return
	}

	if time.Now().After(expiry) {
		db.Exec("DELETE FROM sessions WHERE session_id = ?", cookie.Value)
		http.Redirect(w, r, "/signin", http.StatusSeeOther)
		return
	}

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

	var username string
	err = db.QueryRow("SELECT username FROM users WHERE id = ?", userID).Scan(&username)
	if err != nil {
		username = "Unknown"
	}

	rows, err := db.Query(`
		SELECT posts.id, users.username, posts.content, 
		(SELECT COUNT(*) FROM likes WHERE likes.post_id = posts.id) AS likes
		FROM posts 
		JOIN users ON posts.user_id = users.id 
		ORDER BY posts.id DESC`)
	if err != nil {
		log.Println("Failed to fetch posts:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var posts []Post
	for rows.Next() {
		var post Post
		if err := rows.Scan(&post.ID, &post.Username, &post.Content, &post.Likes); err != nil {
			log.Printf("Error scanning post: %v", err)
			continue
		}

		commentRows, err := db.Query(`
			SELECT users.username, comments.content 
			FROM comments 
			JOIN users ON comments.user_id = users.id 
			WHERE comments.post_id = ? 
			ORDER BY comments.created_at ASC`, post.ID)
		if err != nil {
			log.Printf("Failed to get comments: %v", err)
			continue
		}
		defer commentRows.Close()

		for commentRows.Next() {
			var comment Comment
			if err := commentRows.Scan(&comment.Username, &comment.Content); err != nil {
				log.Printf("Error scanning comment: %v", err)
				continue
			}
			post.Comments = append(post.Comments, comment)
		}
		posts = append(posts, post)
	}

	data := struct {
		Username string
		Posts    []Post
	}{
		Username: username,
		Posts:    posts,
	}

	templates.ExecuteTemplate(w, "homepage.html", data)
}

func AddComment(w http.ResponseWriter, r *http.Request) {
	handlePostAction(w, r, func(userID int, postID string, content string) error {
		_, err := db.Exec("INSERT INTO comments (post_id, user_id, content) VALUES (?, ?, ?)",
			postID, userID, content)
		return err
	})
}

func AddLike(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session")
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var userID int
	err = db.QueryRow("SELECT user_id FROM sessions WHERE session_id = ?", cookie.Value).Scan(&userID)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if r.Method == http.MethodPost {
		r.ParseForm()
		postID := r.FormValue("post_id")

		var exists bool
		err := db.QueryRow("SELECT EXISTS (SELECT 1 FROM likes WHERE post_id = ? AND user_id = ?)", postID, userID).Scan(&exists)
		if err != nil {
			log.Printf("Failed to check like existence: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		if exists {
			_, err = db.Exec("DELETE FROM likes WHERE post_id = ? AND user_id = ?", postID, userID)
		} else {
			_, err = db.Exec("INSERT INTO likes (post_id, user_id) VALUES (?, ?)", postID, userID)
		}

		if err != nil {
			log.Printf("Failed to toggle like: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		var likeCount int
		err = db.QueryRow("SELECT COUNT(*) FROM likes WHERE post_id = ?", postID).Scan(&likeCount)
		if err != nil {
			log.Printf("Failed to get like count: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(fmt.Sprintf("%d", likeCount)))
	}
}

func handlePostAction(w http.ResponseWriter, r *http.Request,
	action func(int, string, string) error) {

	cookie, err := r.Cookie("session")
	if err != nil {
		http.Redirect(w, r, "/signin", http.StatusSeeOther)
		return
	}

	var userID int
	err = db.QueryRow("SELECT user_id FROM sessions WHERE session_id = ?",
		cookie.Value).Scan(&userID)
	if err != nil {
		http.Redirect(w, r, "/signin", http.StatusSeeOther)
		return
	}

	if r.Method == http.MethodPost {
		r.ParseForm()
		postID := r.FormValue("post_id")
		content := r.FormValue("content")

		if err := action(userID, postID, content); err != nil {
			log.Printf("Action failed: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	}

	http.Redirect(w, r, "/homepage", http.StatusSeeOther)
}
