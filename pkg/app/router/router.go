package router

import (
	"database/sql"
	"fmt"
	"net/http"
	"path/filepath"

	"github.com/AguilaMike/lenslocked/pkg/app/controllers"
	"github.com/AguilaMike/lenslocked/pkg/app/models"
	"github.com/AguilaMike/lenslocked/pkg/app/templates"
	"github.com/AguilaMike/lenslocked/pkg/app/views"
	"github.com/go-chi/chi/v5"
)

type Config struct {
	PSQL models.PostgresConfig
	SMTP models.SMTPConfig
	CSRF struct {
		Key    string
		Secure bool
	}
	Server struct {
		Address string
	}
}

func Router(r *chi.Mux, umw controllers.UserMiddleware, cfg Config, db *sql.DB, sessionService *models.SessionService) {

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

	// Setup our model services
	userService := &models.UserService{
		DB: db,
	}
	pwResetService := &models.PasswordResetService{
		DB: db,
	}
	emailService := models.NewEmailService(cfg.SMTP)
	emailService.DefaultSender = cfg.SMTP.Username

	usersC := controllers.Users{
		UserService:          userService,
		SessionService:       sessionService,
		PasswordResetService: pwResetService,
		EmailService:         emailService,
	}
	SignUp(r, usersC)
	SignIn(r, usersC)
	ForgotPassword(r, usersC)
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

func ForgotPassword(r *chi.Mux, usersC controllers.Users) {
	usersC.Templates.ForgotPassword = views.Must(
		views.ParseFS(
			templates.FS,
			JoinPath("layout", "layout.gohtml"),
			JoinPath("pages", "auth", "forgot-pw.gohtml")))
	usersC.Templates.CheckYourEmail = views.Must(
		views.ParseFS(
			templates.FS,
			JoinPath("layout", "layout.gohtml"),
			JoinPath("pages", "auth", "check-your-email.gohtml")))
	r.Get("/forgot-pw", usersC.ForgotPassword)
	r.Post("/forgot-pw", usersC.ProcessForgotPassword)

	usersC.Templates.ResetPassword = views.Must(
		views.ParseFS(
			templates.FS,
			JoinPath("layout", "layout.gohtml"),
			JoinPath("pages", "auth", "reset-pw.gohtml")))
	r.Get("/reset-pw", usersC.ResetPassword)
	r.Post("/reset-pw", usersC.ProcessResetPassword)
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
