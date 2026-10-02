package discord

import (
	"fmt"
	"log/slog"
	"math/rand"
	"strings"

	"github.com/bwmarrin/discordgo"
)

const (
	cmdWheel      = "wheel"
	cmdEightBall  = "eight_ball"
	cmdProcessOld = "process_old_messages"
	cmdShake      = "shake"
)

// respond sends a plain-text interaction response, logging any failure
// instead of dropping it.
func respond(s *discordgo.Session, i *discordgo.Interaction, content string) {
	err := s.InteractionRespond(i, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: content,
		},
	})
	if err != nil {
		slog.Error("responding to interaction", "error", err)
	}
}

func (d *Discord) RegisterCommands() {
	s := d.Session
	appID := s.State.User.ID
	guildID := d.GuildID
	minStringLength := 2
	var defaultMemberPermissions int64 = discordgo.PermissionManageGuild

	commands := []*discordgo.ApplicationCommand{
		{
			Name:        cmdWheel,
			Description: "Give me a selection, and ill pick one for you",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:        "option1",
					Description: "Choice One",
					Type:        discordgo.ApplicationCommandOptionString,
					Required:    true,
					MinLength:   &minStringLength,
				},
				{
					Name:        "option2",
					Description: "Choice Two",
					Type:        discordgo.ApplicationCommandOptionString,
					Required:    true,
					MinLength:   &minStringLength,
				},
				{
					Name:        "option3",
					Description: "Choice Three",
					Type:        discordgo.ApplicationCommandOptionString,
					Required:    false,
					MinLength:   &minStringLength,
				},
				{
					Name:        "option4",
					Description: "Choice Four",
					Type:        discordgo.ApplicationCommandOptionString,
					Required:    false,
					MinLength:   &minStringLength,
				},
				{
					Name:        "option5",
					Description: "Choice Five",
					Type:        discordgo.ApplicationCommandOptionString,
					Required:    false,
					MinLength:   &minStringLength,
				},
				{
					Name:        "option6",
					Description: "Choice Six",
					Type:        discordgo.ApplicationCommandOptionString,
					Required:    false,
					MinLength:   &minStringLength,
				},
			},
		},
		{
			Name:        cmdEightBall,
			Description: "ask DeeDee to shake the mystical 8 ball for you",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:        "question",
					Description: "give me your question so i can find out the answer",
					Type:        discordgo.ApplicationCommandOptionString,
					Required:    true,
					MinLength:   &minStringLength,
				},
			},
		},
		{
			Name:                     cmdProcessOld,
			Description:              "run through all old messages",
			DefaultMemberPermissions: &defaultMemberPermissions,
		},
		{
			Name:        cmdShake,
			Description: "sends special shake emote",
		},
	}

	for _, cmd := range commands {
		_, err := s.ApplicationCommandCreate(appID, guildID, cmd)
		if err != nil {
			slog.Error("registering command", "command", cmd.Name, "error", err)
		} else {
			slog.Info("registered command", "command", cmd.Name)
		}
	}
}

func (d *Discord) processWheelCommand(interaction *discordgo.InteractionCreate) {
	data := interaction.ApplicationCommandData()
	user := interaction.Member.Nick

	if len(data.Options) == 0 {
		respond(d.Session, interaction.Interaction, fmt.Sprintf("sorry, %s, but i cant seem to find your options :(", user))
		return
	}

	options := make([]string, len(data.Options))
	for i, option := range data.Options {
		options[i] = option.StringValue()
	}

	chosen := options[rand.Intn(len(options))]
	respond(d.Session, interaction.Interaction, fmt.Sprintf("Interesting choices... im feeling %s this time.", chosen))
}

