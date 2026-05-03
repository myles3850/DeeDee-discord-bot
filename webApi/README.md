# webApi

Gin-based REST API for managing guild emojis and roles. This layer has no state of its own — it parses HTTP requests and delegates to the `discordApi` package for all Discord operations.

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

### `POST /emoji/:id/role`

Updates the role restrictions on a single emoji. Replaces the existing role list.

**URL param** — `:id` is the Discord emoji ID.

**Request body**

```json
{
  "RoleIds": ["9876543210", "1122334455"]
}
```

`RoleIds` is required. An empty array removes all role restrictions (emoji becomes unrestricted).

**Response** — delegates to Discord; returns no body on success.

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

## Default port

Gin defaults to `:3000`. Override with the standard `PORT` environment variable or Gin's `GIN_MODE`.
