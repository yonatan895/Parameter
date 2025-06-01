package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func setupTestServer() *gin.Engine {
	store = newMemoryStore()
	return setupRouter()
}

type credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func TestRegisterLoginPostFeed(t *testing.T) {
	r := setupTestServer()

	// Register user
	regBody, _ := json.Marshal(credentials{Username: "alice", Password: "pw"})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/register", bytes.NewBuffer(regBody))
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("register failed: %d", w.Code)
	}

	// Login user
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/login", bytes.NewBuffer(regBody))
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("login failed: %d", w.Code)
	}
	cookies := w.Result().Cookies()
	if len(cookies) == 0 {
		t.Fatal("no session cookie")
	}
	session := cookies[0]

	// Post message
	msgBody := []byte(`{"content":"hello"}`)
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/messages", bytes.NewBuffer(msgBody))
	req.AddCookie(session)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("post failed: %d", w.Code)
	}

	// Fetch feed
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/feed", nil)
	req.AddCookie(session)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("feed failed: %d", w.Code)
	}
	var feed []Message
	if err := json.Unmarshal(w.Body.Bytes(), &feed); err != nil {
		t.Fatalf("invalid feed: %v", err)
	}
	if len(feed) != 1 || feed[0].Content != "hello" {
		t.Fatalf("unexpected feed: %+v", feed)
	}
}
