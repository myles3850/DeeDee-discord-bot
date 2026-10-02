// Package webapi is the HTTP layer for managing guild emojis and roles. It
// has no state of its own — it decodes requests, calls the Discord
// operations it's given, and writes JSON responses.
package webapi

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/bwmarrin/discordgo"
)

// discordClient is the slice of *discord.Discord this package calls. It's
// declared here, by the consumer, rather than imported from the discord
// package — the standard Go way to depend on "something that can do these
// three things" without depending on the concrete type that provides them.
// *discord.Discord satisfies this automatically; nothing on that side needs
// to know this interface exists.
type discordClient interface {
	GetAllEmojis() []*discordgo.Emoji
	GetAllRoles() []*discordgo.Role
	EditEmojiRoles(emojiID string, params *discordgo.EmojiParams) error
}

// NewMux builds the routes for the emoji/role API plus a health check.
func NewMux(discord discordClient, checkHealth func() error) *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /emoji", func(w http.ResponseWriter, r *http.Request) {
		emojis := discord.GetAllEmojis()
		if emojis == nil {
			writeError(w, http.StatusInternalServerError, "could not load emojis")
			return
		}
		writeJSON(w, http.StatusOK, emojis)
	})

	mux.HandleFunc("GET /role", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, discord.GetAllRoles())
	})

	mux.HandleFunc("POST /emoji/{id}/role", func(w http.ResponseWriter, r *http.Request) {
		params, err := decodeEmojiRoleUpdate(r)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		if err := discord.EditEmojiRoles(r.PathValue("id"), params); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		if err := checkHealth(); err != nil {
			writeError(w, http.StatusServiceUnavailable, err.Error())
			return
		}
		w.WriteHeader(http.StatusOK)
	})

	return mux
}

type emojiRoleUpdate struct {
	Roles []string `json:"RoleIds"`
}

func decodeEmojiRoleUpdate(r *http.Request) (*discordgo.EmojiParams, error) {
	var update emojiRoleUpdate
	if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
		return nil, fmt.Errorf("decoding request body: %w", err)
	}
	if update.Roles == nil {
		return nil, fmt.Errorf("RoleIds is required")
	}
	return &discordgo.EmojiParams{Roles: update.Roles}, nil
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		slog.Error("encoding response", "error", err)
	}
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
