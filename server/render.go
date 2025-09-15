package server

import (
	"html/template"
	"net/http"
	"path/filepath"
)

// renderTemplate finds the specified template, combines it with the layout,
// and writes the result to the http.ResponseWriter.
func renderTemplate(w http.ResponseWriter, tmplName string, data interface{}) {
	paths := []string{
		"templates/layout.html",
		filepath.Join("templates", tmplName),
	}

	tmpl, err := template.ParseFiles(paths...)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = tmpl.Execute(w, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}