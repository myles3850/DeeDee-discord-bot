package discord

import (
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"math/rand"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

const (
	cmdWheel      = "wheel"
	cmdEightBall  = "eight_ball"
	cmdProcessOld = "process_old_messages"
	cmdShake      = "shake"
	cmdBirthdays  = "birthdays"

	cmdSendBirthdays = "send_birthdays"
	cmdCheckBirthday = "check_birthday"
	cmdWishBirthday  = "wish_birthday"

	cmdBirthdaysAdd    = "add"
	cmdBirthdaysCheck  = "check"
	cmdBirthdaysRemove = "remove"
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

// respondEphemeral is respond, but the reply is only shown to the user who ran
// the command and isn't left in the channel.
func respondEphemeral(s *discordgo.Session, i *discordgo.Interaction, content string) {
	err := s.InteractionRespond(i, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: content,
			Flags:   discordgo.MessageFlagsEphemeral,
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
	minIntValue := 1.0
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
		{
			Name:        cmdBirthdays,
			Description: "the birthdays command group",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:        cmdBirthdaysAdd,
					Description: "add your birthday",
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Options: []*discordgo.ApplicationCommandOption{
						{
							Name:        "day",
							Description: "day of the month",
							Type:        discordgo.ApplicationCommandOptionInteger,
							Required:    true,
							MinValue:    &minIntValue,
							MaxValue:    31,
						},
						{
							Name:        "month",
							Description: "birth month",
							Type:        discordgo.ApplicationCommandOptionInteger,
							Required:    true,
							MinValue:    &minIntValue,
							MaxValue:    12,
						},
					},
				},
				{
					Name:        cmdBirthdaysCheck,
					Description: "check your birthday",
					Type:        discordgo.ApplicationCommandOptionSubCommand,
				},
				{
					Name:        cmdBirthdaysRemove,
					Description: "remove your birthday",
					Type:        discordgo.ApplicationCommandOptionSubCommand,
				},
			},
		},
		{
			Name:                     cmdSendBirthdays,
			Description:              "post today's birthday message now",
			DefaultMemberPermissions: &defaultMemberPermissions,
		},
		{
			Name:                     cmdCheckBirthday,
			Description:              "check someone's birthday",
			DefaultMemberPermissions: &defaultMemberPermissions,
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:        "user",
					Description: "whose birthday to check",
					Type:        discordgo.ApplicationCommandOptionUser,
					Required:    true,
				},
			},
		},
		{
			Name:                     cmdWishBirthday,
			Description:              "post a happy birthday message for someone",
			DefaultMemberPermissions: &defaultMemberPermissions,
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:        "user",
					Description: "who to wish happy birthday",
					Type:        discordgo.ApplicationCommandOptionUser,
					Required:    true,
				},
			},
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

func (d *Discord) processBirthdaysCommand(interaction *discordgo.InteractionCreate) {
	subcommand := interaction.ApplicationCommandData().Options[0]

	switch subcommand.Name {
	case cmdBirthdaysAdd:
		d.processAddBirthdayCommand(interaction, subcommand)
	case cmdBirthdaysCheck:
		d.processCheckBirthdayCommand(interaction, interaction.Member.User)
	case cmdBirthdaysRemove:
		d.processRemoveBirthdayCommand(interaction)
	}
}

func (d *Discord) processAddBirthdayCommand(interaction *discordgo.InteractionCreate, subcommand *discordgo.ApplicationCommandInteractionDataOption) {
	// birthdays are stored month/day only in a DATE column. 2000 is a leap
	// year so Feb 29 birthdays round-trip correctly.
	const placeholderYear = 2000

	user := interaction.Member.User
	day := subcommand.Options[0].IntValue()
	month := subcommand.Options[1].IntValue()

	// time.Date normalises out-of-range days (e.g. Feb 31 -> Mar 2), so a
	// mismatch means the date doesn't exist.
	birthday := time.Date(placeholderYear, time.Month(month), int(day), 0, 0, 0, 0, time.UTC)
	if int64(birthday.Month()) != month || int64(birthday.Day()) != day {
		respondEphemeral(d.Session, interaction.Interaction, fmt.Sprintf("sorry, %s, %02d/%02d isn't a valid date.", user.Username, day, month))
		return
	}

	if err := d.Database.SaveBirthdayForUser(user.ID, user.Username, birthday); err != nil {
		slog.Error("saving birthday", "userId", user.ID, "error", err)
		respondEphemeral(d.Session, interaction.Interaction, fmt.Sprintf("sorry, %s, i couldn't save your birthday :(", user.Username))
		return
	}

	respondEphemeral(d.Session, interaction.Interaction, fmt.Sprintf("got it, %s! i'll remember your birthday is %s.", user.Username, birthday.Format(birthdayDisplayFormat)))
}

