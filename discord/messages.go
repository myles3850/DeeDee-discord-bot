package discord

import (
	"fmt"
	"log/slog"

	"github.com/bwmarrin/discordgo"
)

func (d *Discord) saveMessageToDb(m *discordgo.MessageCreate) {
	if m.Author == nil || m.Author.Bot {
		return
	}

	userID, err := d.Database.SaveUser(m.Author.ID, m.Author.Username)
	if err != nil {
		slog.Error("OnMessageCreate: error saving user", "username", m.Author.Username, "error", err)
		return
	}

	if _, err := d.Database.SaveMessage(m.ID, m.ChannelID, userID, m.Content, m.Timestamp); err != nil {
		slog.Error("OnMessageCreate: error saving message", "messageId", m.ID, "error", err)
	}
}

func (d *Discord) reactToIntroMessage(m *discordgo.MessageCreate) {
	if err := d.Session.MessageReactionAdd(m.ChannelID, m.Message.ID, "birbwave:1492652393433796890"); err != nil {
		slog.Error("reacting to intro message", "messageId", m.Message.ID, "error", err)
	}
}

func (d *Discord) handleDeletedMessage(m *discordgo.MessageDelete) {
	channelID := m.ChannelID
	content := m.Content
	createdAt := m.Timestamp
	var author string

	if m.Author != nil {
		author = fmt.Sprintf("<@%s>", m.Author.ID)
	}

	var channelParentID string
	channel, err := d.Session.Channel(channelID)
	if err == nil {
		channelParentID = channel.ParentID
	}

	// ignore anything happening in modmail so as to not get flooded
	if channelParentID == modmailCategoryID {
		return
	}

	// use db as fallback for any missing fields
	dbContent, dbUsername, dbChannelID, dbCreatedAt, err := d.Database.GetMessageWithAuthor(m.ID)
	if err != nil {
		slog.Error("handleDeletedMessage: db lookup failed", "messageId", m.ID, "error", err)
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

	category := "None"
	if channelParentID != "" {
		if parent, err := d.Session.Channel(channelParentID); err != nil {
			slog.Error("handleDeletedMessage: failed to fetch parent channel", "channelId", channelParentID, "error", err)
		} else {
			category = parent.Name
		}
	}

	embed := &discordgo.MessageEmbed{
		Title: "Message Deleted",
		Color: 0xED4245,
		Fields: []*discordgo.MessageEmbedField{
			{Name: "Author", Value: author, Inline: true},
			{Name: "Channel", Value: fmt.Sprintf("<#%s>", channelID), Inline: true},
			{Name: "Category", Value: category, Inline: true},
			{Name: "Sent", Value: sentAt, Inline: true},
			{Name: "Content", Value: content, Inline: false},
		},
		Footer: &discordgo.MessageEmbedFooter{
			Text: fmt.Sprintf("Message ID: %s", m.ID),
		},
	}

	if _, err := d.Session.ChannelMessageSendEmbed(modLogChannelID, embed); err != nil {
		slog.Error("handleDeletedMessage: failed to send embed", "error", err)
	}
}

func (d *Discord) saveModifiedMessage(m *discordgo.MessageUpdate) {
	if err := d.Database.UpdateMessageContent(m.ID, m.Content); err != nil {
		slog.Error("saving modified message", "messageId", m.ID, "error", err)
	}
}
