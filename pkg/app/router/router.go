package router

import (
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/AguilaMike/lenslocked/pkg/app/controllers"
	"github.com/AguilaMike/lenslocked/pkg/app/models"
	"github.com/AguilaMike/lenslocked/pkg/app/templates"
	"github.com/AguilaMike/lenslocked/pkg/app/views"
	"github.com/go-chi/chi/v5"
)

func Router(r *chi.Mux, userService models.UserService, sessionService models.SessionService) {

	// Home
	registerGetControllerDefaultFs(r, "/", "layout.gohtml", "pages", "home.gohtml")
	// Contact
	registerGetControllerDefaultFs(r, "/contact", "layout.gohtml", "pages", "contact.gohtml")
	// FAQ
	r.Get("/faq", LogMiddleware(controllers.FAQ(
		views.Must(
			views.ParseFS(
				templates.FS,
				JoinPath("layout", "layout.gohtml"),
				JoinPath("pages", "faq.gohtml"),
			),
		),
	)))

	usersC := controllers.Users{
		UserService:    &userService,
		SessionService: &sessionService,
	}
	SignUp(r, usersC)
	SignIn(r, usersC)
	r.Get("/users/me", LogMiddleware(usersC.CurrentUser))
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
	r.Get("/signup", LogMiddleware(usersC.New))
	r.Post("/signup", LogMiddleware(usersC.Create))
}

func SignIn(r *chi.Mux, usersC controllers.Users) {
	usersC.Templates.New = views.Must(
		views.ParseFS(
			templates.FS,
			JoinPath("layout", "layout.gohtml"),
			JoinPath("pages", "auth", "signin.gohtml")))
	r.Get("/signin", LogMiddleware(usersC.New))
	r.Post("/signin", LogMiddleware(usersC.ProcessSignIn))
}

func registerGetControllerDefaultFs(r *chi.Mux, path, layout string, pages ...string) {
	registerGetControllerWithTemplateFs(r, nil, path, JoinPath("layout", layout), JoinPath(pages...))
}

func registerGetControllerWithTemplateFs(r *chi.Mux, data interface{}, path string, templateFile ...string) {
	tpl := views.Must(views.ParseFS(templates.FS, templateFile...))
	r.Get(path, LogMiddleware(controllers.StaticHandler(tpl, data)))
}

func registerGetControllerWithTemplate(r *chi.Mux, path, templateFolder, templateFile string, data interface{}) {
	tpl := views.Must(views.Parse(JoinPath(templateFolder, templateFile)))
	r.Get(path, LogMiddleware(controllers.StaticHandler(tpl, data)))
}

func JoinPath(path ...string) string {
	return filepath.Join(path...)
}

func LogMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		ip, err := getIP(r)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		next.ServeHTTP(w, r)
		fmt.Printf("Request: IP [%s] Method [%s] Path [%s] Time[%s]\n", ip, r.Method, r.URL.Path, time.Since(start))
		log.Printf("Request: IP [%s] Method [%s] Path [%s] Time[%s]", ip, r.Method, r.URL.Path, time.Since(start))
	}
}

// getIP returns the ip address from the http request
func getIP(r *http.Request) (string, error) {
	ips := r.Header.Get("X-Forwarded-For")
	splitIps := strings.Split(ips, ",")

	if len(splitIps) > 0 {
		// get last IP in list since ELB prepends other user defined IPs, meaning the last one is the actual client IP.
		netIP := net.ParseIP(splitIps[len(splitIps)-1])
		if netIP != nil {
			return netIP.String(), nil
		}
	}

	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return "", err
	}

	netIP := net.ParseIP(ip)
	if netIP != nil {
		ip := netIP.String()
		if ip == "::1" {
			return "127.0.0.1", nil
		}
		return ip, nil
	}

	return "", errors.New("IP not found")
}