func (d *Discord) process8BallCommand(interaction *discordgo.InteractionCreate) {
	ballAnswers := []string{
		"It is certain",
		"It is decidedly so",
		"Without a doubt",
		"Yes definitely",
		"You may rely on it",
		"As the ball sees it, yes",
		"Most likely",
		"Outlook good",
		"Yes",
		"Signs point to yes",
		"Reply hazy, try again",
		"Ask again later",
		"Better not tell you now",
		"Cannot predict now",
		"Concentrate and ask again",
		"Don't count on it",
		"It's reply is no",
		"It's sources say no",
		"Outlook not so good",
		"Very doubtful",
	}

	data := interaction.ApplicationCommandData()
	question := data.Options[0].StringValue()
	user := interaction.Member.Nick

	if len(question) == 0 {
		respond(d.Session, interaction.Interaction, fmt.Sprintf("sorry, %s, but the ball refuses to answer my call :(", user))
		return
	}

	questionOfLife := "what is the answer to life the universe and everything"
	if strings.Contains(question, questionOfLife) {
		respond(d.Session, interaction.Interaction, fmt.Sprintf("I asked the magical 8 ball \"%s\" , and it said **42**.", question))
		return
	}

	selectedAnswer := ballAnswers[rand.Intn(len(ballAnswers))]
	respond(d.Session, interaction.Interaction, fmt.Sprintf("I asked the magical 8 ball \"%s\" , and it said **%s**.", question, selectedAnswer))
}

func (d *Discord) ProcessOldMessages(interaction *discordgo.InteractionCreate) {
	const fetchMessageBatchSize = 100

	respond(d.Session, interaction.Interaction, "getting you the info now...")

	channels, err := d.Session.GuildChannels(d.GuildID)
	if err != nil {
		slog.Error("ProcessOldMessages: listing channels", "error", err)
	}

	for _, channel := range channels {
		channelComplete, _ := d.Database.IsChannelCompleted(channel.ID)
		if err := d.Database.SaveChannelName(channel.ID, channel.Name); err != nil {
			slog.Error("ProcessOldMessages: saving channel name", "channelId", channel.ID, "error", err)
		}
		if channelComplete {
			slog.Info("ProcessOldMessages: skipping already-completed channel", "channel", channel.Name)
			continue
		}

		var lastMessage string
		for {
			slog.Info("ProcessOldMessages: processing channel", "channel", channel.Name)
			messages, err := d.Session.ChannelMessages(channel.ID, fetchMessageBatchSize, lastMessage, "", "")
			if err != nil {
				slog.Error("ProcessOldMessages: fetching messages", "channel", channel.Name, "error", err)
			}
			if len(messages) == 0 {
				break
			}

			for _, message := range messages {
				userID, err := d.Database.SaveUser(message.Author.ID, message.Author.Username)
				if err != nil {
					slog.Error("ProcessOldMessages: saving user", "userId", message.Author.ID, "error", err)
					break
				}
				messageID, err := d.Database.SaveMessage(message.ID, channel.ID, userID, "", message.Timestamp)
				if err != nil {
					slog.Error("ProcessOldMessages: saving message", "messageId", message.ID, "error", err)
					break
				}
				for _, react := range message.Reactions {
					if err := d.Database.SaveReaction(messageID, userID, react.Emoji.Name); err != nil {
						slog.Error("ProcessOldMessages: saving reaction", "messageId", message.ID, "error", err)
					}
				}
			}

			if len(messages) < fetchMessageBatchSize {
				if err := d.Database.MarkChannelCompleted(channel.ID); err != nil {
					slog.Error("ProcessOldMessages: marking channel completed", "channelId", channel.ID, "error", err)
				}
				break
			}
			lastMessage = messages[len(messages)-1].ID
		}
	}

	if _, err := d.Session.ChannelMessageSend(interaction.ChannelID, "all messages processed"); err != nil {
		slog.Error("ProcessOldMessages: sending completion message", "error", err)
	}
}

func (d *Discord) processShakeCommand(interaction *discordgo.InteractionCreate) {
	emotes := []string{"<a:choccoREALLYhappyshakehuggers:1483236483132293293>", "<a:iraelythREALLYhappyshakehuggers:1483236185584308407>"}
	selectedAnswer := emotes[rand.Intn(len(emotes))]
	respond(d.Session, interaction.Interaction, selectedAnswer)
}
