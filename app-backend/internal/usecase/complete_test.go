package usecase

import (
	"testing"

	"github.com/pridecrm/app-backend/internal/domain"
)

func TestRatingPlacePointsUsesTopThirtyPercent(t *testing.T) {
	game := &domain.Game{BasePoints: 100}

	tests := []struct {
		total    int
		position int
		want     int
	}{
		{total: 10, position: 1, want: 300},
		{total: 10, position: 3, want: 100},
		{total: 10, position: 4, want: 0},
		{total: 19, position: 1, want: 600},
		{total: 19, position: 6, want: 100},
		{total: 19, position: 7, want: 0},
	}

	for _, tt := range tests {
		if got := ratingPlacePoints(game, tt.total, tt.position); got != tt.want {
			t.Fatalf("ratingPlacePoints(total=%d, position=%d) = %d, want %d", tt.total, tt.position, got, tt.want)
		}
	}
}

func TestRatingPlacePointsAddsExtraPlayerMultiplier(t *testing.T) {
	game := &domain.Game{
		BasePoints:               100,
		PointsPerExtraPlayer:     10,
		MinPlayersForExtraPoints: 10,
	}

	if got := ratingPlacePoints(game, 12, 1); got != 480 {
		t.Fatalf("ratingPlacePoints = %d, want 480", got)
	}
}

func TestMatchUserByNicknameNormalizesCommonVariants(t *testing.T) {
	nick := "DelureKing"
	users := []domain.User{
		{UserID: "1", Username: "someone_else"},
		{UserID: "2", Username: "someone", NickName: &nick},
	}

	matches := matchUserByNickname(users, "delure king")
	if len(matches) != 1 {
		t.Fatalf("matches count = %d, want 1", len(matches))
	}
	if matches[0].UserID != "2" {
		t.Fatalf("matched user = %s, want 2", matches[0].UserID)
	}
}
