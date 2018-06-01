package main

import (
	"net/http"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/stripedpajamas/arkovmay/controllers"
	"github.com/stripedpajamas/arkovmay/database"
)

func main() {
	// connect to mysql and set up models
	database.InitDB()
	defer database.CloseDB()

	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Timeout(60 * time.Second))

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("welcome"))
	})

	r.Route("/users", func(r chi.Router) {
		r.Post("/", controllers.CreateUser)   // user registration
		r.Post("/login", controllers.Login)   // user login
		r.Post("/logout", controllers.Logout) // user logout
	})

	r.Route("/marks", func(r chi.Router) {
		r.Use(controllers.AuthMiddleware)
		r.Get("/", controllers.GetAllMarks) // get all marks for user
		r.Post("/", controllers.CreateMark) // mark creation for user
		r.Route("/{markID}", func(r chi.Router) {
			r.Use(controllers.MarkCtx)
			r.Get("/", controllers.GetMark)       // get specific mark for user
			r.Put("/", controllers.UpdateMark)    // update specific mark for user (process file(s))
			r.Delete("/", controllers.DeleteMark) // delete specific mark for user
		})
	})

	r.Get("/generate/{id}", controllers.Generate) // generate a sentence given a public id
	http.ListenAndServe(":3000", r)
}
