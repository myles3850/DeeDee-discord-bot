package main

import (
	"choccobear.tech/emojiBot/database"
	discordapi "choccobear.tech/emojiBot/discordApi"
	webapi "choccobear.tech/emojiBot/webApi"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

var databaseInstance *database.Db
var discordInstance *discordapi.Discord
var apiInstance *webapi.WebCtx

func init() {
	godotenv.Load()
	var error error

	databaseInstance = database.Setup()
	discordInstance, error = discordapi.Setup(databaseInstance)
	apiInstance = webapi.Setup()

	if error != nil {
		panic(3)
	}

}

func main() {
	defer databaseInstance.Session.Close()

	discordInstance.Session.AddHandler(discordInstance.OnReady)
	discordInstance.Session.AddHandler(discordInstance.OnInteraction)
	discordInstance.Session.AddHandler(discordInstance.OnMessageCreate)
	discordInstance.Session.AddHandler(discordInstance.OnMessageDelete)

	if err := discordInstance.Session.Open(); err != nil {
		panic("Error opening Discord session: " + err.Error())
	}
	defer discordInstance.Session.Close()
	println("🤖 Bot is running and connected to Discord!")

	server := apiInstance.Gin
	registerApiEndpoints(apiInstance.Gin, discordInstance, apiInstance)

	server.Run()

}

func registerApiEndpoints(server *gin.Engine, discord *discordapi.Discord, web *webapi.WebCtx) {
	server.GET("emoji", func(ctx *gin.Context) {
		emojis := discord.GetAllEmojis()
		web.GetAllEmojis(ctx, emojis)
	})

	server.POST("emoji/:id/role", func(ctx *gin.Context) {
		id := ctx.Param("id")
		discord.EditEmojiRoles(id, web.UpdateEmojiRoles(ctx))
	})

	server.GET("role", func(ctx *gin.Context) {
		web.GetAllRoles(ctx, discord.GetAllRoles())
	})
}
