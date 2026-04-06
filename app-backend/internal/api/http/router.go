package httpapi

import (
	"log/slog"

	"github.com/gin-gonic/gin"

	"github.com/pridecrm/app-backend/internal/infrastructure/auth"
)

func Mount(
	r *gin.Engine,
	h *Handlers,
	jwt *auth.JWTService,
	log *slog.Logger,
) {
	r.GET("/health", h.Health)

	api := r.Group("/api")

	api.POST("/auth/telegram/", h.TelegramAuth)
	api.POST("/auth/telegram/validate/", h.TelegramValidate)

	opt := auth.MiddlewareJWTOptional(jwt, log)
	jwtMW := auth.MiddlewareJWT(jwt, log)
	adm := auth.MiddlewareAdmin(log)

	api.GET("/games", opt, h.ListGames)
	api.GET("/games/:id", opt, h.GetGame)

	gamesAdm := api.Group("/games")
	gamesAdm.Use(jwtMW, adm)
	{
		gamesAdm.POST("", h.CreateGame)
		gamesAdm.PATCH("/:id", h.UpdateGame)
		gamesAdm.PUT("/:id", h.UpdateGame)
		gamesAdm.DELETE("/:id", h.DeleteGame)
		gamesAdm.GET("/:id/participants_admin", h.GameParticipantsAdmin)
		gamesAdm.POST("/:id/add_participant_admin", h.GameAddParticipantAdmin)
		gamesAdm.POST("/:id/remove_participant_admin", h.GameRemoveParticipantAdmin)
		gamesAdm.POST("/:id/complete", h.GameComplete)
		gamesAdm.POST("/:id/update_participant_admin", h.GameUpdateParticipantAdmin)
	}

	usersAdm := api.Group("/users")
	usersAdm.Use(jwtMW, adm)
	{
		usersAdm.GET("", h.ListUsers)
		usersAdm.POST("", h.CreateUser)
		usersAdm.GET("/:user_id", h.GetUser)
		usersAdm.PATCH("/:user_id", h.UpdateUser)
		usersAdm.PUT("/:user_id", h.UpdateUser)
		usersAdm.DELETE("/:user_id", h.DeleteUser)
		usersAdm.POST("/:user_id/ban", h.BanUser)
		usersAdm.POST("/:user_id/unban", h.UnbanUser)
		usersAdm.POST("/:user_id/add_points", h.AddPoints)
	}

	part := api.Group("/participants")
	part.Use(jwtMW)
	{
		part.GET("", h.ListParticipants)
		part.POST("/register", h.RegisterParticipant)
		part.DELETE("/unregister", h.UnregisterParticipant)
		part.GET("/:id", h.GetParticipant)
		part.POST("", h.CreateParticipant)
		part.PATCH("/:id", h.UpdateParticipant)
		part.PUT("/:id", h.UpdateParticipant)
		part.DELETE("/:id", h.DeleteParticipant)
		part.POST("/:id/arrived", h.SetParticipantArrived)
	}

	tick := api.Group("/support-tickets")
	tick.Use(jwtMW)
	{
		tick.GET("", h.ListTickets)
		tick.POST("", h.CreateTicket)
		tick.GET("/:id", h.GetTicket)
		tick.PATCH("/:id", h.UpdateTicket)
		tick.PUT("/:id", h.UpdateTicket)
		tick.DELETE("/:id", h.DeleteTicket)
	}

	th := api.Group("/tournament-history")
	th.Use(jwtMW, adm)
	{
		th.GET("", h.ListTournamentHistory)
		th.POST("", h.CreateTournamentHistory)
		th.GET("/:id/participants", h.TournamentParticipants)
		th.GET("/:id", h.GetTournamentHistory)
		th.PATCH("/:id", h.UpdateTournamentHistory)
		th.PUT("/:id", h.UpdateTournamentHistory)
		th.DELETE("/:id", h.DeleteTournamentHistory)
	}

	authed := api.Group("")
	authed.Use(jwtMW)
	{
		authed.GET("/rating", h.Rating)
		authed.GET("/profile", h.ProfileGet)
		authed.PATCH("/profile", h.ProfilePatch)
		authed.GET("/admin/dashboard", h.AdminDashboard)
	}
}
