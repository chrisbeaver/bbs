package database

import (
	"time"
)

// BulletinSeed represents a bulletin for seeding
type BulletinSeed struct {
	Title  string
	Body   string
	Author string
}

// UserSeed represents a user for seeding
type UserSeed struct {
	Username    string
	Password    string
	RealName    string
	Email       string
	AccessLevel int
	IsActive    bool
}

// getSeedUsers returns the default users for seeding the database
func getSeedUsers() []UserSeed {
	return []UserSeed{
		{
			Username:    "sysop",
			Password:    "password", // In production, use proper password hashing
			RealName:    "System Operator",
			Email:       "sysop@localhost",
			AccessLevel: 255, // Maximum access level
			IsActive:    true,
		},
		{
			Username:    "test",
			Password:    "test",
			RealName:    "Test User",
			Email:       "test@localhost",
			AccessLevel: 10,
			IsActive:    true,
		},
	}
}

// getSeedBulletins returns the default bulletins for seeding the database
func getSeedBulletins() []BulletinSeed {
	return []BulletinSeed{
		{
			Title: "Welcome to Searchlight BBS!",
			Body: `Welcome to Searchlight BBS - A Classic Experience Reborn!

This is a faithful recreation of the classic Searchlight BBS software,
bringing the nostalgic bulletin board system experience to the modern age.

Features:
• SSH connectivity - Connect from anywhere with an SSH client
• Multi-user support - Chat and message with other users
• Message areas - Public and private messaging systems
• File areas - Upload and download files
• Online games - Classic BBS door games
• Configurable menus - Customizable interface
• Real-time chat - Talk to other users online

Whether you're a veteran sysop or new to the BBS scene, we hope you
enjoy exploring this classic computing experience!

Happy BBSing!`,
			Author: "Sysop",
		},
		{
			Title: "System Information",
			Body: `Technical Details - Searchlight BBS Server

This BBS is built with modern technology while maintaining the classic feel:

Backend Technology:
• Written in Go (Golang) for performance and reliability
• SQLite database for data persistence
• SSH server for secure remote connections
• YAML configuration for easy customization

Architecture:
• Concurrent connections via goroutines
• Modular design for easy feature additions
• Terminal-based interface with ANSI color support
• Real-time session management

Security Features:
• SSH encryption for all connections
• User access levels (0-255)
• Session management and authentication
• Input validation and sanitization

For technical support or questions, contact the Sysop.`,
			Author: "Sysop",
		},
		{
			Title: "New User Guidelines",
			Body: `Guidelines for New Users

Welcome to our community! To ensure everyone has a great experience,
please follow these simple guidelines:

Conduct:
• Be respectful to all users and the Sysop
• No harassment, spam, or inappropriate content
• Keep discussions family-friendly
• Respect others' privacy and personal information

Messages:
• Check your messages regularly
• Reply promptly to time-sensitive messages
• Use clear, descriptive subject lines
• Respect message area topics and purposes

Files:
• Only upload files you have permission to share
• Scan all files for viruses before uploading
• Use descriptive filenames and include descriptions
• Respect copyright and intellectual property

Help:
• Read the help files and documentation
• Ask questions if you need assistance
• Report any problems to the Sysop
• Help other new users when you can

Thank you for being part of our community!`,
			Author: "Sysop",
		},
		{
			Title: "System Maintenance Schedule",
			Body: `Regular Maintenance Schedule

To keep the system running smoothly, routine maintenance is performed
on a regular schedule. Please plan accordingly:

Daily Maintenance:
• 3:00 AM - 3:15 AM EST: Database optimization
• Log rotation and cleanup
• Automatic user statistics updates

Weekly Maintenance:
• Sunday 2:00 AM - 4:00 AM EST: Full system maintenance
• Database backup and verification
• System updates and security patches
• File area cleanup and organization

Monthly Maintenance:
• First Sunday of each month: Extended maintenance window
• Major system updates if needed
• User account cleanup (inactive accounts)
• Archive old messages and bulletins

Emergency Maintenance:
• Unscheduled maintenance may occur for critical issues
• Users will be notified when possible
• System may be temporarily unavailable

During maintenance windows, the system may be unavailable or
operating with limited functionality. We appreciate your patience!

For questions about maintenance, contact the Sysop.`,
			Author: "Sysop",
		},
		{
			Title: "Feature Updates and Roadmap",
			Body: `Recent Updates and Coming Features

We're constantly working to improve your BBS experience!

Recent Updates:
• Implemented navigable bulletin system with arrow key support
• Enhanced color scheme and terminal formatting
• Improved session management and stability
• Added modular architecture for easy feature expansion
• Better error handling and user feedback

Coming Soon:
• Enhanced message system with threaded conversations
• File upload/download areas with descriptions
• Online games and door game support
• Real-time chat and instant messaging
• User profiles and customizable settings
• Forum-style message boards
• File tagging and search capabilities

In Development:
• Web-based interface (optional)
• Mobile app support
• Integration with modern social features
• Advanced sysop tools and administration
• User statistics and activity tracking

Requested Features:
• Multi-node support for larger systems
• External door game integration
• Advanced file management tools
• Custom user themes and colors

Have a feature request? Send a message to the Sysop with your ideas!
We love hearing from our users and community.`,
			Author: "Sysop",
		},
	}
}

// LoadBulletinsFromSeed loads default bulletins into the database
func (db *DB) LoadBulletinsFromSeed() error {
	seedBulletins := getSeedBulletins()

	// Insert bulletins into database
	for _, seedBulletin := range seedBulletins {
		// Check if bulletin already exists (by title)
		exists, err := db.bulletinExists(seedBulletin.Title)
		if err != nil {
			return err
		}

		if !exists {
			bulletin := &Bulletin{
				Title:     seedBulletin.Title,
				Body:      seedBulletin.Body,
				Author:    seedBulletin.Author,
				CreatedAt: time.Now(),
			}

			if err := db.CreateBulletin(bulletin); err != nil {
				return err
			}
		}
	}

	return nil
}

// LoadUsersFromSeed loads default users into the database
func (db *DB) LoadUsersFromSeed() error {
	seedUsers := getSeedUsers()

	// Insert users into database
	for _, seedUser := range seedUsers {
		// Check if user already exists (by username)
		exists, err := db.userExists(seedUser.Username)
		if err != nil {
			return err
		}

		if !exists {
			user := &User{
				Username:    seedUser.Username,
				Password:    seedUser.Password,
				RealName:    seedUser.RealName,
				Email:       seedUser.Email,
				AccessLevel: seedUser.AccessLevel,
				IsActive:    seedUser.IsActive,
				CreatedAt:   time.Now(),
			}

			if err := db.CreateUser(user); err != nil {
				return err
			}
		}
	}

	return nil
}

// userExists checks if a user with the given username already exists
func (db *DB) userExists(username string) (bool, error) {
	query := `SELECT COUNT(*) FROM users WHERE username = ?`
	var count int
	err := db.conn.QueryRow(query, username).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// bulletinExists checks if a bulletin with the given title already exists
func (db *DB) bulletinExists(title string) (bool, error) {
	query := `SELECT COUNT(*) FROM bulletins WHERE title = ?`
	var count int
	err := db.conn.QueryRow(query, title).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
