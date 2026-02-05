package discordapi

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
)

// TODOs
// - add emoji react limit per channel (1? 2?)
// - error handling for any failures
// - error message for max emoji limit
// - command for removing channel-emoji setup

func (d Discord) ProcessAutoEmojiReactCommand(interaction *discordgo.InteractionCreate) {
	data := interaction.ApplicationCommandData()
	session := d.Session

	if len(data.Options) == 0 {
		session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "sorry, but i cant seem to find your options :(",
			},
		})
	}

	channel := data.GetOption("channel").ChannelValue(session)
	emoji := data.GetOption("emoji").StringValue()

	// d.Session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
	// 	Type: discordgo.InteractionResponseChannelMessageWithSource,
	// 	Data: &discordgo.InteractionResponseData{
	// 		Content: fmt.Sprintf("Added %s reaction to channel %s", emoji, channel.Name),
	// 	},
	// })

	var emojiID string

	if strings.Contains(emoji, "<") {
		emojiID = strings.Trim(emoji, "<>")
	} else {
		emojiID = emoji
	}

	if d.Database.ChannelHasHandler(channel.ID, "MessageCreate", "autoEmojiReact") {
		// only add reaction without handler
		// query saved message? handler? here to pass to AddChannelEmoji
		// d.AddChannelEmoji(interaction, channel, )
	} else {
		// add emoji to channel listener
		d.AddChannelEmojiHandler(interaction, channel, emojiID, emoji)
	}

}

func (d Discord) AddChannelEmojiHandler(interaction *discordgo.InteractionCreate, channel *discordgo.Channel, emojiID string, emoji string) {
	session := d.Session
	// for final ver - check that it doesnt respond to bots AND mods

	// handlerFunc :=
	session.AddHandler(func(s *discordgo.Session, m *discordgo.MessageCreate) {
		message := m.Message
		if message.ChannelID == channel.ID {
			d.AddChannelEmoji(interaction, message, channel, emojiID, emoji)
		}
	})
}

func (d Discord) AddChannelEmoji(interaction *discordgo.InteractionCreate, message *discordgo.Message, channel *discordgo.Channel, emojiID string, emoji string) {
	session := d.Session
	var reactionErr error
	reactionErr = session.MessageReactionAdd(channel.ID, message.ID, emojiID)
	savedEmoji, emojiErr := d.Database.SaveEmojiChannelReaction(channel.ID, emojiID)

	if reactionErr != nil && emojiErr != nil {
		fmt.Printf("%+v", reactionErr)
		fmt.Printf("%+v", emojiErr)
		// session.ChannelMessageSend(interaction.ChannelID, fmt.Sprintf("%+v", reactionErr))
		d.Session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("Failed to add %s reaction to channel %s", emoji, channel.Name),
			},
		})

	} else {
		d.Session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("Added %s reaction to channel %s", emoji, channel.Name),
			},
		})
	}
}

// TODOS 01/02/26
// - save command to table in db
// - save handlers to channel_command_handlers in AddChannelEmojiHandler
// - figure out how to save handlers as string
// -
