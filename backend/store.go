package main

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Store interface {
	CreateUser(ctx context.Context, username, password string) (int64, error)
	GetUserID(ctx context.Context, username, password string) (int64, error)
	CreateMessage(ctx context.Context, userID int64, content string) (Message, error)
	GetFeed(ctx context.Context, userID int64) ([]Message, error)
}

type pgStore struct {
	db *pgxpool.Pool
}

func newPGStore(db *pgxpool.Pool) *pgStore { return &pgStore{db: db} }

func (s *pgStore) CreateUser(ctx context.Context, username, password string) (int64, error) {
	var id int64
	err := s.db.QueryRow(ctx, "INSERT INTO users (username, password) VALUES ($1, $2) RETURNING id", username, password).Scan(&id)
	return id, err
}

func (s *pgStore) GetUserID(ctx context.Context, username, password string) (int64, error) {
	var id int64
	err := s.db.QueryRow(ctx, "SELECT id FROM users WHERE username=$1 AND password=$2", username, password).Scan(&id)
	return id, err
}

func (s *pgStore) CreateMessage(ctx context.Context, userID int64, content string) (Message, error) {
	var m Message
	err := s.db.QueryRow(ctx, "INSERT INTO messages (user_id, content) VALUES ($1, $2) RETURNING id, created_at", userID, content).Scan(&m.ID, &m.CreatedAt)
	m.UserID = userID
	m.Content = content
	return m, err
}

func (s *pgStore) GetFeed(ctx context.Context, userID int64) ([]Message, error) {
	rows, err := s.db.Query(ctx, "SELECT id, user_id, content, created_at FROM messages WHERE user_id=$1 ORDER BY created_at DESC LIMIT 20", userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var feed []Message
	for rows.Next() {
		var m Message
		if err := rows.Scan(&m.ID, &m.UserID, &m.Content, &m.CreatedAt); err != nil {
			return nil, err
		}
		feed = append(feed, m)
	}
	return feed, nil
}

type memoryStore struct {
	mu      sync.Mutex
	nextUID int64
	nextMID int64
	users   map[string]User
	msgs    map[int64][]Message
}

func newMemoryStore() *memoryStore {
	return &memoryStore{
		users:   make(map[string]User),
		msgs:    make(map[int64][]Message),
		nextUID: 1,
		nextMID: 1,
	}
}

func (s *memoryStore) CreateUser(ctx context.Context, username, password string) (int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.users[username]; exists {
		return 0, errors.New("exists")
	}
	id := s.nextUID
	s.nextUID++
	s.users[username] = User{ID: id, Username: username, Password: password}
	return id, nil
}

func (s *memoryStore) GetUserID(ctx context.Context, username, password string) (int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	u, ok := s.users[username]
	if !ok || u.Password != password {
		return 0, errors.New("not found")
	}
	return u.ID, nil
}

func (s *memoryStore) CreateMessage(ctx context.Context, userID int64, content string) (Message, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	m := Message{
		ID:        s.nextMID,
		UserID:    userID,
		Content:   content,
		CreatedAt: time.Now(),
	}
	s.nextMID++
	s.msgs[userID] = append([]Message{m}, s.msgs[userID]...)
	return m, nil
}

func (s *memoryStore) GetFeed(ctx context.Context, userID int64) ([]Message, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	feed := s.msgs[userID]
	if len(feed) > 20 {
		feed = feed[:20]
	}
	cp := make([]Message, len(feed))
	copy(cp, feed)
	return cp, nil
}