// processCheckUserBirthdayCommand is the staff version of /birthdays check,
// which can look up anyone's birthday.
func (d *Discord) processCheckUserBirthdayCommand(interaction *discordgo.InteractionCreate) {
	target := interaction.ApplicationCommandData().Options[0].UserValue(d.Session)
	d.processCheckBirthdayCommand(interaction, target)
}

func (d *Discord) processCheckBirthdayCommand(interaction *discordgo.InteractionCreate, target *discordgo.User) {
	birthday, err := d.Database.GetBirthdayForUser(target.ID)
	if errors.Is(err, sql.ErrNoRows) {
		respondEphemeral(d.Session, interaction.Interaction, fmt.Sprintf("%s hasn't told me their birthday yet.", target.Username))
		return
	}
	if err != nil {
		slog.Error("getting birthday", "userId", target.ID, "error", err)
		respondEphemeral(d.Session, interaction.Interaction, fmt.Sprintf("sorry, i couldn't look up %s's birthday :(", target.Username))
		return
	}

	respondEphemeral(d.Session, interaction.Interaction, fmt.Sprintf("%s's birthday is %s.", target.Username, birthday.Format(birthdayDisplayFormat)))
}

func (d *Discord) processRemoveBirthdayCommand(interaction *discordgo.InteractionCreate) {
	user := interaction.Member.User

	removed, err := d.Database.RemoveBirthdayForUser(user.ID)
	if err != nil {
		slog.Error("removing birthday", "userId", user.ID, "error", err)
		respondEphemeral(d.Session, interaction.Interaction, fmt.Sprintf("sorry, %s, i couldn't remove your birthday :(", user.Username))
		return
	}
	if !removed {
		respondEphemeral(d.Session, interaction.Interaction, fmt.Sprintf("%s, you haven't told me your birthday, so there's nothing to remove.", user.Username))
		return
	}

	respondEphemeral(d.Session, interaction.Interaction, fmt.Sprintf("done, %s, i've forgotten your birthday.", user.Username))
}

func (d *Discord) processSendBirthdaysCommand(interaction *discordgo.InteractionCreate) {
	loc, err := time.LoadLocation(birthdayTimezone)
	if err != nil {
		slog.Error("loading birthday timezone", "timezone", birthdayTimezone, "error", err)
		respondEphemeral(d.Session, interaction.Interaction, "sorry, i couldn't work out what day it is :(")
		return
	}

	sent, err := d.sendBirthdayReminders(time.Now().In(loc))
	if err != nil {
		slog.Error("sending birthday reminders", "error", err)
		respondEphemeral(d.Session, interaction.Interaction, "sorry, i couldn't send today's birthday message :(")
		return
	}
	if sent == 0 {
		respondEphemeral(d.Session, interaction.Interaction, "nobody has a birthday today, so there's nothing to send.")
		return
	}

	people := "people"
	if sent == 1 {
		people = "person"
	}
	respondEphemeral(d.Session, interaction.Interaction, fmt.Sprintf("sent today's birthday message for %d %s.", sent, people))
}

func (d *Discord) processWishBirthdayCommand(interaction *discordgo.InteractionCreate) {
	target := interaction.ApplicationCommandData().Options[0].UserValue(d.Session)

	if err := d.sendBirthdayMessage([]string{target.ID}); err != nil {
		slog.Error("sending manual birthday message", "userId", target.ID, "error", err)
		respondEphemeral(d.Session, interaction.Interaction, fmt.Sprintf("sorry, i couldn't send %s a birthday message :(", target.Username))
		return
	}

	respondEphemeral(d.Session, interaction.Interaction, fmt.Sprintf("sent %s a birthday message.", target.Username))
}
