package main

import (
	"fmt"
	"html/template"
	"net/http"

	"github.com/gorilla/mux"
)

var templates *template.Template

func main() {
	templates = template.Must(template.ParseGlob("../pkg/front/*.html"))
	r := mux.NewRouter()
	r.PathPrefix("/img/").Handler(http.StripPrefix("/img/", http.FileServer(http.Dir("../pkg/front/img/"))))

	r.HandleFunc("/home", homeGetHandler).Methods("GET")
	r.HandleFunc("/home", homePostHandler).Methods("POST")
	http.Handle("/", r)
	http.ListenAndServe(":8080", nil)
}

func homeGetHandler(w http.ResponseWriter, r *http.Request) {
	templates.ExecuteTemplate(w, "home.html", nil)
}

func homePostHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		// Handle the error
		http.Error(w, "Failed to parse form data", http.StatusBadRequest)
		return
	}

	login := r.Form.Get("login")
	email := r.Form.Get("email")
	password := r.Form.Get("password")

	// Print the form data values to the console
	fmt.Println("Login:", login)
	fmt.Println("Email:", email)
	fmt.Println("Password:", password)

	// Continue with your desired logic
	http.Redirect(w, r, "/home", 302)
}
