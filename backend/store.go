package main

import (
	"context"
	"errors"
	"os"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Store abstracts database operations so they can be implemented by
// different backends, such as Postgres or an in-memory map used in tests.
type Store interface {
	CreateUser(ctx context.Context, username, password string) (int64, error)
	GetUserID(ctx context.Context, username, password string) (int64, error)
	CreateMessage(ctx context.Context, userID int64, content string) (Message, error)
	GetFeed(ctx context.Context, userID int64) ([]Message, error)
	Close()
}

// pgStore implements Store backed by Postgres.
type pgStore struct {
	db *pgxpool.Pool
}

func newPGStore(ctx context.Context) (*pgStore, error) {
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		url = "postgres://user:password@localhost:5432/twitter"
	}
	db, err := pgxpool.New(ctx, url)
	if err != nil {
		return nil, err
	}
	return &pgStore{db: db}, nil
}

func (p *pgStore) CreateUser(ctx context.Context, username, password string) (int64, error) {
	var id int64
	err := p.db.QueryRow(ctx,
		"INSERT INTO users (username, password) VALUES ($1,$2) RETURNING id",
		username, password,
	).Scan(&id)
	return id, err
}

func (p *pgStore) GetUserID(ctx context.Context, username, password string) (int64, error) {
	var id int64
	err := p.db.QueryRow(ctx,
		"SELECT id FROM users WHERE username=$1 AND password=$2",
		username, password,
	).Scan(&id)
	return id, err
}

func (p *pgStore) CreateMessage(ctx context.Context, userID int64, content string) (Message, error) {
	var m Message
	err := p.db.QueryRow(ctx,
		"INSERT INTO messages (user_id, content) VALUES ($1,$2) RETURNING id, created_at",
		userID, content,
	).Scan(&m.ID, &m.CreatedAt)
	m.UserID = userID
	m.Content = content
	return m, err
}

func (p *pgStore) GetFeed(ctx context.Context, userID int64) ([]Message, error) {
	rows, err := p.db.Query(ctx,
		"SELECT id, user_id, content, created_at FROM messages WHERE user_id=$1 ORDER BY created_at DESC LIMIT 20",
		userID,
	)
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
	return feed, rows.Err()
}

func (p *pgStore) Close() {
	p.db.Close()
}

// memoryStore is an in-memory implementation for tests.
type memoryStore struct {
	mu       sync.Mutex
	nextID   int64
	users    []User
	messages []Message
}

func newMemoryStore() *memoryStore {
	return &memoryStore{nextID: 1}
}

func (m *memoryStore) CreateUser(ctx context.Context, username, password string) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, u := range m.users {
		if u.Username == username {
			return 0, errors.New("user exists")
		}
	}
	id := m.nextID
	m.nextID++
	m.users = append(m.users, User{ID: id, Username: username, Password: password})
	return id, nil
}

func (m *memoryStore) GetUserID(ctx context.Context, username, password string) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, u := range m.users {
		if u.Username == username && u.Password == password {
			return u.ID, nil
		}
	}
	return 0, errors.New("not found")
}

func (m *memoryStore) CreateMessage(ctx context.Context, userID int64, content string) (Message, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	msg := Message{ID: m.nextID, UserID: userID, Content: content, CreatedAt: time.Now()}
	m.nextID++
	m.messages = append(m.messages, msg)
	return msg, nil
}

func (m *memoryStore) GetFeed(ctx context.Context, userID int64) ([]Message, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	feed := []Message{}
	for i := len(m.messages) - 1; i >= 0 && len(feed) < 20; i-- {
		if m.messages[i].UserID == userID {
			feed = append(feed, m.messages[i])
		}
	}
	return feed, nil
}

func (m *memoryStore) Close() {}
