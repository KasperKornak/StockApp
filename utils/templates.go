package utils

import (
	"html/template"
	"net/http"
)

var templates *template.Template

// load templates
func LoadTemplates(pattern string) {
	templates = template.Must(template.ParseGlob(pattern))
}

// execute template and fill with data
func ExecuteTemplate(w http.ResponseWriter, tmpl string, data interface{}) {
	templates.ExecuteTemplate(w, tmpl, data)
}
