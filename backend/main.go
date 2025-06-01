package main

import (
    "context"
    "fmt"
    "log"
    "net/http"
    "time"

    "github.com/gin-gonic/gin"
    "github.com/jackc/pgx/v5/pgxpool"
)

var db *pgxpool.Pool

func main() {
    ctx := context.Background()

    var err error
    db, err = pgxpool.New(ctx, "postgres://user:password@localhost:5432/twitter")
    if err != nil {
        log.Fatalf("failed to connect to postgres: %v", err)
    }
    defer db.Close()

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
    err := db.QueryRow(c, "INSERT INTO users (username, password) VALUES ($1, $2) RETURNING id", u.Username, u.Password).Scan(&u.ID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusOK, u)
}

func loginHandler(c *gin.Context) {
    var u User
    if err := c.BindJSON(&u); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
        return
    }
    row := db.QueryRow(c, "SELECT id FROM users WHERE username=$1 AND password=$2", u.Username, u.Password)
    if err := row.Scan(&u.ID); err != nil {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
        return
    }
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
    err := db.QueryRow(c, "INSERT INTO messages (user_id, content) VALUES ($1, $2) RETURNING id, created_at", userID, m.Content).Scan(&m.ID, &m.CreatedAt)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    fmt.Sscanf(userID, "%d", &m.UserID)
    c.JSON(http.StatusOK, m)
}

func feedHandler(c *gin.Context) {
    userID := c.GetString("userID")
    rows, err := db.Query(c, "SELECT id, user_id, content, created_at FROM messages WHERE user_id=$1 ORDER BY created_at DESC LIMIT 20", userID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    defer rows.Close()
    feed := []Message{}
    for rows.Next() {
        var m Message
        if err := rows.Scan(&m.ID, &m.UserID, &m.Content, &m.CreatedAt); err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }
        feed = append(feed, m)
    }
    c.JSON(http.StatusOK, feed)
}

func generateTraffic(ctx context.Context) {
    for {
        time.Sleep(5 * time.Second)
        _, err := db.Exec(ctx, "INSERT INTO messages (user_id, content) SELECT id, 'random post #' || floor(random()*1000)::int FROM users")
        if err != nil {
            log.Println("traffic error:", err)
        }
    }
}

