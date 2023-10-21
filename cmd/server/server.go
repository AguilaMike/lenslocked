package main

import (
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/AguilaMike/lenslocked/pkg/app/controllers"
	"github.com/AguilaMike/lenslocked/pkg/app/migrations"
	"github.com/AguilaMike/lenslocked/pkg/app/models"
	"github.com/AguilaMike/lenslocked/pkg/app/router"
	"github.com/go-chi/chi/v5"
	"github.com/gorilla/csrf"
	"github.com/joho/godotenv"
)

func loadEnvConfig() (router.Config, error) {
	var cfg router.Config
	err := godotenv.Load()
	if err != nil {
		return cfg, err
	}
	cfg.PSQL = models.DefaultPostgresConfig()
	if cfg.PSQL.Host == "" && cfg.PSQL.Port == "" {
		return cfg, fmt.Errorf("No PSQL Config provided.")
	}

	// SMTP
	cfg.SMTP.Host = os.Getenv("SMTP_HOST")
	portStr := os.Getenv("SMTP_PORT")
	cfg.SMTP.Port, err = strconv.Atoi(portStr)
	if err != nil {
		return cfg, err
	}
	cfg.SMTP.Username = os.Getenv("SMTP_USER")
	cfg.SMTP.Password = os.Getenv("SMTP_PASS")

	cfg.CSRF.Key = os.Getenv("CSRF_KEY")
	cfg.CSRF.Secure = os.Getenv("CSRF_SECURE") == "1"

	cfg.Server.Address = ":" + os.Getenv("PORT_GO")

	return cfg, nil
}

func main() {
	cfg, err := loadEnvConfig()
	if err != nil {
		panic(err)
	}

	// Setup a database connection
	db, err := models.Open(cfg.PSQL)
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
	sessionService := &models.SessionService{
		DB: db,
	}

	// Setup our router
	r := chi.NewRouter()
	csrfMw := csrf.Protect(
		[]byte(cfg.CSRF.Key),
		csrf.Secure(cfg.CSRF.Secure),
		csrf.Path("/"),
	)
	// These middleware are used everywhere.
	r.Use(csrfMw)
	umw := controllers.UserMiddleware{
		SessionService: sessionService,
	}
	r.Use(umw.SetUser)
	r.Use(LogMiddleware)

	router.Router(r, umw, cfg, db, sessionService)
	fmt.Printf("Starting the server on :%s...", cfg.Server.Address)
	err = http.ListenAndServe(cfg.Server.Address, r)
	if err != nil {
		panic(err)
	}
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
