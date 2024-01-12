package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/gorilla/csrf"
	"github.com/joho/godotenv"
	"github.com/taherk/galleryapp/controllers"
	"github.com/taherk/galleryapp/migrations"
	"github.com/taherk/galleryapp/models"
	"github.com/taherk/galleryapp/templates"
	"github.com/taherk/galleryapp/views"
)

// it is better not to refer the config in the rather pass in the values needed.
// which is why the config structs of components are defined in their package.
type config struct {
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

func loadEnvConfig() (config, error) {
	var cfg config
	err := godotenv.Load()
	if err != nil {
		return cfg, nil
	}

	// psql
	cfg.PSQL = models.DefaultPostgresConfig()

	// csrf
	cfg.CSRF.Key = "hciCfay4reF2GIyx7Fi3CUoakVSRgap9"
	cfg.CSRF.Secure = false

	// http server
	cfg.Server.Address = ":3000"

	// smtp
	host := os.Getenv("SMTP_HOST")
	portStr := os.Getenv("SMTP_PORT")
	port, err := strconv.Atoi(portStr)
	if err != nil {
		panic(fmt.Errorf("invalid port: %w", err))
	}
	username := os.Getenv("SMTP_USERNAME")
	password := os.Getenv("SMTP_PASSWORD")

	cfg.SMTP.Host = host
	cfg.SMTP.Port = port
	cfg.SMTP.Username = username
	cfg.SMTP.Password = password

	return cfg, nil
}

func main() {
	config, err := loadEnvConfig()
	if err != nil {
		panic(err)
	}

	// Setup DB
	cfg := models.DefaultPostgresConfig()
	db, err := models.Open(cfg)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	err = models.MigrateFS(db, migrations.FS, ".")
	if err != nil {
		panic(err)
	}

	// Setup Services
	userService := &models.UserService{
		DB: db,
	}
	sessionService := &models.SessionService{
		DB: db,
	}
	pwResetService := &models.PasswordResetService{
		DB: db,
	}
	emailService, err := models.NewEmailService(config.SMTP)
	if err != nil {
		log.Fatalf("cannot create mail service: %v", err)
	}

	csrfMiddleware := func(next http.Handler) http.Handler {
		csrfMw := csrf.Protect([]byte(config.CSRF.Key), csrf.Secure(config.CSRF.Secure))
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, r)
		})

		return csrfMw(handler)
	}

	umw := controllers.UserMiddleware{
		SessionService: sessionService,
	}

	usersC := controllers.Users{
		UserService:          userService,
		SessionService:       sessionService,
		EmailService:         emailService,
		PasswordResetService: pwResetService,
	}
	usersC.Templates.New = views.Must(views.ParseFS(templates.FS, "sign-up.gohtml", "tailwind.gohtml"))
	usersC.Templates.SignIn = views.Must(views.ParseFS(templates.FS, "sign-in.gohtml", "tailwind.gohtml"))
	usersC.Templates.ForgotPassword = views.Must(views.ParseFS(templates.FS, "forgot-pw.html", "tailwind.gohtml"))

	r := chi.NewRouter()
	r.Use(csrfMiddleware)
	r.Use(umw.SetUser)
	r.Get("/", controllers.StaticHandler(views.Must(views.ParseFS(templates.FS, "home.gohtml", "tailwind.gohtml"))))
	r.Get("/contact", controllers.StaticHandler(views.Must(views.ParseFS(templates.FS, "contact.gohtml", "tailwind.gohtml"))))
	r.Get("/faq", controllers.FAQ(views.Must(views.ParseFS(templates.FS, "faq.gohtml", "tailwind.gohtml"))))

	// views
	r.Get("/signup", usersC.New)
	r.Get("/signin", usersC.SignIn)
	r.Get("/forgot-pw", usersC.ForgotPassword)
	r.Route("/user/me", func(r chi.Router) {
		r.Use(umw.RequireUser)
		r.Get("/", usersC.CurrentUser)
	})
	r.Get("/users/me", usersC.CurrentUser)

	// processing
	r.Post("/users", usersC.Create)
	r.Post("/signin", usersC.ProcessSignIn)
	r.Post("/signout", usersC.ProcessSignout)
	r.Post("/forgot-pw", usersC.ProcessForgotPassword)

	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Page not found", http.StatusNotFound)
	})

	fmt.Printf("Starting the server on %s...\n", config.Server.Address)
	err = http.ListenAndServe(config.Server.Address, r)
	if err != nil {
		panic(err)
	}
}
