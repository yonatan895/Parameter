package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"

)

var store Store

func main() {
	ctx := context.Background()


	url := os.Getenv("DATABASE_URL")
	if url == "" {
		url = "postgres://user:password@localhost:5432/twitter"
	}
	var err error
	store, err = newPGStore(ctx, url)
	if err != nil {
		log.Fatalf("failed to connect to postgres: %v", err)
	}
	defer store.Close()

	go generateTraffic(ctx)

	r := gin.Default()
	r.POST("/register", registerHandler)
	r.POST("/login", loginHandler)
	r.POST("/messages", authMiddleware, postMessageHandler)
	r.GET("/feed", authMiddleware, feedHandler)

	log.Println("server running on :8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatal(err)
	}

}

func setupRouter(s Store) *gin.Engine {
	store = s
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

	id, err := store.GetUserByCredentials(c, u.Username, u.Password)

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
	var body struct {
		Content string `json:"content"`
	}
	if err := c.BindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
		return
	}
	var uid int64
	fmt.Sscanf(userIDStr, "%d", &uid)
	msg, err := store.CreateMessage(c, uid, body.Content)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, msg)
}

func feedHandler(c *gin.Context) {
	userIDStr := c.GetString("userID")
	var uid int64
	fmt.Sscanf(userIDStr, "%d", &uid)
	feed, err := store.ListMessages(c, uid, 20)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, feed)
}

func generateTraffic(ctx context.Context) {
	for {
		time.Sleep(5 * time.Second)
    
		_, err := store.CreateMessage(ctx, 1, fmt.Sprintf("random post #%d", time.Now().UnixNano()))

		if err != nil {
			log.Println("traffic error:", err)
		}
	}
}

