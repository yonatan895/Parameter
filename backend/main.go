package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

var store Store

func main() {
	ctx := context.Background()

	db, err := pgxpool.New(ctx, "postgres://user:password@localhost:5432/twitter")
	if err != nil {
		log.Fatalf("failed to connect to postgres: %v", err)
	}
	defer db.Close()

	store = newPGStore(db)

	go generateTraffic(ctx)

	r := setupRouter()

	log.Println("server running on :8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatal(err)
	}
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
	userID := c.GetString("userID")
	var m Message
	if err := c.BindJSON(&m); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
		return
	}
	uid, err := strconv.ParseInt(userID, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user"})
		return
	}
	m, err = store.CreateMessage(c, uid, m.Content)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, m)
}

func feedHandler(c *gin.Context) {
	userID := c.GetString("userID")
	uid, err := strconv.ParseInt(userID, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user"})
		return
	}
	feed, err := store.GetFeed(c, uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, feed)
}

func generateTraffic(ctx context.Context) {
	for {
		time.Sleep(5 * time.Second)
		// Insert a message for user 1 to simulate activity
		_, err := store.CreateMessage(ctx, 1, fmt.Sprintf("random post #%d", rand.Intn(1000)))
		if err != nil {
			log.Println("traffic error:", err)
		}
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
