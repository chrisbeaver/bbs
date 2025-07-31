package database

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type DB struct {
	conn *sql.DB
}

type User struct {
	ID          int        `json:"id"`
	Username    string     `json:"username"`
	Password    string     `json:"password"`
	RealName    string     `json:"real_name"`
	Email       string     `json:"email"`
	AccessLevel int        `json:"access_level"`
	LastCall    *time.Time `json:"last_call"`
	TotalCalls  int        `json:"total_calls"`
	CreatedAt   time.Time  `json:"created_at"`
	IsActive    bool       `json:"is_active"`
}

type Message struct {
	ID        int       `json:"id"`
	FromUser  string    `json:"from_user"`
	ToUser    string    `json:"to_user"`
	Subject   string    `json:"subject"`
	Body      string    `json:"body"`
	Area      string    `json:"area"`
	CreatedAt time.Time `json:"created_at"`
	IsRead    bool      `json:"is_read"`
}

type Bulletin struct {
	ID        int        `json:"id"`
	Title     string     `json:"title"`
	Body      string     `json:"body"`
	Author    string     `json:"author"`
	CreatedAt time.Time  `json:"created_at"`
	ExpiresAt *time.Time `json:"expires_at"`
}

func Initialize(dbPath string) (*DB, error) {
	conn, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	db := &DB{conn: conn}

	if err := db.createTables(); err != nil {
		return nil, fmt.Errorf("failed to create tables: %w", err)
	}

	return db, nil
}

func (db *DB) Close() error {
	return db.conn.Close()
}

func (db *DB) createTables() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			username TEXT UNIQUE NOT NULL,
			password TEXT NOT NULL,
			real_name TEXT,
			email TEXT,
			access_level INTEGER DEFAULT 0,
			last_call DATETIME,
			total_calls INTEGER DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			is_active BOOLEAN DEFAULT 1
		)`,
		`CREATE TABLE IF NOT EXISTS messages (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			from_user TEXT NOT NULL,
			to_user TEXT NOT NULL,
			subject TEXT NOT NULL,
			body TEXT NOT NULL,
			area TEXT DEFAULT 'general',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			is_read BOOLEAN DEFAULT 0
		)`,
		`CREATE TABLE IF NOT EXISTS bulletins (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			title TEXT NOT NULL,
			body TEXT NOT NULL,
			author TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			expires_at DATETIME
		)`,
		`CREATE TABLE IF NOT EXISTS sessions (
			id TEXT PRIMARY KEY,
			username TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			last_activity DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
	}

	for _, query := range queries {
		if _, err := db.conn.Exec(query); err != nil {
			return fmt.Errorf("failed to execute query: %w", err)
		}
	}

	return nil
}

// User management methods
func (db *DB) GetUser(username string) (*User, error) {
	user := &User{}
	query := `SELECT id, username, password, real_name, email, access_level, 
			  last_call, total_calls, created_at, is_active 
			  FROM users WHERE username = ? AND is_active = 1`

	err := db.conn.QueryRow(query, username).Scan(
		&user.ID, &user.Username, &user.Password, &user.RealName,
		&user.Email, &user.AccessLevel, &user.LastCall, &user.TotalCalls,
		&user.CreatedAt, &user.IsActive,
	)

	if err != nil {
		return nil, err
	}

	return user, nil
}

func (db *DB) CreateUser(user *User) error {
	query := `INSERT INTO users (username, password, real_name, email, access_level, created_at)
			  VALUES (?, ?, ?, ?, ?, ?)`

	_, err := db.conn.Exec(query, user.Username, user.Password, user.RealName,
		user.Email, user.AccessLevel, time.Now())

	return err
}

func (db *DB) UpdateUserLastCall(username string) error {
	query := `UPDATE users SET last_call = ?, total_calls = total_calls + 1 WHERE username = ?`
	_, err := db.conn.Exec(query, time.Now(), username)
	return err
}

// Message methods
func (db *DB) GetMessages(toUser string, limit int) ([]Message, error) {
	query := `SELECT id, from_user, to_user, subject, body, area, created_at, is_read
			  FROM messages WHERE to_user = ? ORDER BY created_at DESC LIMIT ?`

	rows, err := db.conn.Query(query, toUser, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []Message
	for rows.Next() {
		var msg Message
		err := rows.Scan(&msg.ID, &msg.FromUser, &msg.ToUser, &msg.Subject,
			&msg.Body, &msg.Area, &msg.CreatedAt, &msg.IsRead)
		if err != nil {
			return nil, err
		}
		messages = append(messages, msg)
	}

	return messages, nil
}

func (db *DB) CreateMessage(msg *Message) error {
	query := `INSERT INTO messages (from_user, to_user, subject, body, area, created_at)
			  VALUES (?, ?, ?, ?, ?, ?)`

	_, err := db.conn.Exec(query, msg.FromUser, msg.ToUser, msg.Subject,
		msg.Body, msg.Area, time.Now())

	return err
}

// Bulletin methods
func (db *DB) GetBulletins(limit int) ([]Bulletin, error) {
	query := `SELECT id, title, body, author, created_at, expires_at
			  FROM bulletins 
			  WHERE expires_at IS NULL OR expires_at > ?
			  ORDER BY created_at DESC LIMIT ?`

	rows, err := db.conn.Query(query, time.Now(), limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var bulletins []Bulletin
	for rows.Next() {
		var bulletin Bulletin
		err := rows.Scan(&bulletin.ID, &bulletin.Title, &bulletin.Body,
			&bulletin.Author, &bulletin.CreatedAt, &bulletin.ExpiresAt)
		if err != nil {
			return nil, err
		}
		bulletins = append(bulletins, bulletin)
	}

	return bulletins, nil
}

func (db *DB) CreateBulletin(bulletin *Bulletin) error {
	query := `INSERT INTO bulletins (title, body, author, created_at)
			  VALUES (?, ?, ?, ?)`

	_, err := db.conn.Exec(query, bulletin.Title, bulletin.Body, bulletin.Author, time.Now())
	return err
}

// UpdateBulletin updates an existing bulletin
func (db *DB) UpdateBulletin(id int, title, body string) error {
	query := `UPDATE bulletins SET title = ?, body = ? WHERE id = ?`
	_, err := db.conn.Exec(query, title, body, id)
	return err
}

// DeleteBulletin deletes a bulletin by ID
func (db *DB) DeleteBulletin(id int) error {
	query := `DELETE FROM bulletins WHERE id = ?`
	_, err := db.conn.Exec(query, id)
	return err
}

// GetBulletinByID retrieves a single bulletin by ID
func (db *DB) GetBulletinByID(id int) (*Bulletin, error) {
	query := `SELECT id, title, body, author, created_at, expires_at
			  FROM bulletins WHERE id = ?`

	bulletin := &Bulletin{}
	err := db.conn.QueryRow(query, id).Scan(
		&bulletin.ID, &bulletin.Title, &bulletin.Body,
		&bulletin.Author, &bulletin.CreatedAt, &bulletin.ExpiresAt)

	if err != nil {
		return nil, err
	}

	return bulletin, nil
}
