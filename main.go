package main

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/taherk/galleryapp/controllers"
	"github.com/taherk/galleryapp/migrations"
	"github.com/taherk/galleryapp/models"
	"github.com/taherk/galleryapp/templates"
	"github.com/taherk/galleryapp/views"
)

func main() {
	r := chi.NewRouter()
	r.Get("/", controllers.StaticHandler(views.Must(views.ParseFS(templates.FS, "home.gohtml", "tailwind.gohtml"))))
	r.Get("/contact", controllers.StaticHandler(views.Must(views.ParseFS(templates.FS, "contact.gohtml", "tailwind.gohtml"))))
	r.Get("/faq", controllers.FAQ(views.Must(views.ParseFS(templates.FS, "faq.gohtml", "tailwind.gohtml"))))

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

	userService := models.UserService{
		DB: db,
	}
	usersC := controllers.Users{
		UserService: &userService,
	}
	usersC.Templates.New = views.Must(views.ParseFS(templates.FS, "sign-up.gohtml", "tailwind.gohtml"))
	usersC.Templates.SignIn = views.Must(views.ParseFS(templates.FS, "sign-in.gohtml", "tailwind.gohtml"))

	// views
	r.Get("/signup", usersC.New)
	r.Get("/signin", usersC.SignIn)

	// processing
	r.Post("/users", usersC.Create)
	r.Post("/signin", usersC.ProcessSignIn)

	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Page not found", http.StatusNotFound)
	})

	fmt.Println("Starting the server on :3000...")
	http.ListenAndServe(":3000", r)
}
