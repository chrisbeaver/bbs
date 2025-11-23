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

// Message board structures
type Topic struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	IsActive    bool      `json:"is_active"`
}

type Post struct {
	ID        int       `json:"id"`
	TopicID   int       `json:"topic_id"`
	UserID    int       `json:"user_id"`
	Username  string    `json:"username"`
	Subject   string    `json:"subject"`
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"created_at"`
}

type Reply struct {
	ID        int       `json:"id"`
	PostID    int       `json:"post_id"`
	UserID    int       `json:"user_id"`
	Username  string    `json:"username"`
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"created_at"`
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
		`CREATE TABLE IF NOT EXISTS topics (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT UNIQUE NOT NULL,
			description TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			is_active BOOLEAN DEFAULT 1
		)`,
		`CREATE TABLE IF NOT EXISTS posts (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			topic_id INTEGER NOT NULL,
			user_id INTEGER NOT NULL,
			username TEXT NOT NULL,
			subject TEXT NOT NULL,
			body TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (topic_id) REFERENCES topics(id),
			FOREIGN KEY (user_id) REFERENCES users(id)
		)`,
		`CREATE TABLE IF NOT EXISTS replies (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			post_id INTEGER NOT NULL,
			user_id INTEGER NOT NULL,
			username TEXT NOT NULL,
			body TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (post_id) REFERENCES posts(id),
			FOREIGN KEY (user_id) REFERENCES users(id)
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

// GetAllUsers retrieves all users (for sysop management)
func (db *DB) GetAllUsers(limit int) ([]User, error) {
	query := `SELECT id, username, password, real_name, email, access_level, 
			  last_call, total_calls, created_at, is_active 
			  FROM users ORDER BY username LIMIT ?`

	rows, err := db.conn.Query(query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var user User
		err := rows.Scan(&user.ID, &user.Username, &user.Password, &user.RealName,
			&user.Email, &user.AccessLevel, &user.LastCall, &user.TotalCalls,
			&user.CreatedAt, &user.IsActive)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	return users, nil
}

// GetUserByID retrieves a single user by ID
func (db *DB) GetUserByID(id int) (*User, error) {
	user := &User{}
	query := `SELECT id, username, password, real_name, email, access_level, 
			  last_call, total_calls, created_at, is_active 
			  FROM users WHERE id = ?`

	err := db.conn.QueryRow(query, id).Scan(
		&user.ID, &user.Username, &user.Password, &user.RealName,
		&user.Email, &user.AccessLevel, &user.LastCall, &user.TotalCalls,
		&user.CreatedAt, &user.IsActive,
	)

	if err != nil {
		return nil, err
	}

	return user, nil
}

// UpdateUser updates user information
func (db *DB) UpdateUser(id int, username, password, realName, email string, accessLevel int, isActive bool) error {
	query := `UPDATE users SET username = ?, password = ?, real_name = ?, 
			  email = ?, access_level = ?, is_active = ? WHERE id = ?`
	_, err := db.conn.Exec(query, username, password, realName, email, accessLevel, isActive, id)
	return err
}

// DeleteUser deletes a user by ID
func (db *DB) DeleteUser(id int) error {
	query := `DELETE FROM users WHERE id = ?`
	_, err := db.conn.Exec(query, id)
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

// Message Board Methods

// Topic management
func (db *DB) GetTopics() ([]Topic, error) {
	query := `SELECT id, name, description, created_at, is_active FROM topics WHERE is_active = 1 ORDER BY name`

	rows, err := db.conn.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var topics []Topic
	for rows.Next() {
		var topic Topic
		err := rows.Scan(&topic.ID, &topic.Name, &topic.Description, &topic.CreatedAt, &topic.IsActive)
		if err != nil {
			return nil, err
		}
		topics = append(topics, topic)
	}

	return topics, nil
}

func (db *DB) GetTopic(id int) (*Topic, error) {
	topic := &Topic{}
	query := `SELECT id, name, description, created_at, is_active FROM topics WHERE id = ?`

	err := db.conn.QueryRow(query, id).Scan(&topic.ID, &topic.Name, &topic.Description,
		&topic.CreatedAt, &topic.IsActive)
	if err != nil {
		return nil, err
	}

	return topic, nil
}

func (db *DB) CreateTopic(name, description string) error {
	query := `INSERT INTO topics (name, description, created_at, is_active) VALUES (?, ?, ?, 1)`
	_, err := db.conn.Exec(query, name, description, time.Now())
	return err
}

// Post management
func (db *DB) GetPostsByTopic(topicID int) ([]Post, error) {
	query := `SELECT id, topic_id, user_id, username, subject, body, created_at 
			  FROM posts WHERE topic_id = ? ORDER BY created_at DESC`

	rows, err := db.conn.Query(query, topicID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []Post
	for rows.Next() {
		var post Post
		err := rows.Scan(&post.ID, &post.TopicID, &post.UserID, &post.Username,
			&post.Subject, &post.Body, &post.CreatedAt)
		if err != nil {
			return nil, err
		}
		posts = append(posts, post)
	}

	return posts, nil
}

func (db *DB) GetPost(id int) (*Post, error) {
	post := &Post{}
	query := `SELECT id, topic_id, user_id, username, subject, body, created_at FROM posts WHERE id = ?`

	err := db.conn.QueryRow(query, id).Scan(&post.ID, &post.TopicID, &post.UserID,
		&post.Username, &post.Subject, &post.Body, &post.CreatedAt)
	if err != nil {
		return nil, err
	}

	return post, nil
}

func (db *DB) CreatePost(topicID, userID int, username, subject, body string) error {
	query := `INSERT INTO posts (topic_id, user_id, username, subject, body, created_at) 
			  VALUES (?, ?, ?, ?, ?, ?)`
	_, err := db.conn.Exec(query, topicID, userID, username, subject, body, time.Now())
	return err
}

// Reply management
func (db *DB) GetRepliesByPost(postID int) ([]Reply, error) {
	query := `SELECT id, post_id, user_id, username, body, created_at 
			  FROM replies WHERE post_id = ? ORDER BY created_at ASC`

	rows, err := db.conn.Query(query, postID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var replies []Reply
	for rows.Next() {
		var reply Reply
		err := rows.Scan(&reply.ID, &reply.PostID, &reply.UserID, &reply.Username,
			&reply.Body, &reply.CreatedAt)
		if err != nil {
			return nil, err
		}
		replies = append(replies, reply)
	}

	return replies, nil
}

func (db *DB) CreateReply(postID, userID int, username, body string) error {
	query := `INSERT INTO replies (post_id, user_id, username, body, created_at) 
			  VALUES (?, ?, ?, ?, ?)`
	_, err := db.conn.Exec(query, postID, userID, username, body, time.Now())
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
