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

	clock := chi.NewRouter()

	clock.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})

	clock.Post("/auth/login", authHandler.Login)

	clock.Route("/tournaments", func(r chi.Router) {
		r.Use(auth.AdminAuth)

		r.Post("/", tournamentHandler.CreateTournament)
		r.Get("/", tournamentHandler.ListTournaments)

		r.Route("/{id}", func(r chi.Router) {
			r.Get("/", tournamentHandler.GetTournament)
			r.Delete("/", tournamentHandler.DeleteTournament)
			r.Post("/levels", levelHandler.AddLevel)
			r.Get("/levels", levelHandler.ListLevels)
			r.Delete("/levels/{levelId}", levelHandler.DeleteLevel)

			r.Post("/start", timerHandler.StartTournament)
			r.Post("/pause", timerHandler.PauseTournament)
			r.Post("/resume", timerHandler.ResumeTournament)
			r.Post("/next", timerHandler.NextLevel)
			r.Post("/stats", timerHandler.UpdateStats)
		})
	})

	clock.Get("/tournaments/{id}/timer/ws", timerHandler.TimerWS)

	r.Mount("/clock", clock)

	return r
}
