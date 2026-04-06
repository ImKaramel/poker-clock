package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	httpdelivery "github.com/pridecrm/app-backend/internal/api/http"
	"github.com/pridecrm/app-backend/internal/infrastructure/auth"
	"github.com/pridecrm/app-backend/internal/infrastructure/db"
	"github.com/pridecrm/app-backend/internal/repository/postgres"
	"github.com/pridecrm/app-backend/internal/services"
	"github.com/pridecrm/app-backend/internal/usecase"
	"github.com/pridecrm/app-backend/pkg/config"
)

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	cfg := config.Load()
	dsn := cfg.DatabaseURL
	if dsn == "" {
		dsn = "postgres://postgres:postgres@localhost:5432/pridecrm?sslmode=disable"
		log.Info("DATABASE_URL not set, using default local DSN")
	}
	secret := cfg.JWTSecret
	if secret == "" {
		secret = "dev-insecure-change-me"
		log.Warn("JWT_SECRET not set, using insecure default")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	pool, err := db.Connect(ctx, dsn)
	if err != nil {
		log.Error("db connect", "err", err)
		os.Exit(1)
	}
	defer pool.Close()

	ctxMigrate, cancelM := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancelM()
	if err := db.Migrate(ctxMigrate, pool); err != nil {
		log.Error("db migrate", "err", err)
		os.Exit(1)
	}

	jwtSvc := auth.NewJWTService(secret, cfg.JWTTTL)
	clock := &services.Clock{}

	urepo := postgres.NewUserRepo(pool)
	grepo := postgres.NewGameRepo(pool)
	prepo := postgres.NewParticipantRepo(pool)
	srepo := postgres.NewSupportRepo(pool)
	trepo := postgres.NewTournamentRepo(pool)

	uc := &usecase.Service{
		Users:        urepo,
		Games:        grepo,
		Participants: prepo,
		Tickets:      srepo,
		Tournaments:  trepo,
		JWT:          jwtSvc,
		Log:          log,
		Clock:        clock,
	}

	h := &httpdelivery.Handlers{Log: log, UC: uc}
	h.Repo.Users = urepo
	h.Repo.Games = grepo
	h.Repo.Participants = prepo
	h.Repo.Tickets = srepo
	h.Repo.Tournaments = trepo

	engine := gin.New()
	engine.Use(gin.Recovery())
	engine.Use(httpdelivery.RequestLogger(log))
	httpdelivery.Mount(engine, h, jwtSvc, log)

	srvCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		log.Info("listening", "addr", cfg.Addr)
		if err := engine.Run(cfg.Addr); err != nil {
			log.Error("server stopped", "err", err)
			stop()
		}
	}()

	<-srvCtx.Done()
	log.Info("shutdown")
}
