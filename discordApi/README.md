# discordApi

Discord bot integration. Manages the gateway connection, registers slash commands, and handles all incoming events.

## Setup

`Setup(db, sheet)` creates a discordgo session with the token from `DISCORD_BOT_TOKEN` and the guild ID from `DISCORD_GUILD_ID`. It stores references to the database and Google Sheets clients on the `Discord` struct so all handlers can reach them.

Gateway intents requested:

| Intent | Why |
|---|---|
| `IntentsGuilds` | Guild metadata (channels, roles, emojis) |
| `IntentsGuildMessages` | Message create / update / delete events |
| `IntentsMessageContent` | Access to message text |
| `IntentsGuildMessageReactions` | Reaction add / remove events |

## Event handlers

Handlers are methods on the `Discord` struct and are registered in `main.go` using `session.AddHandler(...)`. Because they live on the struct, every handler has access to `d.Database`, `d.Sheet`, and `d.Session` without needing to pass anything around.

discordgo works out which handler to call by matching the function signature — each event type has its own payload struct (e.g. `*discordgo.MessageCreate`), so you just write a method with the right signature and register it.

| Handler | Trigger | What it does |
|---|---|---|
| `OnReady` | Bot connects | Registers all slash commands with Discord |
| `OnMessageCreate` | New message | Saves the user and message to the DB; auto-reacts to intro posts in the new-member channels |
| `OnMessageModified` | Message edited | Pushes the previous content into `edit_history` in the DB; skips embed-load updates that Discord fires with no `EditedTimestamp` |
| `OnMessageDelete` | Message deleted | Posts a rich embed to the mod channel; fills in any fields Discord omits from the event by looking them up in the DB |
| `OnInteraction` | Slash command used | Reads the command name and dispatches to the right processor function |

### Adding a new event handler

Only one handler is registered per event type. Rather than registering multiple handlers for the same event, the handler itself calls out to focused functions — one per task. `OnMessageCreate` is a good example of this pattern already:

```go
func (d *Discord) OnMessageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
    d.saveMessageToDb(m)
    d.reactToIntroMessage(m)
}
```

Each task is its own method on `*Discord`, kept in `messages.go`. The handler just coordinates them. To add new behaviour for an existing event, write a new method and call it from the handler — don't register a second handler for the same event type.

**If you need to handle a new event type that isn't covered yet:**

1. Write the handler method in `handlers.go` with the appropriate discordgo signature:
   ```go
   func (d *Discord) OnReactionAdd(s *discordgo.Session, r *discordgo.MessageReactionAdd) {
       d.someReactionTask(r)
   }
   ```
2. Write the task function(s) it calls — by convention these live in a file named after what they do (e.g. `reactions.go`).
3. Register the handler in `main.go`:
   ```go
   discordInstance.Session.AddHandler(discordInstance.OnReactionAdd)
   ```
4. Make sure the relevant intent is requested in `Setup()` — without it Discord won't send the event.

## Slash commands

Registered on `OnReady` via the Discord Application Commands API. All commands are guild-scoped (`DISCORD_GUILD_ID`).

### How it works

There are three moving parts that need to stay in sync:

**1. The name constant** — every command name is stored as a field on the `CommandName` struct and set in the `names` variable at the top of `slashCommands.go`. This means the name is defined once and referenced everywhere, so a typo won't silently break routing.

