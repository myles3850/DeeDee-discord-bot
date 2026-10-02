package discord

import (
	"fmt"
	"log/slog"
	"slices"
	"time"

	"github.com/bwmarrin/discordgo"
)

func (d *Discord) timeoutBotRole(m *discordgo.GuildMemberUpdate) {
	timeoutUntil := time.Now().Add(botRoleTimeoutDuration)

	memberHasBotRole := slices.Contains(m.Roles, botSpamRoleID)
	alreadyTimedOut := m.CommunicationDisabledUntil != nil && m.CommunicationDisabledUntil.After(time.Now())

	if memberHasBotRole && !alreadyTimedOut {
		d.timeoutMember(m.User, &timeoutUntil, "Chose the Bot spam role")
	}
}

func (d *Discord) timeoutMember(user *discordgo.User, timeoutUntil *time.Time, reason string) {
	slog.Info("timing out member", "userId", user.ID, "reason", reason)

	if err := d.Session.GuildMemberTimeout(d.GuildID, user.ID, timeoutUntil); err != nil {
		slog.Error("timing out member", "userId", user.ID, "error", err)
	}

	embed := &discordgo.MessageEmbed{
		Title: "A user has been timed out",
		Color: 0xFFDE21,
		Fields: []*discordgo.MessageEmbedField{
			{Name: "User", Value: user.DisplayName(), Inline: true},
			{Name: "Reason", Value: reason, Inline: true},
			{Name: "Time", Value: fmt.Sprintf("<t:%d:F>", time.Now().Unix()), Inline: true},
		},
	}

	if _, err := d.Session.ChannelMessageSendEmbed(modLogChannelID, embed); err != nil {
		slog.Error("sending timeout embed", "error", err)
	}
}
