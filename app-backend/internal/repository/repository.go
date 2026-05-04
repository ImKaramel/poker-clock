package repository

import (
	"context"
	"time"

	"github.com/pridecrm/app-backend/internal/domain"
)

type UserRepository interface {
	Create(ctx context.Context, u *domain.User) error
	GetByID(ctx context.Context, userID string) (*domain.User, error)
	GetByUsername(ctx context.Context, username string) (*domain.User, error)
	Update(ctx context.Context, u *domain.User) error
	Delete(ctx context.Context, userID string) error
	List(ctx context.Context) ([]domain.User, error)
	Count(ctx context.Context) (int64, error)
	CountBanned(ctx context.Context) (int64, error)
	ListRecent(ctx context.Context, limit int) ([]domain.User, error)
	ListForRating(ctx context.Context) ([]domain.User, error)
}

type GameRepository interface {
	Create(ctx context.Context, g *domain.Game) error
	GetByID(ctx context.Context, id int64) (*domain.Game, error)
	Update(ctx context.Context, g *domain.Game) error
	Delete(ctx context.Context, id int64) error
	ListAll(ctx context.Context) ([]domain.Game, error)
	ListActive(ctx context.Context) ([]domain.Game, error)
	ListRecent(ctx context.Context, limit int) ([]domain.Game, error)
	Count(ctx context.Context) (int64, error)
	CountActive(ctx context.Context) (int64, error)
}

type ParticipantRepository interface {
	Create(ctx context.Context, p *domain.Participant) error
	GetByID(ctx context.Context, id int64) (*domain.Participant, error)
	GetByUserAndGame(ctx context.Context, userID string, gameID int64) (*domain.Participant, error)
	Update(ctx context.Context, p *domain.Participant, rebuyDelta int) error
	Delete(ctx context.Context, id int64) error
	DeleteByUserAndGame(ctx context.Context, userID string, gameID int64) error
	ListByGame(ctx context.Context, gameID int64) ([]domain.Participant, error)
	ListByUser(ctx context.Context, userID string) ([]domain.Participant, error)
	ListAll(ctx context.Context) ([]domain.Participant, error)
	ListUpcomingForUser(ctx context.Context, userID string, fromDate time.Time) ([]domain.Participant, error)
	Count(ctx context.Context) (int64, error)
	CountByGame(ctx context.Context, gameID int64) (int64, error)
	SetArrived(ctx context.Context, participantID int64, arrived bool) error
}

type SupportTicketRepository interface {
	Create(ctx context.Context, t *domain.SupportTicket) error
	GetByID(ctx context.Context, id int64) (*domain.SupportTicket, error)
	Update(ctx context.Context, t *domain.SupportTicket) error
	Delete(ctx context.Context, id int64) error
	ListByUser(ctx context.Context, userID string) ([]domain.SupportTicket, error)
	ListAll(ctx context.Context) ([]domain.SupportTicket, error)
	CountOpen(ctx context.Context) (int64, error)
}

type TournamentRepository interface {
	CreateHistory(ctx context.Context, h *domain.TournamentHistory) error
	GetHistoryByID(ctx context.Context, id int64) (*domain.TournamentHistory, error)
	ListHistory(ctx context.Context) ([]domain.TournamentHistory, error)
	ListHistoryByUser(ctx context.Context, userID string) ([]domain.TournamentHistory, error)
	UpdateHistory(ctx context.Context, h *domain.TournamentHistory) error
	DeleteHistory(ctx context.Context, id int64) error
	AddTournamentParticipant(ctx context.Context, p *domain.TournamentParticipant) error
	ListTournamentParticipants(ctx context.Context, historyID int64) ([]domain.TournamentParticipant, error)
}
