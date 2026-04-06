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

	tournamentService := service.NewTournamentService(tournamentRepo)
	timerService := service.NewTimerService(tournamentRepo, timerRepo, timerManager)

	r := api.NewRouter(tournamentService, timerService, timerManager)

	srv := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}

	go func() {
		log.Println("server started on :8081")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("shutting down server...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	timerManager.Stop()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("server shutdown error: %v", err)
	}
}
