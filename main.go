package main

import (
	"fmt"
	"log"
	"os"

	"qr-scan/internal/database"
	"qr-scan/internal/handlers"
	"qr-scan/internal/repository"
	ws "qr-scan/internal/websocket"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
)

func main() {
	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		host := os.Getenv("DB_HOST")
		user := os.Getenv("DB_USER")
		pass := os.Getenv("DB_PASSWORD")
		dbName := os.Getenv("DB_NAME")
		port := os.Getenv("DB_PORT")

		if host != "" && user != "" && dbName != "" {
			if port == "" {
				port = "3306"
			}
			dsn = fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true", user, pass, host, port, dbName)
		} else {
			dsn = "root:password@tcp(127.0.0.1:3306)/qrscan?parseTime=true"
		}
	}
	db := database.InitDB(dsn)
	defer db.Close()

	repo := repository.NewMemberRepository(db)
	hub := ws.NewHub()
	go hub.Run()

	h := handlers.NewHandler(repo, hub)

	r := gin.Default()

	store := cookie.NewStore([]byte("secret_session_key"))
	r.Use(sessions.Sessions("mysession", store))

	r.LoadHTMLGlob("templates/*")

	// Routes
	r.GET("/", h.LoginPage)
	r.POST("/login", h.Login)
	r.GET("/logout", h.Logout)

	r.GET("/dashboard", h.UserDashboard)
	r.POST("/api/checkin", h.DoCheckin)

	r.GET("/admin", h.AdminDashboard)
	r.POST("/api/admin/toggle", h.ToggleCheckin)

	r.GET("/ws", h.WSHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server running on port %s", port)
	r.Run(":" + port)
}
