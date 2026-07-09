package discordapi

import (
	"fmt"
	"slices"
	"time"

	"github.com/bwmarrin/discordgo"
)

func (d *Discord) timeoutBotRole(m *discordgo.GuildMemberUpdate) {
	var role string = "1519435918224785458"
	// timeout length is 7 days
	var timeoutLength time.Time = time.Now().Add(168 * time.Hour)

	memberRoles := m.Roles
	memberHasBotRole := slices.Contains(memberRoles, role)
	alreadyTimedOut := m.CommunicationDisabledUntil != nil && m.CommunicationDisabledUntil.After(time.Now())

	if memberHasBotRole && !alreadyTimedOut {
		d.timeoutMember(m.User, &timeoutLength, "Chose the Bot spam role")
	}

}

func (d *Discord) timeoutMember(user *discordgo.User, timeoutLength *time.Time, reason string) {
	var botChannel = "1483226520951455750"

	fmt.Printf("entering timeout block for %s\n", user.ID)
	err := d.Session.GuildMemberTimeout(d.GuildId, user.ID, timeoutLength)
	if err != nil {
		fmt.Printf("error timing out: %v\n", err.Error())
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

	if _, err := d.Session.ChannelMessageSendEmbed(botChannel, embed); err != nil {
		fmt.Printf("error sending embed: %v\n", err.Error())
	}

}
