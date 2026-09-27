# webapi

Stdlib `net/http` REST API for managing guild emojis and roles, plus a health check for Docker. This layer has no state of its own — it decodes requests, calls the Discord operations it's handed, and writes JSON responses.

It parses HTTP requests and delegates to Discord for the actual work, but doesn't import the `discord` package to do it. Instead, `webapi.go` declares its own small `discordClient` interface — the three methods this package actually calls:

```go
type discordClient interface {
    GetAllEmojis() []*discordgo.Emoji
    GetAllRoles() []*discordgo.Role
    EditEmojiRoles(emojiID string, params *discordgo.EmojiParams) error
}

func NewMux(discord discordClient, checkHealth func() error) *http.ServeMux
```

This is the standard Go way to depend on "something that can do these things" without depending on the concrete type that provides them — the interface is declared by the *consumer* (`webapi`), not the producer (`discord`). `*discord.Discord` already has these three methods, so it satisfies `discordClient` automatically; nothing in the `discord` package needs to know this interface exists. `main.go` just passes the whole `*discord.Discord` instance:

```go
mux := webapi.NewMux(discordInstance, databaseInstance.Ping)
```

`checkHealth` stays a plain `func() error` rather than a one-method interface — it's a single, self-explanatory dependency (`databaseInstance.Ping`), and a named interface would add ceremony without adding clarity.

## Endpoints

### `GET /emoji`

Returns all custom emojis in the guild.

**Response** `200 OK` — array of Discord emoji objects.

```json
[
  {
    "id": "1234567890",
    "name": "choccowave",
    "roles": ["9876543210"],
    "animated": false
  }
]
```

---

### `POST /emoji/{id}/role`

Updates the role restrictions on a single emoji. Replaces the existing role list.

**URL param** — `{id}` is the Discord emoji ID.

**Request body**

```json
{
  "RoleIds": ["9876543210", "1122334455"]
}
```

`RoleIds` is required. An empty array removes all role restrictions (emoji becomes unrestricted); a missing/`null` field returns `400`.

**Response** `204 No Content` on success.

---

### `GET /role`

Returns all roles in the guild.

**Response** `200 OK` — array of Discord role objects.

```json
[
  {
    "id": "9876543210",
    "name": "Member",
    "color": 3447003
  }
]
```

---

### `GET /health`

Reports whether the bot can reach its database — used by Docker's `HEALTHCHECK` (see the root [README.md](../README.md#docker)) so a dashboard like [Arcane](https://getarcane.app/) shows accurate live status for the container.

**Response** — `200 OK` if the database is reachable, `503 Service Unavailable` otherwise.

## Listen address

The server listens on `:8080`, matching `compose.yaml`'s `3863:8080` port mapping. `main.go` sets this explicitly on the `http.Server` it builds around `webapi.NewMux(...)`.
