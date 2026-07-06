package discordapi

import (
	"fmt"
	"slices"
	"time"

	"github.com/bwmarrin/discordgo"
)

func (d *Discord) timeoutBotRole(m *discordgo.GuildMemberUpdate) {
	var role string = "1519435918224785458"
	var botChannel = "1483226520951455750"
	// timeout length is 7 days
	var timeoutLength time.Time = time.Now().Add(168 * time.Hour)

	memberRoles := m.Roles
	memberHasBotRole := slices.Contains(memberRoles, role)
	alreadyTimedOut := m.CommunicationDisabledUntil != nil && m.CommunicationDisabledUntil.After(time.Now())
	fmt.Printf("roles: %v \n", memberRoles)
	fmt.Printf("has bot role %b \n", memberHasBotRole)
	fmt.Printf("times out befdore: %b \n\n", alreadyTimedOut)
	
	if memberHasBotRole && !alreadyTimedOut {
		d.Session.GuildMemberTimeout(d.GuildId, m.User.ID, &timeoutLength)
		embed := &discordgo.MessageEmbed{
			Title: "A user has selected the Bot role",
			Color: 0xFFDE21,
			Fields: []*discordgo.MessageEmbedField{
				{Name: "User", Value: m.DisplayName(), Inline: true},
				{Name: "time", Value: fmt.Sprintf("<t:%d:F>", time.Now().Unix()), Inline: true},
			},
		}

		d.Session.ChannelMessageSendEmbed(botChannel, embed)
	}
}
