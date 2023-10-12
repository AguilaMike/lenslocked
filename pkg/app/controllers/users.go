package controllers

import (
	"fmt"
	"net/http"
)

type Users struct {
	Templates struct {
		New Template
	}
}

func (u Users) New(w http.ResponseWriter, r *http.Request) {
	var data struct {
		Email string
	}
	data.Email = r.FormValue("email")
	u.Templates.New.Execute(w, data)
}

func (u Users) Create(w http.ResponseWriter, r *http.Request) {
	type userForm struct {
		Email    string `forman:"email"`
		Password string `forman:"password"`
	}
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Unable to parse form submission.", http.StatusBadRequest)
		return
	}
	// // Convert the map to JSON
	// jsonData, _ := json.Marshal(r.Form)

	// // Convert the JSON to a struct
	// var user userForm
	// json.Unmarshal(jsonData, &user)

	fmt.Fprintln(w, r.Form)
	// fmt.Fprintln(w, user)
	fmt.Fprintf(w, "<p>Email: %s</p>", r.PostForm.Get("email"))
	fmt.Fprintf(w, "<p>Password: %s</p>", r.PostForm.Get("password"))
	fmt.Fprint(w, "Temporary response")
}
