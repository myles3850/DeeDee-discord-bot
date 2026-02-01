package database

import (
	"database/sql"
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
	CreatedAt        time.Time
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

type ChannelEmoji struct {
	ChannelID int
	EmojiID   string
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
	sqlQuery := `INSERT INTO messages (discord_message_id, channel_id, author_id, created_at)
VALUES ($1, $2, $3, $4)
ON CONFLICT (discord_message_id)
DO UPDATE SET channel_id = EXCLUDED.channel_id,
			  author_id = EXCLUDED.author_id,
			  created_at = EXCLUDED.created_at
RETURNING id;`

	err := d.Session.QueryRow(sqlQuery, m.DiscordMessageID, m.ChannelID, m.AuthorID, m.CreatedAt).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
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

func (d *Db) ChannelHasHandler(channelId string, handlerType string, commandName string) bool {
	var exists bool
	sqlQuery := `
	SELECT EXISTS(
		SELECT 1 FROM channel_command_handlers cch
		INNER JOIN channel_name ON channel_name.id = cch.channel_id
		INNER JOIN commands ON commands.id = cch.command_id
		WHERE channel_name.discord_channeL_id = $1
		AND cch.handler_type = $2
		AND commands.name = $3
	);
	`
	d.Session.QueryRow(sqlQuery, channelId, handlerType, commandName).Scan(&exists)
	return exists
}

func (d *Db) SaveEmojiChannelReaction(channelId string, emojiId string) (string, error) {
	var savedEmojiId string
	sqlQuery := `
	INSERT INTO channels_emojis (discord_channel_id, discord_emoji_id)
	VALUES ($1, $2);
	`

	err := d.Session.QueryRow(sqlQuery, channelId, emojiId).Scan(&savedEmojiId)

	if err != nil {
		return "", err
	}
	return savedEmojiId, nil
}

// cant figure out how to save handler function data type
// more research needed
func (d *Db) SaveChannelCommandHandler(commandName string, discordChannelId string, handlerType string, handlerFunc string) (int, error) {
	var dbCommandId int
	var dbChannelId int

	channelSqlQuery := "SELECT id FROM channel_name WHERE discord_channel_id = $1"
	channelErr := d.Session.QueryRow(channelSqlQuery, discordChannelId).Scan(&dbChannelId)

	commandSqlQuery := "SELECT id FROM commands WHERE name = $1"
	commandErr := d.Session.QueryRow(commandSqlQuery, discordChannelId).Scan(&dbCommandId)

	// only do the next err if prev one doesnt error
	// this is so ugly need refactoring
	switch channelErr {
	case sql.ErrNoRows:
		return 0, channelErr
	default:
		switch commandErr {
		case sql.ErrNoRows:
			return 0, commandErr
		default:
			var id int
			sqlQuery := `
	INSERT INTO channel_command_handlers (command_id, channel_id, handler_type, handler_func)
	VALUES ($1, $2, $3, $4);
	`

			err := d.Session.QueryRow(sqlQuery, dbCommandId, dbChannelId, handlerType, handlerFunc).Scan(&id)

			if err != nil {
				return 0, err
			}
			return id, nil
		}
	}

}
