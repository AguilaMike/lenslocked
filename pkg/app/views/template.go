package views

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"log"
	"net/http"
	"path"

	"github.com/AguilaMike/lenslocked/pkg/app/context"
	"github.com/AguilaMike/lenslocked/pkg/app/models"
	"github.com/gorilla/csrf"
)

type Template struct {
	htmlTpl *template.Template
}

func (t Template) Execute(w http.ResponseWriter, r *http.Request, data interface{}, errs ...error) {
	tpl, err := t.htmlTpl.Clone()
	if err != nil {
		log.Printf("cloning template: %v", err)
		http.Error(w, "There was an error rendering the page.", http.StatusInternalServerError)
		return
	}
	errMsgs, statusCode := errMessages(errs...)
	tpl = tpl.Funcs(
		template.FuncMap{
			"csrfField": func() template.HTML {
				return csrf.TemplateField(r)
			},
			"currentUser": func() *models.User {
				return context.User(r.Context())
			},
			"errors": func() []string {
				return errMsgs
			},
		},
	)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	var buf bytes.Buffer
	err = tpl.Execute(&buf, data)
	if err != nil {
		log.Printf("executing template: %v", err)
		http.Error(w, "There was an error executing the template.", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(statusCode)
	io.Copy(w, &buf)
}

func ParseFS(fs fs.FS, pattern ...string) (Template, error) {
	htmlTpl := template.New(path.Base(pattern[0]))
	htmlTpl = htmlTpl.Funcs(
		template.FuncMap{
			"currentUser": func() (*models.User, error) {
				return nil, fmt.Errorf("currentUser not implemented")
			},
			"csrfField": func() (template.HTML, error) {
				return "", fmt.Errorf("csrfField not implemented")
			},
			"errors": func() []string {
				return nil
			},
		},
	)
	htmlTpl, err := htmlTpl.ParseFS(fs, pattern...)
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
		log.Printf("parsing template (%s): %v", nameFunction, err)
		return Template{}, fmt.Errorf("parsing template (%s): %w", nameFunction, err)
	}

	return Template{htmlTpl: htmlTpl}, nil
}

// We will use this to determine if an error provides the Public method.
type public interface {
	Public() string
}

func errMessages(errs ...error) ([]string, int) {
	var msgs []string
	if errs == nil || len(errs) == 0 {
		return msgs, http.StatusOK
	}
	var statusCode int = http.StatusBadRequest
	for _, err := range errs {
		var pubErr public
		if errors.As(err, &pubErr) {
			msgs = append(msgs, pubErr.Public())
		} else {
			fmt.Println(err)
			msgs = append(msgs, "Something went wrong.")
			statusCode = http.StatusInternalServerError
		}
	}
	return msgs, statusCode
}
