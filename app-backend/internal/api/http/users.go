package httpapi

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pridecrm/app-backend/internal/domain"
)

// --- Users (admin) ---

func (h *Handlers) ListUsers(c *gin.Context) {
	users, err := h.Repo.Users.List(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	out := make([]map[string]any, 0, len(users))
	for i := range users {
		out = append(out, userToMap(&users[i]))
	}
	c.JSON(http.StatusOK, out)
}

func (h *Handlers) CreateUser(c *gin.Context) {
	var body userCreateBody
	if err := c.ShouldBindJSON(&body); err != nil || body.UserID == "" || body.Username == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id and username required"})
		return
	}
	u := &domain.User{
		UserID:      body.UserID,
		Username:    body.Username,
		FirstName:   body.FirstName,
		LastName:    body.LastName,
		PhoneNumber: body.Phone,
		Email:       body.Email,
		IsActive:    true,
	}
	if err := h.Repo.Users.Create(c.Request.Context(), u); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, userToMap(u))
}

func (h *Handlers) GetUser(c *gin.Context) {
	id := c.Param("user_id")
	u, err := h.Repo.Users.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if u == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.JSON(http.StatusOK, userToMap(u))
}
func (h *Handlers) UpdateUser(c *gin.Context) {
	id := c.Param("user_id")
	u, err := h.Repo.Users.GetByID(c.Request.Context(), id)
	if err != nil || u == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	var body userPatch
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if body.Username != nil {
		u.Username = *body.Username
	}
	if body.NickName != nil {
		u.NickName = body.NickName
	}
	if body.FirstName != nil {
		u.FirstName = body.FirstName
	}
	if body.LastName != nil {
		u.LastName = body.LastName
	}
	if body.Phone != nil {
		u.PhoneNumber = body.Phone
	}
	if body.Email != nil {
		u.Email = body.Email
	}
	if body.DOB != nil && *body.DOB != "" {
		t, err := time.Parse("2006-01-02", *body.DOB)
		if err == nil {
			u.DateOfBirth = &t
		}
	}
	if err := h.Repo.Users.Update(c.Request.Context(), u); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, userToMap(u))
}

func (h *Handlers) DeleteUser(c *gin.Context) {
	id := c.Param("user_id")
	if err := h.Repo.Users.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Handlers) BanUser(c *gin.Context) {
	h.setBanned(c, true)
}

func (h *Handlers) UnbanUser(c *gin.Context) {
	h.setBanned(c, false)
}

func (h *Handlers) setBanned(c *gin.Context, banned bool) {
	id := c.Param("user_id")
	u, err := h.Repo.Users.GetByID(c.Request.Context(), id)
	if err != nil || u == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	u.IsBanned = banned
	if err := h.Repo.Users.Update(c.Request.Context(), u); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if banned {
		c.JSON(http.StatusOK, gin.H{"status": "user banned"})
	} else {
		c.JSON(http.StatusOK, gin.H{"status": "user unbanned"})
	}
}

type addPointsBody struct {
	Points int `json:"points"`
}

func (h *Handlers) AddPoints(c *gin.Context) {
	id := c.Param("user_id")
	var body addPointsBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid points value"})
		return
	}
	u, err := h.Repo.Users.GetByID(c.Request.Context(), id)
	if err != nil || u == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	u.Points += body.Points
	if err := h.Repo.Users.Update(c.Request.Context(), u); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "points added", "new_points": u.Points})
}
