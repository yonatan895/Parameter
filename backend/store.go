package main

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Store interface {
	CreateUser(ctx context.Context, username, password string) (int64, error)
	GetUserByCredentials(ctx context.Context, username, password string) (int64, error)
	CreateMessage(ctx context.Context, userID int64, content string) (Message, error)
	ListMessages(ctx context.Context, userID int64, limit int) ([]Message, error)
	Close()
}

type pgStore struct {
	db *pgxpool.Pool
}

func newPGStore(ctx context.Context, url string) (*pgStore, error) {
	pool, err := pgxpool.New(ctx, url)
	if err != nil {
		return nil, err
	}
	return &pgStore{db: pool}, nil
}

func (p *pgStore) CreateUser(ctx context.Context, username, password string) (int64, error) {
	var id int64
	err := p.db.QueryRow(ctx, "INSERT INTO users (username, password) VALUES ($1, $2) RETURNING id", username, password).Scan(&id)
	return id, err
}

func (p *pgStore) GetUserByCredentials(ctx context.Context, username, password string) (int64, error) {
	var id int64
	err := p.db.QueryRow(ctx, "SELECT id FROM users WHERE username=$1 AND password=$2", username, password).Scan(&id)
	return id, err
}

func (p *pgStore) CreateMessage(ctx context.Context, userID int64, content string) (Message, error) {
	var m Message
	err := p.db.QueryRow(ctx, "INSERT INTO messages (user_id, content) VALUES ($1, $2) RETURNING id, created_at", userID, content).Scan(&m.ID, &m.CreatedAt)
	m.UserID = userID
	m.Content = content
	return m, err
}

func (p *pgStore) ListMessages(ctx context.Context, userID int64, limit int) ([]Message, error) {
	rows, err := p.db.Query(ctx, "SELECT id, user_id, content, created_at FROM messages WHERE user_id=$1 ORDER BY created_at DESC LIMIT $2", userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var msgs []Message
	for rows.Next() {
		var m Message
		if err := rows.Scan(&m.ID, &m.UserID, &m.Content, &m.CreatedAt); err != nil {
			return nil, err
		}
		msgs = append(msgs, m)
	}
	return msgs, nil
}

func (p *pgStore) Close() {
	p.db.Close()
}

type memoryStore struct {
	users    []User
	messages []Message
}

func newMemoryStore() *memoryStore {
	return &memoryStore{}
}

func (m *memoryStore) CreateUser(ctx context.Context, username, password string) (int64, error) {
	id := int64(len(m.users) + 1)
	m.users = append(m.users, User{ID: id, Username: username, Password: password})
	return id, nil
}

func (m *memoryStore) GetUserByCredentials(ctx context.Context, username, password string) (int64, error) {
	for _, u := range m.users {
		if u.Username == username && u.Password == password {
			return u.ID, nil
		}
	}
	return 0, fmt.Errorf("invalid credentials")
}

func (m *memoryStore) CreateMessage(ctx context.Context, userID int64, content string) (Message, error) {
	msg := Message{ID: int64(len(m.messages) + 1), UserID: userID, Content: content, CreatedAt: time.Now()}
	m.messages = append(m.messages, msg)
	return msg, nil
}

func (m *memoryStore) ListMessages(ctx context.Context, userID int64, limit int) ([]Message, error) {
	res := []Message{}
	for i := len(m.messages) - 1; i >= 0 && len(res) < limit; i-- {
		if m.messages[i].UserID == userID {
			res = append(res, m.messages[i])
		}
	}
	return res, nil
}

func (m *memoryStore) Close() {}
