package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandlers(t *testing.T) {
	store = newMemoryStore()
	r := setupRouter()

	body, _ := json.Marshal(map[string]string{"username": "alice", "password": "secret"})
	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(body))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("register status %d", w.Code)
	}

	req = httptest.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(body))
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("login status %d", w.Code)
	}
	cookie := w.Result().Cookies()[0]

	msgBody, _ := json.Marshal(map[string]string{"content": "hello"})
	req = httptest.NewRequest(http.MethodPost, "/messages", bytes.NewBuffer(msgBody))
	req.AddCookie(cookie)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("post message status %d", w.Code)
	}

	req = httptest.NewRequest(http.MethodGet, "/feed", nil)
	req.AddCookie(cookie)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("feed status %d", w.Code)
	}
	var feed []Message
	if err := json.Unmarshal(w.Body.Bytes(), &feed); err != nil {
		t.Fatalf("unmarshal feed: %v", err)
	}
	if len(feed) != 1 || feed[0].Content != "hello" {
		t.Fatalf("unexpected feed %#v", feed)
	}
}