**2. The command definition** — inside `RegisterCommands()` there is a `commands` slice. Each entry tells Discord what the command is called, what it does, and what options (arguments) it accepts. Discord shows this to users in the slash command picker. Options have a type (`String`, `Integer`, etc.), a description, and a `Required` flag. The full list of available option types is on the [discordgo pkg.go.dev page for `ApplicationCommandOptionType`](https://pkg.go.dev/github.com/bwmarrin/discordgo#ApplicationCommandOptionType).

**3. The processor function** — the actual logic. When a user runs a command, `OnInteraction` reads the command name and uses a `switch` to call the right processor:

```go
switch data.Name {
case names.wheel:
    d.processWheelCommand(i)
case names.eightBall:
    d.process8BallCommand(i)
// ...
}
```

Each processor is a method on `*Discord`, so it has full access to `d.Database`, `d.Session`, and `d.Sheet`. They live in `slashCommands.go` and respond to the user by calling `d.Session.InteractionRespond(...)`.

### Reading an existing command

To understand any command, find these three things in `slashCommands.go`:

- Its entry in the `names` variable — this is the string Discord routes on
- Its entry in the `commands` slice in `RegisterCommands()` — this shows what options the user can pass
- Its `process<Name>Command` method — this is what actually runs

Options passed by the user come through as `interaction.ApplicationCommandData().Options`. Each option is accessed by index or by iterating the slice, and values are read with `.StringValue()`, `.IntValue()`, etc. depending on the type.

### Adding a new slash command

The example below adds a `/coinflip` command that picks heads or tails.

1. **Register the name.** Add a field to the `CommandName` struct and set it in `names`:
   ```go
   type CommandName struct {
       // ...
       coinFlip string
   }
   var names = &CommandName{..., coinFlip: "coinflip"}
   ```

2. **Define the command.** Add an entry to the `commands` slice in `RegisterCommands()`. This one takes no options — just omit the `Options` field entirely:
   ```go
   {
       Name:        names.coinFlip,
       Description: "Flip a coin",
   },
   ```

3. **Write the processor.** Add a method to `*Discord` in `slashCommands.go`:
   ```go
   func (d *Discord) processCoinFlipCommand(i *discordgo.InteractionCreate) {
       sides := []string{"Heads", "Tails"}
       result := sides[rand.Intn(len(sides))]

       d.Session.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
           Type: discordgo.InteractionResponseChannelMessageWithSource,
           Data: &discordgo.InteractionResponseData{
               Content: fmt.Sprintf("🪙 %s!", result),
           },
       })
   }
   ```

4. **Wire up the router.** Add a `case` to the `switch` in `OnInteraction` in `handlers.go`:
   ```go
   case names.coinFlip:
       d.processCoinFlipCommand(i)
   ```

The command registers with Discord automatically the next time the bot starts. During development you may need to wait a moment for Discord to propagate the new command before it appears in the UI.

### Working with options

`/coinflip` takes no input, but most commands will need options. Here's how to add them, using `/eight_ball` in the codebase as a real reference point.

**Defining an option** — add it to the `Options` field in the command definition:
```go
{
    Name:        names.coinFlip,
    Description: "Flip a coin",
    Options: []*discordgo.ApplicationCommandOption{
        {
            Name:        "wager",
            Description: "What are you betting?",
            Type:        discordgo.ApplicationCommandOptionString,
            Required:    false,
        },
    },
},
```

Set `Required: true` if the command shouldn't run without it. Discord enforces this in the UI before the interaction is ever sent to the bot.

**Reading the value in the processor** — options arrive as a slice on the interaction data. For required options you can access by index safely; for optional ones, check the length first:

```go
func (d *Discord) processCoinFlipCommand(i *discordgo.InteractionCreate) {
    sides := []string{"Heads", "Tails"}
    result := sides[rand.Intn(len(sides))]

    // read the optional wager if the user provided one
    var wagerText string
    opts := i.ApplicationCommandData().Options
    if len(opts) > 0 {
        wagerText = fmt.Sprintf(" You wagered: %s.", opts[0].StringValue())
    }

    d.Session.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
        Type: discordgo.InteractionResponseChannelMessageWithSource,
        Data: &discordgo.InteractionResponseData{
            Content: fmt.Sprintf("🪙 %s!%s", result, wagerText),
        },
    })
}
```

For the full list of option types (`String`, `Integer`, `Boolean`, `User`, `Channel`, etc.) see the [discordgo `ApplicationCommandOptionType` docs](https://pkg.go.dev/github.com/bwmarrin/discordgo#ApplicationCommandOptionType).

### `/wheel`

Picks one option at random from up to six choices.

| Option | Required | Description |
|---|---|---|
| `option1` | yes | First choice |
| `option2` | yes | Second choice |
| `option3–6` | no | Additional choices |

### `/eight_ball`

Asks the magic 8-ball a yes/no question. Returns one of 20 classic answers, or `42` if the question is the ultimate question.

| Option | Required | Description |
|---|---|---|
| `question` | yes | The question to ask |

### `/process_old_messages`

Bulk-archives all historical messages from every guild channel into the database. Skips channels that are already marked complete. Processes in batches of 100. Requires **Manage Guild** permission.

### `/shake`

Sends one of two custom animated shake emotes at random.

## Guild data methods

These are called by the REST API layer (`webApi`) via the `Discord` struct:

| Method | Description |
|---|---|
| `GetAllEmojis()` | Returns all custom emojis in the guild |
| `GetOneEmoji(id)` | Returns a single emoji by ID |
| `GetAllRoles()` | Returns all roles in the guild |
| `EditEmojiRoles(emojiId, params)` | Updates the role restrictions on an emoji |
