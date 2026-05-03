package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"time"

	_ "github.com/lib/pq"
)

type Db struct {
	Session *sql.DB
}

type User struct {
	id              int
	DiscordID       string
	DiscordUsername string
}

type Message struct {
	id               int
	DiscordMessageID string
	ChannelID        string
	AuthorID         int
	Content          string
	CreatedAt        time.Time
	EditHistory      []EditEntry
}

type EditEntry struct {
	Content   string    `json:"content"`
	ChangedAt time.Time `json:"changed_at"`
}

type Reaction struct {
	id        int
	MessageID int
	Emoji     string
	ReactorID int
}

type CompletedChannel struct {
	id          int
	ChannelID   string
	CompletedAt time.Time
}

func Setup() *Db {
	var db Db
	host := os.Getenv("DATABASE_HOST")
	port := os.Getenv("DATABASE_PORT")
	user := os.Getenv("DATABASE_USER")
	pass := os.Getenv("DATABASE_PASS")
	database := os.Getenv("DATABASE_DB")

	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s "+
		"password=%s dbname=%s sslmode=disable",
		host, port, user, pass, database)

	d, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		panic(err)
	}
	db.Session = d

	err = db.Session.Ping()
	if err != nil {
		panic(err)
	}
	fmt.Printf("database connected and online: \n %+v \n", db.Session.Stats())

	if err := HandleMigrations(db.Session); err != nil {
		panic(err)
	}

	return &db
}

func (d *Db) GetUser(userId int) User {
	var user User
	sqlQuery := "SELECT * FROM users WHERE discord_id = $1"
	err := d.Session.QueryRow(sqlQuery, userId).Scan(
		&user.id,
		&user.DiscordID,
		&user.DiscordUsername,
	)
	switch err {
	case sql.ErrNoRows:
		println(err.Error())
		return User{}
	default:
		return user
	}
}

func (d *Db) SaveUser(u *User) (int, error) {
	var id int
	sqlQuery := `INSERT INTO users (discord_id, discord_user)
VALUES ($1, $2)
ON CONFLICT (discord_id) DO UPDATE SET discord_user = EXCLUDED.discord_user
RETURNING id;`

	err := d.Session.QueryRow(sqlQuery, u.DiscordID, u.DiscordUsername).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (d *Db) SaveMessage(m *Message) (int, error) {
	var id int
	sqlQuery := `INSERT INTO messages (discord_message_id, channel_id, author_id, content, created_at)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (discord_message_id)
DO UPDATE SET channel_id = EXCLUDED.channel_id,
			  author_id = EXCLUDED.author_id,
			  content = EXCLUDED.content,
			  created_at = EXCLUDED.created_at
RETURNING id;`

	err := d.Session.QueryRow(sqlQuery, m.DiscordMessageID, m.ChannelID, m.AuthorID, m.Content, m.CreatedAt).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (d *Db) GetMessageWithAuthor(discordMessageID string) (content, username, channelID string, createdAt time.Time, err error) {
	sqlQuery := `
SELECT m.content, u.discord_user, m.channel_id, m.created_at
	FROM messages m
	JOIN users u ON m.author_id = u.id
	WHERE m.discord_message_id = $1`

	err = d.Session.QueryRow(sqlQuery, discordMessageID).Scan(&content, &username, &channelID, &createdAt)
	return
}

func (d *Db) SaveMessageWithAuthor(m *Message, author *User) (int, int, error) {
	uid, err := d.SaveUser(author)
	if err != nil {
		return 0, 0, err
	}
	m.AuthorID = uid
	mid, err := d.SaveMessage(m)
	if err != nil {
		return 0, uid, err
	}
	return mid, uid, nil
}

func (d *Db) SaveReaction(r *Reaction) {
	sqlQuery := `INSERT INTO reactions (message_id, emoji, reactor_id)
		VALUES ($1, $2, $3)
		ON CONFLICT (message_id, emoji, reactor_id)
		DO UPDATE SET emoji = EXCLUDED.emoji, reactor_id = EXCLUDED.reactor_id;`

	_, err := d.Session.Exec(sqlQuery, r.MessageID, r.Emoji, r.ReactorID)
	if err != nil {
		fmt.Printf("unable to save reaction for message %d: %+v \n", r.MessageID, err.Error())
	}
}

func (d *Db) MarkChannelCompleted(channelId string) error {
	sqlQuery := `INSERT INTO completed_channels (channel_id, completed_at)
VALUES ($1, $2)
ON CONFLICT (channel_id) DO UPDATE SET completed_at = EXCLUDED.completed_at;`

	_, err := d.Session.Exec(sqlQuery, channelId, time.Now())
	if err != nil {
		return err
	}
	return nil
}

func (d *Db) IsChannelCompleted(channelId string) (bool, error) {
	var id int
	sqlQuery := "SELECT id FROM completed_channels WHERE channel_id = $1"
	err := d.Session.QueryRow(sqlQuery, channelId).Scan(&id)
	switch err {
	case sql.ErrNoRows:
		return false, nil
	case nil:
		return true, nil
	default:
		return false, err
	}
}

func (d *Db) SaveChannelName(channelId string, channelName string) {
	sqlQuery := `INSERT INTO channel_name (name, discord_id)
VALUES ($1, $2)
ON CONFLICT (discord_id) DO UPDATE SET discord_id = EXCLUDED.discord_id;`

	_, err := d.Session.Exec(sqlQuery, channelName, channelId)
	if err != nil {
		fmt.Printf("unable to save channel %s name %s: %+v \n", channelId, channelName, err.Error())
	}
}

func (d *Db) UpdateMessageContent(messageId string, newMessage string) error {
	getQuery := "SELECT id, edit_history, content, discord_message_id FROM messages WHERE discord_message_id = $1"
	putQuery := "UPDATE messages SET content = $1, edit_history = $2 WHERE id = $3"
	updateTime := time.Now()

	var message Message
	var historyBytes []byte
	err := d.Session.QueryRow(getQuery, messageId).Scan(&message.id, &historyBytes, &message.Content, &message.DiscordMessageID)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("message not found")
		}
		return err
	}

	if historyBytes != nil {
		if err := json.Unmarshal(historyBytes, &message.EditHistory); err != nil {
			return fmt.Errorf("failed to parse edit_history: %w", err)
		}
	}

	editHistory := append(message.EditHistory, EditEntry{Content: message.Content, ChangedAt: updateTime})

	historyJSON, err := json.Marshal(editHistory)
	if err != nil {
		return fmt.Errorf("failed to marshal edit_history: %w", err)
	}

	_, err = d.Session.Exec(putQuery, newMessage, historyJSON, message.id)
	if err != nil {
		fmt.Printf("unable to save updated message: '%s' messageId: %s error: %+v \n", newMessage, message.DiscordMessageID, err.Error())
		return err
	}

	return nil
}
