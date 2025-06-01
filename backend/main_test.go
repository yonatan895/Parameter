package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandlers(t *testing.T) {
	s := newMemoryStore()
	router := setupRouter(s)

	// register
	reqBody := bytes.NewBufferString(`{"username":"alice","password":"pw"}`)
	req := httptest.NewRequest(http.MethodPost, "/register", reqBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("register failed: %s", w.Body.String())
	}
	var user User
	if err := json.Unmarshal(w.Body.Bytes(), &user); err != nil {
		t.Fatal(err)
	}

	// login
	reqBody = bytes.NewBufferString(`{"username":"alice","password":"pw"}`)
	req = httptest.NewRequest(http.MethodPost, "/login", reqBody)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("login failed: %s", w.Body.String())
	}
	cookie := w.Result().Cookies()[0]

	// post message
	reqBody = bytes.NewBufferString(`{"content":"hello"}`)
	req = httptest.NewRequest(http.MethodPost, "/messages", reqBody)
	req.AddCookie(cookie)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("post failed: %s", w.Body.String())
	}

	// feed
	req = httptest.NewRequest(http.MethodGet, "/feed", nil)
	req.AddCookie(cookie)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("feed failed: %s", w.Body.String())
	}
	var msgs []Message
	if err := json.Unmarshal(w.Body.Bytes(), &msgs); err != nil {
		t.Fatal(err)
	}
	if len(msgs) != 1 || msgs[0].Content != "hello" {
		t.Fatalf("unexpected feed: %+v", msgs)
	}
}
