package discord

import (
	"fmt"
	"log/slog"

	"github.com/bwmarrin/discordgo"
)

func (d *Discord) GetAllEmojis() []*discordgo.Emoji {
	emojis, err := d.Session.GuildEmojis(d.GuildID)
	if err != nil {
		slog.Error("getting emojis", "error", err)
		return nil
	}
	return emojis
}

func (d *Discord) GetAllRoles() []*discordgo.Role {
	roles, err := d.Session.GuildRoles(d.GuildID)
	if err != nil {
		slog.Error("getting roles", "error", err)
		return nil
	}
	return roles
}

func (d *Discord) EditEmojiRoles(emojiID string, params *discordgo.EmojiParams) error {
	if _, err := d.Session.GuildEmojiEdit(d.GuildID, emojiID, params); err != nil {
		return fmt.Errorf("updating emoji %s roles: %w", emojiID, err)
	}
	return nil
}
