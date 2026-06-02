package usecase

import (
	"context"
	"errors"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"
	"unicode"

	"github.com/pridecrm/app-backend/internal/domain"
)

var ErrCompletionResultsRequired = errors.New("completion results are required")
var ErrCompletionUnresolved = errors.New("completion results have unresolved users")
var ErrCompletionAlreadyDone = errors.New("game is already completed")

type CompleteResultInput struct {
	Position int    `json:"position"`
	Nickname string `json:"nickname"`
	KOCount  int    `json:"ko_count"`
	UserID   string `json:"user_id"`
}

type CompletePreview struct {
	Results    []CompletePreviewResult `json:"results"`
	Unresolved []CompletePreviewResult `json:"unresolved"`
}

type CompletePreviewResult struct {
	Position    int           `json:"position"`
	Nickname    string        `json:"nickname"`
	KOCount     int           `json:"ko_count"`
	BasePoints  int           `json:"base_points"`
	TotalPoints int           `json:"total_points"`
	Status      string        `json:"status"`
	User        *domain.User  `json:"user"`
	Candidates  []domain.User `json:"candidates"`
}

func reentryPrice(g *domain.Game) float64 {
	if g.ReentryBuyin > 0 {
		return g.ReentryBuyin
	}
	return g.Buyin
}

func normalizeMatchName(value string) string {
	value = strings.TrimSpace(strings.TrimPrefix(value, "@"))
	value = strings.ToLower(strings.ReplaceAll(value, "ё", "е"))
	var out strings.Builder
	for _, r := range value {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			out.WriteRune(r)
		}
	}
	return out.String()
}

func userMatchKeys(u domain.User) []string {
	keys := []string{
		normalizeMatchName(u.Username),
		normalizeMatchName(derefStr(u.NickName)),
		normalizeMatchName(derefStr(u.FirstName)),
		normalizeMatchName(derefStr(u.LastName)),
		normalizeMatchName(strings.TrimSpace(derefStr(u.FirstName) + " " + derefStr(u.LastName))),
	}
	seen := make(map[string]bool)
	out := make([]string, 0, len(keys))
	for _, key := range keys {
		if key == "" || seen[key] {
			continue
		}
		seen[key] = true
		out = append(out, key)
	}
	return out
}

func matchUserByNickname(users []domain.User, nickname string) []domain.User {
	target := normalizeMatchName(nickname)
	if target == "" {
		return nil
	}
	matches := make([]domain.User, 0, 1)
	for _, u := range users {
		for _, key := range userMatchKeys(u) {
			if key == target {
				matches = append(matches, u)
				break
			}
		}
	}
	return matches
}

func ratingPaidPlaces(totalPlayers int) int {
	if totalPlayers <= 0 {
		return 0
	}
	paid := int(math.Ceil(float64(totalPlayers) * 0.30))
	if paid < 1 {
		return 1
	}
	return paid
}

func ratingBasePoints(g *domain.Game, totalPlayers int) int {
	base := g.BasePoints
	if base <= 0 {
		base = 100
	}
	minPlayers := g.MinPlayersForExtraPoints
	if minPlayers <= 0 {
		minPlayers = 10
	}
	if g.PointsPerExtraPlayer > 0 && totalPlayers > minPlayers {
		base += (totalPlayers - minPlayers) * g.PointsPerExtraPlayer
	}
	return base
}

func ratingPlacePoints(g *domain.Game, totalPlayers int, position int) int {
	paid := ratingPaidPlaces(totalPlayers)
	if position <= 0 || position > paid {
		return 0
	}
	return ratingBasePoints(g, totalPlayers) * (paid - position + 1)
}

func sortedResults(results []CompleteResultInput) []CompleteResultInput {
	out := make([]CompleteResultInput, 0, len(results))
	for i, result := range results {
		result.Nickname = strings.TrimSpace(result.Nickname)
		result.UserID = strings.TrimSpace(result.UserID)
		if result.Position <= 0 {
			result.Position = i + 1
		}
		if result.KOCount < 0 {
			result.KOCount = 0
		}
		out = append(out, result)
	}
	sort.SliceStable(out, func(i, j int) bool {
		return out[i].Position < out[j].Position
	})
	return out
}

func (s *Service) CompleteGamePreview(ctx context.Context, gameID int64, results []CompleteResultInput) (*CompletePreview, error) {
	g, err := s.Games.GetByID(ctx, gameID)
	if err != nil {
		return nil, err
	}
	if g == nil {
		return nil, ErrNotFound
	}
	return s.buildCompletePreview(ctx, g, results)
}

