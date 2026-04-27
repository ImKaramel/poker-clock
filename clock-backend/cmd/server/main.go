package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"backend/internal/api"
	"backend/internal/service"
	"backend/internal/storage/postgres"
	"backend/internal/timer"
)

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")

		allowedOrigins := map[string]bool{
			"http://localhost:3000":                   true,
			"https://admin-panel-midnight.vercel.app": true,
			"https://midnight-club-app.ru":            true,
			"https://www.midnight-club-app.ru":        true,
		}

		if allowedOrigins[origin] {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		}

		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization, Accept")
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func main() {
	connString := os.Getenv("DATABASE_URL")
	if connString == "" {
		log.Fatal("DATABASE_URL is required")
	}

	db, err := postgres.NewDB(connString)
	if err != nil {
		log.Fatalf("database: %v", err)
	}
	defer db.Close()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if err := postgres.RunMigrations(ctx, db); err != nil {
		log.Fatalf("migrations: %v", err)
	}

	tournamentRepo := postgres.NewTournamentRepository(db)
	timerRepo := postgres.NewTimerRepository(db)

	timerManager := timer.NewManager(ctx, tournamentRepo, timerRepo)

	tournamentService := service.NewTournamentService(tournamentRepo, timerManager)
	timerService := service.NewTimerService(tournamentRepo, timerRepo, timerManager)

	router := api.NewRouter(tournamentService, timerService, timerManager)

	handler := corsMiddleware(router)

	srv := &http.Server{
		Addr:    ":8081",
		Handler: handler,
	}

	go func() {
		log.Println("clock-backend started on :8081")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("shutting down clock-backend...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	timerManager.Stop()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("shutdown error: %v", err)
	}
}
