package main

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Store abstracts database operations so they can be backed by Postgres or an in-memory implementation for tests.
type Store interface {
	CreateUser(ctx context.Context, username, password string) (User, error)
	GetUserByCredentials(ctx context.Context, username, password string) (User, error)
	CreateMessage(ctx context.Context, userID int64, content string) (Message, error)
	GetFeed(ctx context.Context, userID int64) ([]Message, error)
}

// pgStore implements Store using a Postgres database.
type pgStore struct {
	db *pgxpool.Pool
}

func newPGStore(db *pgxpool.Pool) Store {
	return &pgStore{db: db}
}

func (p *pgStore) CreateUser(ctx context.Context, username, password string) (User, error) {
	var u User
	err := p.db.QueryRow(ctx, "INSERT INTO users (username, password) VALUES ($1, $2) RETURNING id", username, password).Scan(&u.ID)
	if err != nil {
		return u, err
	}
	u.Username = username
	u.Password = password
	return u, nil
}

func (p *pgStore) GetUserByCredentials(ctx context.Context, username, password string) (User, error) {
	var u User
	row := p.db.QueryRow(ctx, "SELECT id, username, password FROM users WHERE username=$1 AND password=$2", username, password)
	if err := row.Scan(&u.ID, &u.Username, &u.Password); err != nil {
		return u, err
	}
	return u, nil
}

func (p *pgStore) CreateMessage(ctx context.Context, userID int64, content string) (Message, error) {
	var m Message
	err := p.db.QueryRow(ctx, "INSERT INTO messages (user_id, content) VALUES ($1, $2) RETURNING id, created_at", userID, content).Scan(&m.ID, &m.CreatedAt)
	if err != nil {
		return m, err
	}
	m.UserID = userID
	m.Content = content
	return m, nil
}

func (p *pgStore) GetFeed(ctx context.Context, userID int64) ([]Message, error) {
	rows, err := p.db.Query(ctx, "SELECT id, user_id, content, created_at FROM messages WHERE user_id=$1 ORDER BY created_at DESC LIMIT 20", userID)
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
