package discordapi

import (
	"fmt"

	"choccobear.tech/emojiBot/database"
	"github.com/bwmarrin/discordgo"
)

func (d *Discord) handleDeletedMessage(m *discordgo.MessageDelete) {
	//? we should find a place for these, maybe db but for now its living here
	var botChannel = "1483226520951455750"

	channelID := m.ChannelID
	content := m.Content
	createdAt := m.Timestamp
	var author string

	if m.Author != nil {
		author = fmt.Sprintf("<@%s>", m.Author.ID)
	}

	// use db as fallback for any missing fields
	dbContent, dbUsername, dbChannelID, dbCreatedAt, err := d.Database.GetMessageWithAuthor(m.ID)
	if err != nil {
		fmt.Printf("handleDeletedMessage: db lookup failed for %s: %v\n", m.ID, err)
	} else {
		if channelID == "" {
			channelID = dbChannelID
		}
		if content == "" {
			content = dbContent
		}
		if author == "" {
			author = dbUsername
		}
		if createdAt.IsZero() {
			createdAt = dbCreatedAt
		}
	}

	var sentAt string
	if !createdAt.IsZero() {
		sentAt = fmt.Sprintf("<t:%d:F>", createdAt.Unix())
	}

	embed := &discordgo.MessageEmbed{
		Title: "Message Deleted",
		Color: 0xED4245,
		Fields: []*discordgo.MessageEmbedField{
			{Name: "Author", Value: author, Inline: true},
			{Name: "Channel", Value: fmt.Sprintf("<#%s>", channelID), Inline: true},
			{Name: "Sent", Value: sentAt, Inline: true},
			{Name: "Content", Value: content, Inline: false},
		},
		Footer: &discordgo.MessageEmbedFooter{
			Text: fmt.Sprintf("Message ID: %s", m.ID),
		},
	}

	d.Session.ChannelMessageSendEmbed(botChannel, embed)
}

func (d *Discord) saveMessageToDb(m *discordgo.MessageCreate) {
	if m.Author == nil || m.Author.Bot {
		return
	}

	userID, err := d.Database.SaveUser(&database.User{
		DiscordID:       m.Author.ID,
		DiscordUsername: m.Author.Username,
	})
	if err != nil {
		fmt.Printf("OnMessageCreate: error saving user %s: %v\n", m.Author.Username, err)
		return
	}

	_, err = d.Database.SaveMessage(&database.Message{
		DiscordMessageID: m.ID,
		ChannelID:        m.ChannelID,
		AuthorID:         userID,
		Content:          m.Content,
		CreatedAt:        m.Timestamp,
	})
	if err != nil {
		fmt.Printf("OnMessageCreate: error saving message %s: %v\n", m.ID, err)
	}

}

func (d *Discord) reactToIntroMessage(m *discordgo.MessageCreate) {
	d.Session.MessageReactionAdd(m.ChannelID, m.Message.ID, "birbwave:1492652393433796890")
}

func (d *Discord) saveModifiedMessage(m *discordgo.MessageUpdate) {
	_ = d.Database.UpdateMessageContent(m.ID, m.Content)
}
