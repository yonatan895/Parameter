package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

var store Store

func main() {
	ctx := context.Background()

	var err error
	store, err = newPGStore(ctx)
	if err != nil {
		log.Fatalf("failed to connect to store: %v", err)
	}
	defer store.Close()

	go generateTraffic(ctx)

	r := setupRouter()

	log.Println("server running on :8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}

func setupRouter() *gin.Engine {
	r := gin.Default()
	r.POST("/register", registerHandler)
	r.POST("/login", loginHandler)
	r.POST("/messages", authMiddleware, postMessageHandler)
	r.GET("/feed", authMiddleware, feedHandler)
	return r
}

type User struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type Message struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

func registerHandler(c *gin.Context) {
	var u User
	if err := c.BindJSON(&u); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
		return
	}
	id, err := store.CreateUser(c, u.Username, u.Password)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	u.ID = id
	c.JSON(http.StatusOK, u)
}

func loginHandler(c *gin.Context) {
	var u User
	if err := c.BindJSON(&u); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
		return
	}
	id, err := store.GetUserID(c, u.Username, u.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}
	u.ID = id
	http.SetCookie(c.Writer, &http.Cookie{Name: "session", Value: fmt.Sprint(u.ID), Path: "/"})
	c.JSON(http.StatusOK, u)
}

func authMiddleware(c *gin.Context) {
	cookie, err := c.Request.Cookie("session")
	if err != nil {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	c.Set("userID", cookie.Value)
	c.Next()
}

func postMessageHandler(c *gin.Context) {
	userIDStr := c.GetString("userID")
	var userID int64
	fmt.Sscanf(userIDStr, "%d", &userID)
	var m Message
	if err := c.BindJSON(&m); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
		return
	}
	msg, err := store.CreateMessage(c, userID, m.Content)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, msg)
}

func feedHandler(c *gin.Context) {
	userIDStr := c.GetString("userID")
	var userID int64
	fmt.Sscanf(userIDStr, "%d", &userID)
	feed, err := store.GetFeed(c, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, feed)
}

func generateTraffic(ctx context.Context) {
	pg, ok := store.(*pgStore)
	if !ok {
		return
	}
	for {
		time.Sleep(5 * time.Second)
		_, err := pg.db.Exec(ctx, "INSERT INTO messages (user_id, content) SELECT id, 'random post #' || floor(random()*1000)::int FROM users")
		if err != nil {
			log.Println("traffic error:", err)
		}
	}
}