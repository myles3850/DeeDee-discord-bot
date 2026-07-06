package discordapi

import (
	"fmt"
	"os"

	"choccobear.tech/emojiBot/database"
	googleplatform "choccobear.tech/emojiBot/googlePlatform"
	"github.com/bwmarrin/discordgo"
)

type Discord struct {
	Session  *discordgo.Session
	Database *database.Db
	Sheet    *googleplatform.Sheet
	GuildId  string
}

func Setup(db *database.Db, sheet *googleplatform.Sheet) (*Discord, error) {
	token := os.Getenv("DISCORD_BOT_TOKEN")
	guildId := os.Getenv("DISCORD_GUILD_ID")

	discord, err := discordgo.New("Bot " + token)
	discord.Identify.Intents = discordgo.IntentsGuilds | discordgo.IntentsGuildMessages | discordgo.IntentsMessageContent | discordgo.IntentsGuildMessageReactions
	discord.StateEnabled = true

	if err != nil {
		fmt.Println("error creating Discord session: ", err)
		return &Discord{Session: discord, GuildId: guildId}, err
	}
	return &Discord{Session: discord, GuildId: guildId, Database: db, Sheet: sheet}, err
}

func (d *Discord) GetAllEmojis() []*discordgo.Emoji {
	emojis, err := d.Session.GuildEmojis(d.GuildId)

	if err != nil {
		fmt.Println("error getting emojis: ", err)
		return nil
	}
	return emojis
}

func (d *Discord) GetOneEmoji(emojiId string) *discordgo.Emoji {
	emoji, err := d.Session.GuildEmoji(d.GuildId, emojiId)

	if err != nil {
		fmt.Println("error getting emoji: ", err)
		return nil
	}
	return emoji
}

func (d *Discord) GetAllRoles() []*discordgo.Role {
	roles, err := d.Session.GuildRoles(d.GuildId)
	if err != nil {
		fmt.Println("error getting emoji: ", err)
		return nil
	}
	return roles
}

func (d *Discord) EditEmojiRoles(emojiId string, params *discordgo.EmojiParams) error {
	_, err := d.Session.GuildEmojiEdit(d.GuildId, emojiId, params)
	if err != nil {
		fmt.Println("error updating emoji: ", err)
		return err
	}
	return nil
}
