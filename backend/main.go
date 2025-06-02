package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/segmentio/kafka-go"
)

var (
	db          *pgxpool.Pool
	store       Store
	redisClient *redis.Client
	kafkaWriter *kafka.Writer
)

// main is the entry point of the backend service. It initialises all external
// dependencies and starts the HTTP server.
func main() {
	ctx := context.Background()

	redisClient = newRedisClient()
	defer func() {
		if err := redisClient.Close(); err != nil {
			log.Println("redis close:", err)
		}
	}()

	kafkaWriter = newKafkaWriter()
	defer func() {
		if err := kafkaWriter.Close(); err != nil {
			log.Println("kafka close:", err)
		}
	}()

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://user:password@localhost:5432/twitter"
	}

	var err error
	db, err = pgxpool.New(ctx, dsn)
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

// registerHandler handles user registration requests.
func registerHandler(c *gin.Context) {
	var u User
	if err := c.BindJSON(&u); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
		return
	}
	created, err := store.CreateUser(c, u.Username, u.Password)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, created)
}

// loginHandler handles user login and sets a session cookie.
func loginHandler(c *gin.Context) {
	var u User
	if err := c.BindJSON(&u); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
		return
	}
	found, err := store.GetUserByCredentials(c, u.Username, u.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}
	http.SetCookie(c.Writer, &http.Cookie{Name: "session", Value: fmt.Sprint(found.ID), Path: "/"})
	c.JSON(http.StatusOK, found)
}

// authMiddleware ensures a valid session cookie is present before allowing
// access to protected endpoints.
func authMiddleware(c *gin.Context) {
	cookie, err := c.Request.Cookie("session")
	if err != nil {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	c.Set("userID", cookie.Value)
	c.Next()
}

// postMessageHandler stores a new message, caches it in Redis and publishes an
// event to Kafka.
func postMessageHandler(c *gin.Context) {
	userID := c.GetString("userID")
	var m Message
	if err := c.BindJSON(&m); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
		return
	}
	id, err := strconv.ParseInt(userID, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user"})
		return
	}
	msg, err := store.CreateMessage(c, id, m.Content)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if err := redisClient.Set(c, fmt.Sprintf("message:%d", msg.ID), m.Content, 0).Err(); err != nil {
		log.Println("redis set:", err)
	}
	if err := kafkaWriter.WriteMessages(c, kafka.Message{Value: []byte(m.Content)}); err != nil {
		log.Println("kafka write:", err)
	}
	c.JSON(http.StatusOK, msg)
}

// feedHandler returns the 20 most recent messages for the authenticated user.
func feedHandler(c *gin.Context) {
	userID := c.GetString("userID")
	id, err := strconv.ParseInt(userID, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user"})
		return
	}
	feed, err := store.GetFeed(c, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, feed)
}

// setupRouter wires up all HTTP routes and returns the Gin engine.
func setupRouter() *gin.Engine {
	r := gin.Default()
	r.POST("/register", registerHandler)
	r.POST("/login", loginHandler)
	r.POST("/messages", authMiddleware, postMessageHandler)
	r.GET("/feed", authMiddleware, feedHandler)
	return r
}

// generateTraffic periodically inserts random messages to keep the demo
// populated with data.
func generateTraffic(ctx context.Context) {
	for {
		time.Sleep(5 * time.Second)
		_, err := db.Exec(ctx, "INSERT INTO messages (user_id, content) SELECT id, 'random post #' || floor(random()*1000)::int FROM users")
		if err != nil {
			log.Println("traffic error:", err)
		}
	}
}
