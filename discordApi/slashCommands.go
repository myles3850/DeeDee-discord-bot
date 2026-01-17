package discordapi

import (
	"fmt"
	"math/rand"
	"strings"

	"choccobear.tech/emojiBot/database"
	"github.com/bwmarrin/discordgo"
)

type CommandName struct {
	wheel      string
	eightBall  string
	processOld string
	shake      string
}

var names = &CommandName{wheel: "wheel", eightBall: "eight_ball", processOld: "process_old_messages", shake: "shake"}

func (d *Discord) RegisterCommands() {
	s := d.Session
	appID := s.State.User.ID
	guildID := d.GuildId
	minStringLength := 2
	var defaultMemberPermissions int64 = discordgo.PermissionManageGuild

	commands := []*discordgo.ApplicationCommand{
		{
			Name:        names.wheel,
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
			Name:        names.eightBall,
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
			Name:                     names.processOld,
			Description:              "run through all old messages",
			DefaultMemberPermissions: &defaultMemberPermissions,
		},
		{
			Name:        names.shake,
			Description: "sends special shake emote",
		},
	}

	for _, cmd := range commands {
		_, err := s.ApplicationCommandCreate(appID, guildID, cmd)
		if err != nil {
			fmt.Printf("❌ Cannot create command '%v': %v\n", cmd.Name, err)
		} else {
			fmt.Printf("✅ Registered command '%v'\n", cmd.Name)
		}
	}
}

func (d *Discord) OnInteraction(s *discordgo.Session, i *discordgo.InteractionCreate) {

	if i.Type != discordgo.InteractionApplicationCommand {
		return
	}

	data := i.ApplicationCommandData()

	switch data.Name {
	case names.wheel:
		d.processWheelCommand(i)
		return
	case names.eightBall:
		d.process8BallCommand(i)
		return
	case names.processOld:
		d.ProcessOldMessages(i)
		return
	case names.shake:
		d.processShakeCommand(i)
	}

}

//from here all functions are processing functions

func (d *Discord) processWheelCommand(interaction *discordgo.InteractionCreate) {
	var options []string
	data := interaction.ApplicationCommandData()
	session := d.Session
	user := interaction.Member.Nick

	if len(data.Options) == 0 {
		session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("sorry, %s, but i cant seem to find your options :(", user),
			},
		})

	}
	for _, option := range data.Options {
		options = append(options, option.StringValue())
	}

	// i know im going to forget how this shit works, it gets the length of options and passes to intn which gives a random number based on how many are in the slice
	chosen := options[rand.Intn(len(options))]

	session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: fmt.Sprintf("Interesting choices... im feeling %s this time.", chosen),
		},
	})
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
		"Very doubtful"}

	session := d.Session
	data := interaction.ApplicationCommandData()
	question := data.Options[0].StringValue()
	user := interaction.Member.Nick

	if len(question) == 0 {
		session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("sorry, %s, but the ball refuses to answer my call :(", user),
			},
		})
		return
	}

	questionOfLife := "what is the answer to life the universe and everything"

	if strings.Contains(question, questionOfLife) {
		session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("I asked the magical 8 ball \"%s\" , and it said **42**.", question),
			},
		})
		return
	}

	selectedAnswer := ballAnswers[rand.Intn(len(ballAnswers))]

	session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: fmt.Sprintf("I asked the magical 8 ball \"%s\" , and it said **%s**.", question, selectedAnswer),
		},
	})
}

func (d *Discord) ProcessOldMessages(interaction *discordgo.InteractionCreate) {
	const fetchMessageBatchSize = 100

	d.Session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "getting you the info now...",
		},
	})

	channels, err := d.Session.GuildChannels(d.GuildId)
	if err != nil {
		fmt.Printf("%+v", err)
	}

	for _, channel := range channels {
		channelComplete, _ := d.Database.IsChannelCompleted(channel.ID)
		d.Database.SaveChannelName(channel.ID, channel.Name)
		if channelComplete {
			fmt.Printf("skipping channel %s as already completed \n", channel.Name)
			continue
		}
		var lastMessage string
		for {
			fmt.Printf("processing messages from channel %s \n", channel.Name)
			messages, err := d.Session.ChannelMessages(channel.ID, fetchMessageBatchSize, lastMessage, "", "")
			if err != nil {
				fmt.Printf("%+v", err)
			}
			if len(messages) == 0 {
				break
			}
			for _, message := range messages {
				fmt.Printf("Processing message: %s\n", message.ID)
				userID, err := d.Database.SaveUser(&database.User{
					DiscordID:       message.Author.ID,
					DiscordUsername: message.Author.Username,
				})
				if err != nil {
					fmt.Printf("Error saving user: %+v\n", err)
					break
				}
				messageID, err := d.Database.SaveMessage(&database.Message{
					DiscordMessageID: message.ID,
					ChannelID:        channel.ID,
					AuthorID:         userID,
					CreatedAt:        message.Timestamp,
				})
				if err != nil {
					fmt.Printf("Error saving message: %+v\n", err)
					break
				}
				if len(message.Reactions) > 0 {
					for _, react := range message.Reactions {
						d.Database.SaveReaction(&database.Reaction{
							MessageID: messageID,
							Emoji:     react.Emoji.Name,
							ReactorID: userID,
						})
					}
				}
			}
			if len(messages) < fetchMessageBatchSize {
				_ = d.Database.MarkChannelCompleted(channel.ID)
				lastMessage = ""
				break
			}
			lastMessage = messages[len(messages)-1].ID
		}
	}
	d.Session.ChannelMessageSend(interaction.ChannelID, "all messages processed")
}

func (d *Discord) processShakeCommand(interaction *discordgo.InteractionCreate) {
	session := d.Session
	emotes := []string{"<a:choccoREALLYhappyshakehuggers:1460017063513821267>", "<a:iraelythREALLYhappyshakehuggers:1462197698236256389>"}

	selectedAnswer := emotes[rand.Intn(len(emotes))]

	session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: selectedAnswer,
		},
	})
}
