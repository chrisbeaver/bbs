package messages

import (
	"fmt"
	"strconv"
	"strings"

	"bbs/internal/database"
	"bbs/internal/menu"
	"bbs/internal/modules"
	"bbs/internal/modules/base"
)

// TopicOption represents a message board topic option
type TopicOption struct {
	topic       *database.Topic
	index       int
	colorScheme menu.ColorScheme
}

// NewTopicOption creates a new topic option
func NewTopicOption(topic *database.Topic, index int, colorScheme menu.ColorScheme) base.MenuOption {
	return &TopicOption{
		topic:       topic,
		index:       index,
		colorScheme: colorScheme,
	}
}

// GetID implements MenuOption interface
func (t *TopicOption) GetID() string {
	return strconv.Itoa(t.topic.ID)
}

// GetTitle implements MenuOption interface
func (t *TopicOption) GetTitle() string {
	return t.topic.Name
}

// GetDescription implements MenuOption interface
func (t *TopicOption) GetDescription() string {
	return t.topic.Description
}

// Execute implements MenuOption interface
func (t *TopicOption) Execute(writer modules.Writer, keyReader modules.KeyReader, db *database.DB, colorScheme menu.ColorScheme) bool {
	t.showTopicPosts(writer, keyReader, *t.topic, db, colorScheme)
	return true
}

func (t *TopicOption) showTopicPosts(writer modules.Writer, keyReader modules.KeyReader, topic database.Topic, db *database.DB, colorScheme menu.ColorScheme) {
	for {
		// Clear screen and show header
		writer.Write([]byte("\033[2J\033[H"))
		writer.Write([]byte(colorScheme.Colorize(fmt.Sprintf("=== %s ===", strings.ToUpper(topic.Name)), "header") + "\n\n"))

		// Get posts for this topic
		posts, err := db.GetPostsByTopic(topic.ID)
		if err != nil {
			writer.Write([]byte(colorScheme.Colorize("Error loading posts: "+err.Error(), "error") + "\n"))
			writer.Write([]byte(colorScheme.Colorize("Press any key to return...", "prompt")))
			keyReader.ReadKey()
			return
		}

		if len(posts) == 0 {
			writer.Write([]byte(colorScheme.Colorize("No posts in this topic yet.\n", "text")))
			writer.Write([]byte(colorScheme.Colorize("N) Create New Post  Q) Back to Topics\n", "prompt")))
		} else {
			writer.Write([]byte(colorScheme.Colorize("Posts:\n", "text")))
			for i, post := range posts {
				postInfo := fmt.Sprintf("%d) %s by %s (%s)",
					i+1,
					post.Subject,
					post.Username,
					post.CreatedAt.Format("Jan 2, 2006"))
				writer.Write([]byte(colorScheme.Colorize(postInfo, "text") + "\n"))
			}
			writer.Write([]byte(colorScheme.Colorize("\nEnter post number to view, N) New Post, Q) Back: ", "prompt")))
		}

		// Read user input
		input, err := keyReader.ReadKey()
		if err != nil {
			return
		}

		inputStr := strings.TrimSpace(strings.ToLower(string(input)))

		if inputStr == "q" || inputStr == "quit" {
			return
		}

		if inputStr == "n" || inputStr == "new" {
			t.showNewPostForm(writer, keyReader, topic, db, colorScheme)
			continue
		}

		// Try to parse as post number
		if postNum, err := strconv.Atoi(inputStr); err == nil && postNum > 0 && postNum <= len(posts) {
			t.showPost(writer, keyReader, topic, posts[postNum-1], db, colorScheme)
		}
	}
}

func (t *TopicOption) showPost(writer modules.Writer, keyReader modules.KeyReader, topic database.Topic, post database.Post, db *database.DB, colorScheme menu.ColorScheme) {
	// Clear screen and show post
	writer.Write([]byte("\033[2J\033[H"))
	writer.Write([]byte(colorScheme.Colorize(fmt.Sprintf("=== %s: %s ===", topic.Name, post.Subject), "header") + "\n\n"))

	postHeader := fmt.Sprintf("From: %s\nDate: %s\n\n",
		post.Username,
		post.CreatedAt.Format("January 2, 2006 at 3:04 PM"))
	writer.Write([]byte(colorScheme.Colorize(postHeader, "text")))

	// Show post content
	writer.Write([]byte(colorScheme.Colorize(post.Body, "text") + "\n\n"))

	// Get replies
	replies, err := db.GetRepliesByPost(post.ID)
	if err == nil && len(replies) > 0 {
		writer.Write([]byte(colorScheme.Colorize("--- Replies ---\n", "header")))
		for _, reply := range replies {
			replyHeader := fmt.Sprintf("\nFrom: %s (%s)\n",
				reply.Username,
				reply.CreatedAt.Format("Jan 2, 2006 3:04 PM"))
			writer.Write([]byte(colorScheme.Colorize(replyHeader, "text")))
			writer.Write([]byte(colorScheme.Colorize(reply.Body, "text") + "\n"))
		}
	}

	writer.Write([]byte(colorScheme.Colorize("\nR) Reply  Q) Back to topic: ", "prompt")))

	// Read user input
	input, err := keyReader.ReadKey()
	if err != nil {
		return
	}

	inputStr := strings.TrimSpace(strings.ToLower(string(input)))

	if inputStr == "r" || inputStr == "reply" {
		t.showReplyForm(writer, keyReader, topic, post, db, colorScheme)
	}
}

func (t *TopicOption) showNewPostForm(writer modules.Writer, keyReader modules.KeyReader, topic database.Topic, db *database.DB, colorScheme menu.ColorScheme) {
	writer.Write([]byte("\033[2J\033[H"))
	writer.Write([]byte(colorScheme.Colorize(fmt.Sprintf("=== New Post in %s ===", topic.Name), "header") + "\n\n"))
	writer.Write([]byte(colorScheme.Colorize("This feature is coming soon...", "text") + "\n"))
	writer.Write([]byte(colorScheme.Colorize("Press any key to return...", "prompt")))
	keyReader.ReadKey()
}

func (t *TopicOption) showReplyForm(writer modules.Writer, keyReader modules.KeyReader, topic database.Topic, post database.Post, db *database.DB, colorScheme menu.ColorScheme) {
	writer.Write([]byte("\033[2J\033[H"))
	writer.Write([]byte(colorScheme.Colorize(fmt.Sprintf("=== Reply to: %s ===", post.Subject), "header") + "\n\n"))
	writer.Write([]byte(colorScheme.Colorize("This feature is coming soon...", "text") + "\n"))
	writer.Write([]byte(colorScheme.Colorize("Press any key to return...", "prompt")))
	keyReader.ReadKey()
}
