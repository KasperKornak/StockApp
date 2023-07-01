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
		http.Error(w, "Failed to parse form data", http.StatusBadRequest)
		return
	}

	// Print out the POST data to the terminal
	fmt.Println("Received POST data:")
	for key, values := range r.Form {
		for _, value := range values {
			fmt.Printf("%s: %s\n", key, value)
		}
	}

	// You can perform further processing with the POST data here

	// Send a response to the client
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Form submission successful"))
}
