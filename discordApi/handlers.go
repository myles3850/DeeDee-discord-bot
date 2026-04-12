package discordapi

import (
	"github.com/bwmarrin/discordgo"
)

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

func (d *Discord) OnReady(s *discordgo.Session, r *discordgo.Ready) {
	d.RegisterCommands()
}

func (d *Discord) OnMessageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	d.saveMessageToDb(m);

	//if were in new members or welcome AND the message isn't a reply
	if (m.ChannelID == "1483226521014636762" || m.ChannelID == "1483226521014636763") && m.ReferencedMessage == nil {
		d.reactToIntroMessage(m)
	}
}

func (d *Discord) OnMessageDelete(s *discordgo.Session, m *discordgo.MessageDelete) {
	d.handleDeletedMessage(m)
}

func (d *Discord) OnMessageModified(s *discordgo.Session, m *discordgo.MessageUpdate) {
	if m.Author.Bot {
		return
	}
	d.reportModifiedMessage(m)
}
