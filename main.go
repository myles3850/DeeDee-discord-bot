package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"choccobear.tech/deedee/database"
	"choccobear.tech/deedee/discord"
	"choccobear.tech/deedee/googleplatform"
	"choccobear.tech/deedee/webapi"

	"github.com/joho/godotenv"
)

var (
	databaseInstance *database.Db
	discordInstance  *discord.Discord
)

func init() {
	_ = godotenv.Load()

	sheet, err := googleplatform.NewSheet(context.Background())
	if err != nil {
		log.Fatalf("initialising Google Sheet: %v", err)
	}

	databaseInstance, err = database.Setup()
	if err != nil {
		log.Fatalf("initialising database: %v", err)
	}

	discordInstance, err = discord.Setup(databaseInstance, sheet)
	if err != nil {
		log.Fatalf("initialising discord session: %v", err)
	}
}

func main() {
	discordInstance.Session.AddHandler(discordInstance.OnReady)
	discordInstance.Session.AddHandler(discordInstance.OnInteraction)
	discordInstance.Session.AddHandler(discordInstance.OnMessageCreate)
	discordInstance.Session.AddHandler(discordInstance.OnMessageDelete)
	discordInstance.Session.AddHandler(discordInstance.OnMessageModified)
	discordInstance.Session.AddHandler(discordInstance.OnMemberUpdated)

	if err := discordInstance.Session.Open(); err != nil {
		log.Fatalf("opening discord session: %v", err)
	}
	slog.Info("bot is running and connected to discord")

	mux := webapi.NewMux(discordInstance, databaseInstance.Ping)
	httpServer := &http.Server{Addr: ":8080", Handler: mux}

	go func() {
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("http server error", "error", err)
		}
	}()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	<-ctx.Done()

	slog.Info("shutting down")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		slog.Error("shutting down http server", "error", err)
	}
	if err := discordInstance.Session.Close(); err != nil {
		slog.Error("closing discord session", "error", err)
	}
	if err := databaseInstance.Close(); err != nil {
		slog.Error("closing database", "error", err)
	}
}
