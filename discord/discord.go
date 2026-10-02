// Package discord is the Discord bot: gateway connection, event handlers,
// and slash commands.
package discord

import (
	"fmt"
	"os"

	"choccobear.tech/deedee/database"
	"choccobear.tech/deedee/googleplatform"
	"github.com/bwmarrin/discordgo"
)

type Discord struct {
	Session  *discordgo.Session
	Database *database.Db
	// Sheet is not read by any handler yet — it's wired up ahead of an
	// unfinished feature (see googleplatform/README.md), not dead code.
	Sheet   *googleplatform.Sheet
	GuildID string
}

func Setup(db *database.Db, sheet *googleplatform.Sheet) (*Discord, error) {
	token := os.Getenv("DISCORD_BOT_TOKEN")
	guildID := os.Getenv("DISCORD_GUILD_ID")

	session, err := discordgo.New("Bot " + token)
	if err != nil {
		return nil, fmt.Errorf("creating discord session: %w", err)
	}
	session.Identify.Intents = discordgo.IntentsGuilds | discordgo.IntentsGuildMessages | discordgo.IntentsMessageContent | discordgo.IntentsGuildMessageReactions | discordgo.IntentsAll
	session.StateEnabled = true

	return &Discord{Session: session, GuildID: guildID, Database: db, Sheet: sheet}, nil
}
