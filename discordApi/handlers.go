package discordapi

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
)

func (d *Discord) HandleMessageCreated(s *discordgo.Session, i *discordgo.MessageCreate) {
	go d.logMessageDetails(i)
}

func (d *Discord) logMessageDetails(i *discordgo.MessageCreate){
	author := i.Author
	messageLength := len(i.Content)
	channel := i.ChannelID

	fmt.Printf("message received: from:%s, in:%s, message Length:%v", author, channel, messageLength)
}