package main

import (
	"fmt"
	"log"
	"net/http"

	myserver "test/src"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	http.HandleFunc("/", myserver.SignUp)
	http.HandleFunc("/signin", myserver.SignIn)
	http.HandleFunc("/homepage", myserver.HomePage)

	fmt.Println("Server running at http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
