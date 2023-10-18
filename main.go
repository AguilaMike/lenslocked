package main

import (
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/AguilaMike/lenslocked/pkg/app/controllers"
	"github.com/AguilaMike/lenslocked/pkg/app/migrations"
	"github.com/AguilaMike/lenslocked/pkg/app/models"
	"github.com/AguilaMike/lenslocked/pkg/app/router"
	"github.com/go-chi/chi/v5"
	"github.com/gorilla/csrf"
)

func main() {
	// Setup a database connection
	cfg := models.DefaultPostgresConfig()
	db, err := models.Open(cfg)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	// Add the following code
	// err = models.Migrate(db, "pkg/app/migrations")
	err = models.MigrateFS(db, migrations.FS, ".")
	if err != nil {
		panic(err)
	}

	// Setup our model services
	userService := models.UserService{
		DB: db,
	}

	sessionService := models.SessionService{
		DB: db,
	}

	// Setup our router
	r := chi.NewRouter()
	csrfKey := "gFvi45R4fy5xNBlnEeZtQbfAVCYEIAUX"
	csrfMw := csrf.Protect(
		[]byte(csrfKey),
		// TODO: Fix this before deploying
		csrf.Secure(false),
	)
	// These middleware are used everywhere.
	r.Use(csrfMw)
	umw := controllers.UserMiddleware{
		SessionService: &sessionService,
	}
	r.Use(umw.SetUser)
	r.Use(LogMiddleware)

	router.Router(r, umw, userService, sessionService)
	fmt.Println("Starting the server on :8080...")
	fmt.Println(http.ListenAndServe(":8080", r))
}

func LogMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		ip, err := getIP(r)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		next.ServeHTTP(w, r)
		fmt.Printf("Request: IP [%s] Method [%s] Path [%s] Time[%s]\n", ip, r.Method, r.URL.Path, time.Since(start))
		log.Printf("Request: IP [%s] Method [%s] Path [%s] Time[%s]", ip, r.Method, r.URL.Path, time.Since(start))
	})
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