func (s *Service) buildCompletePreview(ctx context.Context, g *domain.Game, results []CompleteResultInput) (*CompletePreview, error) {
	results = sortedResults(results)
	if len(results) == 0 {
		return nil, ErrCompletionResultsRequired
	}

	users, err := s.Users.List(ctx)
	if err != nil {
		return nil, err
	}
	usersByID := make(map[string]domain.User, len(users))
	for _, u := range users {
		usersByID[u.UserID] = u
	}

	preview := &CompletePreview{
		Results: make([]CompletePreviewResult, 0, len(results)),
	}
	resolvedCounts := make(map[string]int)

	for _, result := range results {
		row := CompletePreviewResult{
			Position:   result.Position,
			Nickname:   result.Nickname,
			KOCount:    result.KOCount,
			BasePoints: ratingPlacePoints(g, len(results), result.Position),
			Status:     "unresolved",
		}
		row.TotalPoints = row.BasePoints + result.KOCount*100

		if result.UserID != "" {
			if u, ok := usersByID[result.UserID]; ok {
				user := u
				row.User = &user
				row.Status = "resolved"
				resolvedCounts[user.UserID]++
			}
		} else {
			candidates := matchUserByNickname(users, result.Nickname)
			row.Candidates = candidates
			switch len(candidates) {
			case 0:
				row.Status = "unresolved"
			case 1:
				user := candidates[0]
				row.User = &user
				row.Status = "resolved"
				resolvedCounts[user.UserID]++
			default:
				row.Status = "ambiguous"
			}
		}

		preview.Results = append(preview.Results, row)
	}

	for i := range preview.Results {
		row := &preview.Results[i]
		if row.User != nil && resolvedCounts[row.User.UserID] > 1 {
			row.Status = "duplicate"
		}
		if row.Status != "resolved" {
			preview.Unresolved = append(preview.Unresolved, *row)
		}
	}

	return preview, nil
}

func (s *Service) CompleteGame(
	ctx context.Context,
	gameID int64,
	parts []CompleteParticipantInput,
	results []CompleteResultInput,
) (*domain.TournamentHistory, error) {
	g, err := s.Games.GetByID(ctx, gameID)
	if err != nil {
		return nil, err
	}
	if g == nil {
		return nil, ErrNotFound
	}
	if g.Completed {
		return nil, ErrCompletionAlreadyDone
	}

	preview, err := s.buildCompletePreview(ctx, g, results)
	if err != nil {
		return nil, err
	}
	if len(preview.Unresolved) > 0 {
		return nil, ErrCompletionUnresolved
	}

	liveParts, err := s.Participants.ListByGame(ctx, gameID)
	if err != nil {
		return nil, err
	}
	liveByUser := make(map[string]domain.Participant)
	for _, p := range liveParts {
		liveByUser[p.UserID] = p
	}
	inputByUser := make(map[string]CompleteParticipantInput)
	for _, p := range parts {
		p.UserID = strings.TrimSpace(p.UserID)
		if p.UserID == "" {
			continue
		}
		inputByUser[p.UserID] = p
	}

	re := reentryPrice(g)
	buyinI := int(math.Round(g.Buyin))
	reI := int(math.Round(re))
	var rePtr *int
	if g.ReentryBuyin > 0 {
		rePtr = &reI
	} else {
		rePtr = &buyinI
	}
	name := strings.TrimSpace(g.Name)
	if name == "" {
		name = strings.TrimSpace(g.Description)
	}
	if name == "" {
		name = fmt.Sprintf("Турнир %s", g.Date.Format("2006-01-02"))
	}

	h := &domain.TournamentHistory{
		GameID:            gameID,
		Date:              g.Date,
		Time:              timePtr(g.Time),
		TournamentName:    name,
		Location:          g.Location,
		Buyin:             buyinI,
		ReentryBuyin:      rePtr,
		ParticipantsCount: len(preview.Results),
	}
	if err := s.Tournaments.CreateHistory(ctx, h); err != nil {
		return nil, err
	}

	totalRev := 0
	for _, result := range preview.Results {
		u := result.User
		if u == nil {
			return nil, fmt.Errorf("resolved user missing for %s", result.Nickname)
		}

		p := inputByUser[u.UserID]
		if live, ok := liveByUser[u.UserID]; ok {
			if p.UserID == "" {
				p.UserID = live.UserID
			}
			if p.Entries <= 0 {
				p.Entries = live.Entries
			}
			if p.Rebuys <= 0 {
				p.Rebuys = live.Rebuys
			}
			if p.Addons <= 0 {
				p.Addons = live.Addons
			}
		}
		ent := p.Entries
		if ent <= 0 {
			ent = 1
		}
		spent := int(math.Round(float64(ent)*g.Buyin + float64(p.Rebuys)*re + float64(p.Addons)*re))
		totalRev += spent
		position := result.Position
		tp := &domain.TournamentParticipant{
			TournamentHistoryID: h.ID,
			UserID:              u.UserID,
			Username:            u.Username,
			FirstName:           derefStr(u.FirstName),
			LastName:            derefStr(u.LastName),
			Entries:             ent,
			Rebuys:              p.Rebuys,
			Addons:              p.Addons,
			TotalSpent:          spent,
			PaymentMethod:       p.PaymentMethod,
			Position:            &position,
			FinalPoints:         result.TotalPoints,
			KOCount:             result.KOCount,
		}
		if err := s.Tournaments.AddTournamentParticipant(ctx, tp); err != nil {
			return nil, err
		}

		u.Points += result.TotalPoints
		u.TotalGamesPlayed++
		if err := s.Users.Update(ctx, u); err != nil {
			return nil, err
		}
	}
	h.TotalRevenue = totalRev
	if err := s.Tournaments.UpdateHistory(ctx, h); err != nil {
		return nil, err
	}

	now := time.Now()
	g.IsActive = false
	g.Completed = true
	g.CompletedAt = &now
	if err := s.Games.Update(ctx, g); err != nil {
		return nil, err
	}
	full, err := s.Tournaments.GetHistoryByID(ctx, h.ID)
	if s.Clock != nil {
		live, _ := s.Participants.ListByGame(ctx, gameID)
		s.Clock.SyncParticipants(fmt.Sprint(gameID), live)
	}
	return full, err
}
