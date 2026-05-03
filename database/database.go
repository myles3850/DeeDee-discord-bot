package database

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"

	_ "github.com/lib/pq"
	"github.com/sqlc-dev/pqtype"

	sqlcdb "choccobear.tech/emojiBot/database/sqlc"
)

type Db struct {
	Session *sql.DB
	Queries *sqlcdb.Queries
}

type EditEntry struct {
	Content   string    `json:"content"`
	ChangedAt time.Time `json:"changed_at"`
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
	db.Queries = sqlcdb.New(d)

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

func (d *Db) SaveUser(discordID string, username string) (int32, error) {
	id, err := strconv.ParseInt(discordID, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid discord ID %s: %w", discordID, err)
	}
	return d.Queries.SaveUser(context.Background(), sqlcdb.SaveUserParams{
		DiscordID:   id,
		DiscordUser: username,
	})
}

func (d *Db) SaveMessage(discordMessageID, channelID string, authorID int32, content string, createdAt time.Time) (int32, error) {
	return d.Queries.SaveMessage(context.Background(), sqlcdb.SaveMessageParams{
		DiscordMessageID: discordMessageID,
		ChannelID:        channelID,
		AuthorID:         authorID,
		Content:          content,
		CreatedAt:        sql.NullTime{Time: createdAt, Valid: !createdAt.IsZero()},
	})
}

func (d *Db) GetMessageWithAuthor(discordMessageID string) (content, username, channelID string, createdAt time.Time, err error) {
	msg, err := d.Queries.GetMessageWithAuthor(context.Background(), discordMessageID)
	if err != nil {
		return
	}
	content = msg.Content
	username = msg.DiscordUser
	channelID = msg.ChannelID
	if msg.CreatedAt.Valid {
		createdAt = msg.CreatedAt.Time
	}
	return
}

func (d *Db) SaveReaction(messageID, reactorID int32, emoji string) error {
	return d.Queries.SaveReaction(context.Background(), sqlcdb.SaveReactionParams{
		MessageID: messageID,
		Emoji:     emoji,
		ReactorID: reactorID,
	})
}

func (d *Db) UpdateMessageContent(messageID string, newMessage string) error {
	ctx := context.Background()

	msg, err := d.Queries.GetMessageForEdit(ctx, messageID)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("message not found")
		}
		return err
	}

	var history []EditEntry
	if msg.EditHistory.Valid {
		if err := json.Unmarshal(msg.EditHistory.RawMessage, &history); err != nil {
			return fmt.Errorf("failed to parse edit_history: %w", err)
		}
	}

	history = append(history, EditEntry{Content: msg.Content, ChangedAt: time.Now()})

	historyJSON, err := json.Marshal(history)
	if err != nil {
		return fmt.Errorf("failed to marshal edit_history: %w", err)
	}

	return d.Queries.UpdateMessageContent(ctx, sqlcdb.UpdateMessageContentParams{
		Content:     newMessage,
		EditHistory: pqtype.NullRawMessage{RawMessage: historyJSON, Valid: true},
		ID:          msg.ID,
	})
}

func (d *Db) MarkChannelCompleted(channelId string) error {
	sqlQuery := `INSERT INTO completed_channels (channel_id, completed_at)
VALUES ($1, $2)
ON CONFLICT (channel_id) DO UPDATE SET completed_at = EXCLUDED.completed_at;`

	_, err := d.Session.Exec(sqlQuery, channelId, time.Now())
	return err
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
