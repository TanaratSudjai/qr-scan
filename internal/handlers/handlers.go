package handlers

import (
	"log"
	"net/http"

	"qr-scan/internal/repository"
	ws "qr-scan/internal/websocket"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type Handler struct {
	Repo *repository.MemberRepository
	Hub  *ws.Hub
}

func NewHandler(repo *repository.MemberRepository, hub *ws.Hub) *Handler {
	return &Handler{
		Repo: repo,
		Hub:  hub,
	}
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all for simplicity
	},
}

// WSHandler handles websocket requests
func (h *Handler) WSHandler(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println(err)
		return
	}
	h.Hub.Register(conn)

	// Read pump
	go func() {
		defer func() {
			h.Hub.Unregister(conn)
			conn.Close()
		}()
		for {
			_, _, err := conn.ReadMessage()
			if err != nil {
				break
			}
		}
	}()
}

// LoginPage renders the login form
func (h *Handler) LoginPage(c *gin.Context) {
	session := sessions.Default(c)
	phone := session.Get("phoneNumber")
	if phone != nil {
		c.Redirect(http.StatusFound, "/dashboard")
		return
	}
	c.HTML(http.StatusOK, "login.html", nil)
}

// Login handles the form submission
func (h *Handler) Login(c *gin.Context) {
	phone := c.PostForm("phoneNumber")
	member, err := h.Repo.GetByPhoneNumber(phone)
	if err != nil || member == nil {
		c.HTML(http.StatusOK, "login.html", gin.H{"Error": "Invalid phone number or user not found"})
		return
	}

	session := sessions.Default(c)
	session.Set("phoneNumber", member.PhoneNumber)
	session.Set("userID", member.ID)
	session.Set("isAdmin", member.CountCheckin >= 99)
	session.Save()

	if member.CountCheckin >= 99 {
		c.Redirect(http.StatusFound, "/admin")
	} else {
		c.Redirect(http.StatusFound, "/dashboard")
	}
}

// Logout clears the session
func (h *Handler) Logout(c *gin.Context) {
	session := sessions.Default(c)
	session.Clear()
	session.Save()
	c.Redirect(http.StatusFound, "/")
}

// UserDashboard renders the user dashboard
func (h *Handler) UserDashboard(c *gin.Context) {
	session := sessions.Default(c)
	phone := session.Get("phoneNumber")
	if phone == nil {
		c.Redirect(http.StatusFound, "/")
		return
	}

	member, err := h.Repo.GetByPhoneNumber(phone.(string))
	if err != nil || member == nil {
		c.Redirect(http.StatusFound, "/")
		return
	}

	// Redirect to admin if count >= 99 to be safe
	if member.CountCheckin >= 99 {
		c.Redirect(http.StatusFound, "/admin")
		return
	}

	c.HTML(http.StatusOK, "user.html", gin.H{
		"Member":       member,
		"IsOpen":       h.Hub.IsOpen(),
		"HasCheckedIn": h.Hub.HasCheckedIn(member.ID),
		"SessionID":    h.Hub.CurrentSessionID(),
	})
}

// DoCheckin handles the actual checkin request
func (h *Handler) DoCheckin(c *gin.Context) {
	session := sessions.Default(c)
	userID := session.Get("userID")
	if userID == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	if !h.Hub.IsOpen() {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Check-in is closed"})
		return
	}

	if h.Hub.HasCheckedIn(userID.(int)) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "You have already checked in for this session"})
		return
	}

	err := h.Repo.IncrementCheckin(userID.(int))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check in"})
		return
	}

	h.Hub.MarkCheckedIn(userID.(int))

	c.JSON(http.StatusOK, gin.H{"message": "Check-in successful"})
}

// AdminDashboard renders the admin dashboard
func (h *Handler) AdminDashboard(c *gin.Context) {
	session := sessions.Default(c)
	isAdmin := session.Get("isAdmin")
	if isAdmin == nil || isAdmin.(bool) == false {
		c.Redirect(http.StatusFound, "/dashboard") // or redirect to login
		return
	}

	members, err := h.Repo.GetAll()
	if err != nil {
		c.String(http.StatusInternalServerError, "Failed to load members")
		return
	}

	c.HTML(http.StatusOK, "admin.html", gin.H{
		"Members": members,
		"IsOpen":  h.Hub.IsOpen(),
	})
}

// ToggleCheckin handles the admin request to open/close check-in
func (h *Handler) ToggleCheckin(c *gin.Context) {
	session := sessions.Default(c)
	isAdmin := session.Get("isAdmin")
	if isAdmin == nil || isAdmin.(bool) == false {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req struct {
		IsOpen bool `json:"is_open"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	h.Hub.SetCheckinStatus(req.IsOpen)
	c.JSON(http.StatusOK, gin.H{"status": "updated", "is_open": req.IsOpen})
}
