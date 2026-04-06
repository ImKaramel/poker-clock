package api

import (
	"backend/internal/timer"
	"net/http"

	"github.com/go-chi/chi/v5"

	"backend/internal/api/handlers"
	"backend/internal/auth"
	"backend/internal/service"
)

func NewRouter(
	tournamentService *service.TournamentService,
	timerService *service.TimerService,
	timerManager timer.Manager,
) http.Handler {

	r := chi.NewRouter()

	authHandler := handlers.NewAuthHandler()
	tournamentHandler := handlers.NewTournamentHandler(tournamentService)
	levelHandler := handlers.NewLevelHandler(tournamentService)
	timerHandler := handlers.NewTimerHandler(timerService)

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})

	r.Post("/auth/login", authHandler.Login)

	r.Route("/tournaments", func(r chi.Router) {
		r.Use(auth.AdminAuth)

		r.Post("/", tournamentHandler.CreateTournament)
		r.Get("/", tournamentHandler.ListTournaments)

		r.Route("/{id}", func(r chi.Router) {
			r.Get("/", tournamentHandler.GetTournament)
			r.Post("/levels", levelHandler.AddLevel)
			r.Get("/levels", levelHandler.ListLevels)

			r.Post("/start", timerHandler.StartTournament)
			r.Post("/pause", timerHandler.PauseTournament)
			r.Post("/resume", timerHandler.ResumeTournament)
			r.Post("/next", timerHandler.NextLevel)
			r.Post("/tournaments/{id}/stats", timerHandler.UpdateStats)
		})
	})

	r.Get("/tournaments/{id}/timer/ws", timerHandler.TimerWS)

	return r
}
