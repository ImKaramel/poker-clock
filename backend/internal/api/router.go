package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"backend/internal/api/handlers"
	"backend/internal/auth"
	"backend/internal/service"
)

func NewRouter() http.Handler {
	r := chi.NewRouter()

	authService := service.NewAuthService()
	tournamentService := service.NewTournamentService()

	authHandler := handlers.NewAuthHandler(authService)
	tournamentHandler := handlers.NewTournamentHandler(tournamentService)

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})

	r.Post("/auth/register", authHandler.Register)
	r.Post("/auth/login", authHandler.Login)

	r.Route("/tournaments", func(r chi.Router) {
		r.Use(auth.AuthMiddleware)

		r.Post("/", tournamentHandler.CreateTournament)
		r.Get("/", tournamentHandler.ListTournaments)
	})

	return r
}
