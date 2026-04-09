package discordapi

import (
	"database/sql"
	"fmt"

	"choccobear.tech/emojiBot/database"
	"github.com/bwmarrin/discordgo"
)

func (d *Discord) handleDeletedMessage(m *discordgo.MessageDelete) {
	var author, content string
	//? we should find a place for these, maybe db but for now its living here
	var botChannel = "1483226520951455750"

	if m.Author != nil {
		author = m.Author.Username
		content = m.Content
	} else {
		var err error
		content, author, err = d.Database.GetMessageWithAuthor(m.ID)
		if err == sql.ErrNoRows {
			return
		}
		if err != nil {
			fmt.Printf("handleDeletedMessage: db lookup failed for %s: %v\n", m.ID, err)
			return
		}
	}

	embed := &discordgo.MessageEmbed{
		Title: "Message Deleted",
		Fields: []*discordgo.MessageEmbedField{
			{Name: "Author", Value: author, Inline: true},
			{Name: "Content", Value: content, Inline: false},
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
