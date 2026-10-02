package discord

import (
	"github.com/bwmarrin/discordgo"
)

func (d *Discord) OnReady(s *discordgo.Session, r *discordgo.Ready) {
	d.RegisterCommands()
}

func (d *Discord) OnInteraction(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.Type != discordgo.InteractionApplicationCommand {
		return
	}

	data := i.ApplicationCommandData()

	switch data.Name {
	case cmdWheel:
		d.processWheelCommand(i)
	case cmdEightBall:
		d.process8BallCommand(i)
	case cmdProcessOld:
		d.ProcessOldMessages(i)
	case cmdShake:
		d.processShakeCommand(i)
	}
}

func (d *Discord) OnMessageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	d.saveMessageToDb(m)

	introChannel := m.ChannelID == newMemberChannelID || m.ChannelID == welcomeChannelID
	if introChannel && m.ReferencedMessage == nil {
		d.reactToIntroMessage(m)
	}
}

func (d *Discord) OnMessageDelete(s *discordgo.Session, m *discordgo.MessageDelete) {
	d.handleDeletedMessage(m)
}

func (d *Discord) OnMessageModified(s *discordgo.Session, m *discordgo.MessageUpdate) {
	if m.Author == nil || m.Author.Bot {
		return
	}
	// EditedTimestamp is only set when the user actually edits message text, so if it's
	// nil this is an embed-load update, not a real edit.
	if m.EditedTimestamp == nil {
		return
	}
	d.saveModifiedMessage(m)
}

func (d *Discord) OnMemberUpdated(s *discordgo.Session, m *discordgo.GuildMemberUpdate) {
	d.timeoutBotRole(m)
}
