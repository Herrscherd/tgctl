# tgctl

**Telegram Bot API client for Go.** Dependency-free: standard library only,
enforced by a test that fails on any non-stdlib import.

The Telegram counterpart of [dctl](https://github.com/Herrscherd/dctl).

No long poll loop, no event dispatch, no logging. Every call is one on-demand
HTTP request. `GetUpdates` makes a single long poll; the loop that calls it, its
backoff and its shutdown belong to the caller.

```sh
go get github.com/Herrscherd/tgctl
```

```go
bot := tgctl.New(token)

me, err := bot.GetMe(ctx)
sent, err := bot.SendMessage(ctx, tgctl.OutgoingMessage{
	ChatID: chatID,
	Text:   "deploy finished",
})

var offset int64
for {
	updates, err := bot.GetUpdates(ctx, offset) // 50 s long poll
	if err != nil {
		// handle the error, wait, retry
	}
	for _, update := range updates {
		offset = max(offset, update.UpdateID+1)
		// dispatch update
	}
}
```

## Methods

| Method | Bot API call |
|---|---|
| `GetMe(ctx)` | `getMe` |
| `GetUpdates(ctx, offset, timeout...)` | `getUpdates`, long poll (50 s by default) |
| `SendMessage(ctx, OutgoingMessage)` | `sendMessage` |
| `EditMessageText(ctx, EditMessageText)` | `editMessageText` |
| `EditMessageReplyMarkup(ctx, EditMessageReplyMarkup)` | `editMessageReplyMarkup` |
| `SendMessageDraft(ctx, MessageDraft)` | `sendMessageDraft` |
| `SetMessageReaction(ctx, chatID, messageID, emoji)` | `setMessageReaction`, an empty emoji clears it |
| `AnswerCallbackQuery(ctx, CallbackAnswer)` | `answerCallbackQuery` |
| `GetFile(ctx, fileID)` | `getFile` |
| `Download(ctx, filePath, limit, dst)` | `GET /file/bot<token>/<path>`, capped at `limit` bytes |

Wire types (`Update`, `Message`, `User`, `Chat`, `Entity`, `Document`,
`PhotoSize`, `CallbackQuery`, `File`, `OutgoingMessage`, `ReplyParameters`,
`LinkPreviewOptions`, `InlineKeyboardMarkup`, `InlineKeyboardButton`, and the
request bodies) carry the JSON tags of the Bot API. They model the fields the
herrscher gateways read or write, not the whole Bot API.

## Configuration

Nothing is read from the environment. Pass the token explicitly.

- `WithBaseURL(url)` points the client at another Bot API root, such as a local
  Bot API server or a test server. The default is `https://api.telegram.org`.
- `WithHTTPClient(client)` replaces the default client, which times out after
  70 s. That covers the default 50 s poll; a replacement must allow at least as
  long for the polls it carries.

`GetUpdates` takes an optional timeout as its last argument; only the first one
is read. Telegram answers a poll as soon as an update arrives, or when the
timeout ends.

## Errors

Every failure is a value a caller can match with `errors.Is` or `errors.As`.
Context is added with `%w`. Transport errors have the token replaced with
`[redacted]`, so it never reaches a log line through them.

| Value | Matches when |
|---|---|
| `*APIError` | Telegram refused the call. It carries `Method`, `Code`, `Description` and `RetryAfter` |
| `ErrConflict` | an `APIError` with status 409: another caller polls the bot, or a webhook is set |
| `ErrNotModified` | an `APIError` for an edit that leaves the message unchanged |
| `ErrTransport` | Telegram cannot be reached or the connection breaks; a cancelled context also matches `context.Canceled` |
| `ErrBadResponse` | the reply is not the expected JSON |
| `ErrRequest` | the request cannot be built from the configuration |
| `ErrBadFilePath` | `Download` refuses a path that is empty, absolute, not clean, or holds `?`, `#` or `\` |
| `ErrFileTooLarge` | `Download` reads more than `limit` bytes |
| `ErrFileStatus` | the file server answers with a status other than 200 |

`RetryAfter(err)` returns the delay a flood refusal (429) asks for, and whether
`err` holds one.

## Out of scope

- The long poll loop, its backoff and offset bookkeeping.
- Dispatch of updates, commands, menus and callback routing.
- Staging of downloaded files, file naming and size policy beyond the cap of
  `Download`.
- Text formatting, entity offsets and message splitting.
- Webhooks, and Bot API methods the herrscher gateways do not call.

## License

MIT
