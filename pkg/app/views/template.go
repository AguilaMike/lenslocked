package views

import (
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"net/http"
)

type Template struct {
	htmlTpl *template.Template
}

func (t Template) Execute(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	err := t.htmlTpl.Execute(w, data)
	if err != nil {
		log.Printf("executing template: %v", err)
		http.Error(w, "There was an error executing the template.", http.StatusInternalServerError)
		return
	}
}

func Parse(filepath string) (Template, error) {
	htmlTpl, err := template.ParseFiles(filepath)
	return parseInternal(htmlTpl, err, "ParseFiles")
}

func ParseFS(fs fs.FS, pattern ...string) (Template, error) {
	htmlTpl, err := template.ParseFS(fs, pattern...)
	return parseInternal(htmlTpl, err, "ParseFS")
}

func Must(t Template, err error) Template {
	if err != nil {
		panic(err)
	}
	return t
}

func parseInternal(htmlTpl *template.Template, err error, nameFunction string) (Template, error) {
	if err != nil {
		return Template{}, fmt.Errorf("parsing template (%s): %w", nameFunction, err)
	}

	return Template{htmlTpl: htmlTpl}, nil
}
