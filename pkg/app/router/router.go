package router

import (
	"fmt"
	"net/http"
	"path/filepath"

	"github.com/AguilaMike/lenslocked/pkg/app/controllers"
	"github.com/AguilaMike/lenslocked/pkg/app/models"
	"github.com/AguilaMike/lenslocked/pkg/app/templates"
	"github.com/AguilaMike/lenslocked/pkg/app/views"
	"github.com/go-chi/chi/v5"
)

func Router(r *chi.Mux, umw controllers.UserMiddleware, userService models.UserService, sessionService models.SessionService) {

	// Home
	registerGetControllerDefaultFs(r, "/", "layout.gohtml", "pages", "home.gohtml")
	// Contact
	registerGetControllerDefaultFs(r, "/contact", "layout.gohtml", "pages", "contact.gohtml")
	// FAQ
	r.Get("/faq", controllers.FAQ(
		views.Must(
			views.ParseFS(
				templates.FS,
				JoinPath("layout", "layout.gohtml"),
				JoinPath("pages", "faq.gohtml"),
			),
		),
	))

	usersC := controllers.Users{
		UserService:    &userService,
		SessionService: &sessionService,
	}
	SignUp(r, usersC)
	SignIn(r, usersC)
	UsersViews(r, umw, usersC)
	r.Post("/signout", usersC.ProcessSignOut)

	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, fmt.Sprintf("404 Not Found: %s", r.URL.Path), http.StatusNotFound)
	})
}

func SignUp(r *chi.Mux, usersC controllers.Users) {
	usersC.Templates.New = views.Must(
		views.ParseFS(
			templates.FS,
			JoinPath("layout", "layout.gohtml"),
			JoinPath("pages", "auth", "signup.gohtml")))
	r.Get("/signup", usersC.New)
	r.Post("/signup", usersC.Create)
}

func SignIn(r *chi.Mux, usersC controllers.Users) {
	usersC.Templates.New = views.Must(
		views.ParseFS(
			templates.FS,
			JoinPath("layout", "layout.gohtml"),
			JoinPath("pages", "auth", "signin.gohtml")))
	r.Get("/signin", usersC.New)
	r.Post("/signin", usersC.ProcessSignIn)
}

func UsersViews(r *chi.Mux, umw controllers.UserMiddleware, usersC controllers.Users) {
	usersC.Templates.UserMe = views.Must(
		views.ParseFS(
			templates.FS,
			JoinPath("layout", "layout.gohtml"),
			JoinPath("pages", "auth", "userme.gohtml")))
	r.Route("/users", func(r chi.Router) {
		r.Use(umw.RequireUser)
		r.Get("/me", usersC.CurrentUser)
	})
}

func registerGetControllerDefaultFs(r *chi.Mux, path, layout string, pages ...string) {
	registerGetControllerWithTemplateFs(r, nil, path, JoinPath("layout", layout), JoinPath(pages...))
}

func registerGetControllerWithTemplateFs(r *chi.Mux, data interface{}, path string, templateFile ...string) {
	tpl := views.Must(views.ParseFS(templates.FS, templateFile...))
	r.Get(path, controllers.StaticHandler(tpl, data))
}

func registerPostControllerWithTemplateFs(r *chi.Mux, data interface{}, path string, templateFile ...string) {
	tpl := views.Must(views.ParseFS(templates.FS, templateFile...))
	r.Post(path, controllers.StaticHandler(tpl, data))
}

func JoinPath(path ...string) string {
	return filepath.Join(path...)
}
