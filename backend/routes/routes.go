package routes

import (
	"github.com/gin-gonic/gin"
	config "github.com/phillip/contribution-tracker-go/config"
	controllers "github.com/phillip/contribution-tracker-go/controllers"
	middleware "github.com/phillip/contribution-tracker-go/middleware"
)

func SetupRoutes(r *gin.Engine, cfg *config.Config) {
	// public
	r.POST("/auth/register", controllers.Register(cfg))
	r.POST("/auth/login", controllers.Login(cfg))
	r.POST("/auth/refresh", controllers.RefreshToken(cfg))

	// otp
	r.POST("/auth/request-otp", controllers.RequestOTP(cfg))
	r.POST("/auth/verify-otp", controllers.VerifyOTP(cfg))

	// protected
	auth := middleware.AuthMiddleware(cfg)

	users := r.Group("/users")
	users.Use(auth)
	{
		// users.POST("", controllers.ListUsers(cfg))
		users.GET("", controllers.ListUsers(cfg))
		users.GET(":id", controllers.GetUser(cfg))
		users.PATCH(":id", controllers.UpdateUser(cfg))
		users.DELETE(":id", controllers.DeleteUser(cfg))
	}

	notifs := r.Group("/notifications")
	notifs.Use(auth) // protected
	{
		notifs.GET("", controllers.ListNotifications(cfg))
		notifs.PATCH("/:id/read", controllers.MarkNotificationRead(cfg))
	}

	// Events
	events := r.Group("/events")
	events.Use(auth)
	{
		events.POST("", controllers.CreateEvent(cfg))
		events.GET("", controllers.ListEvents(cfg))
		events.GET("/:id", controllers.GetEvent(cfg))
		events.PATCH("/:id", controllers.UpdateEvent(cfg))
		events.DELETE("/:id", controllers.DeleteEvent(cfg))
	}

	// Contributions
	contribs := r.Group("/contributions")
	contribs.Use(auth)
	{
		contribs.POST("", controllers.CreateContribution(cfg))
		contribs.GET("", controllers.ListContributions(cfg))
		contribs.GET("/:id", controllers.GetContribution(cfg))
		contribs.PATCH("/:id", controllers.UpdateContribution(cfg))
		contribs.DELETE("/:id", controllers.DeleteContribution(cfg))
	}

}
