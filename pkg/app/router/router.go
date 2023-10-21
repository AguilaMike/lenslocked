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

	galleryService := &models.GalleryService{
		DB: db,
	}

	usersC.Templates.New = views.Must(
		views.ParseFS(
			templates.FS,
			JoinPath("layout", "layout.gohtml"),
			JoinPath("pages", "auth", "signup.gohtml")))
	usersC.Templates.SignIn = views.Must(
		views.ParseFS(
			templates.FS,
			JoinPath("layout", "layout.gohtml"),
			JoinPath("pages", "auth", "signin.gohtml")))
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
	usersC.Templates.ResetPassword = views.Must(
		views.ParseFS(
			templates.FS,
			JoinPath("layout", "layout.gohtml"),
			JoinPath("pages", "auth", "reset-pw.gohtml")))
	usersC.Templates.UserMe = views.Must(
		views.ParseFS(
			templates.FS,
			JoinPath("layout", "layout.gohtml"),
			JoinPath("pages", "auth", "userme.gohtml")))

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

	// signup
	r.Get("/signup", usersC.New)
	r.Post("/signup", usersC.Create)
	// signin
	r.Get("/signin", usersC.SignIn)
	r.Post("/signin", usersC.ProcessSignIn)
	// forgot-pw
	r.Get("/forgot-pw", usersC.ForgotPassword)
	r.Post("/forgot-pw", usersC.ProcessForgotPassword)
	// reset-pw
	r.Get("/reset-pw", usersC.ResetPassword)
	r.Post("/reset-pw", usersC.ProcessResetPassword)
	// users
	r.Route("/users", func(r chi.Router) {
		r.Use(umw.RequireUser)
		r.Get("/me", usersC.CurrentUser)
	})

	r.Post("/signout", usersC.ProcessSignOut)

	// Add this where the other controllers are created
	galleriesC := controllers.Galleries{
		GalleryService: galleryService,
	}

	galleriesC.Templates.Show = views.Must(views.ParseFS(
		templates.FS,
		JoinPath("layout", "layout.gohtml"),
		JoinPath("pages", "galleries", "show.gohtml"),
	))
	galleriesC.Templates.Index = views.Must(views.ParseFS(
		templates.FS,
		JoinPath("layout", "layout.gohtml"),
		JoinPath("pages", "galleries", "index.gohtml"),
	))
	galleriesC.Templates.New = views.Must(views.ParseFS(
		templates.FS,
		JoinPath("layout", "layout.gohtml"),
		JoinPath("pages", "galleries", "new.gohtml"),
	))
	galleriesC.Templates.Edit = views.Must(views.ParseFS(
		templates.FS,
		JoinPath("layout", "layout.gohtml"),
		JoinPath("pages", "galleries", "edit.gohtml"),
	))

	// galleries
	r.Route("/galleries", func(r chi.Router) {
		r.Get("/{id}", galleriesC.Show)
		r.Get("/{id}/images/{filename}", galleriesC.Image)
		r.Group(func(r chi.Router) {
			r.Use(umw.RequireUser)
			r.Get("/", galleriesC.Index)
			r.Get("/new", galleriesC.New)
			r.Post("/", galleriesC.Create)
			r.Get("/{id}/edit", galleriesC.Edit)
			r.Post("/{id}", galleriesC.Update)
			r.Post("/{id}/delete", galleriesC.Delete)
			// Images
			r.Post("/{id}/images", galleriesC.UploadImage)
			r.Post("/{id}/images/{filename}/delete", galleriesC.DeleteImage)
		})
	})

	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, fmt.Sprintf("404 Not Found: %s", r.URL.Path), http.StatusNotFound)
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
