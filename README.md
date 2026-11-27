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
| `GetUpdates(ctx, offset, timeout...)` | `getUpdates`, long poll (50 s by default, at most 50 s) |
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
- `WithHTTPClient(client)` replaces the default client. The default times out
  after 70 s, which covers the longest poll of 50 s, and refuses redirects: the
  token travels in the URL path, so a 3xx answer is returned as the reply, as
  `ErrBadResponse` for a method call or `ErrFileStatus` for a download. A
  replacement must allow at least 50 s for the polls it carries, and should
  refuse redirects too.

`GetUpdates` takes an optional timeout as its last argument; only the first one
is read. Telegram answers a poll as soon as an update arrives, or when the
timeout ends. The Bot API counts whole seconds: a positive timeout is rounded up
to the next second, and a timeout that is negative or over 50 s fails with
`ErrBadTimeout` before any request is sent. `Download` likewise refuses a limit
that is not positive with `ErrBadLimit`.

## Errors

Every failure is a value a caller can match with `errors.Is` or `errors.As`.
Context is added with `%w`. Transport errors have the token replaced with
`[redacted]`, so it never reaches a log line through them.

| Value | Matches when |
|---|---|
| `*APIError` | Telegram refused the call. It carries `Method`, `Code`, `Description` and `RetryAfter` |
| `ErrConflict` | an `APIError` with status 409: another caller polls the bot, or a webhook is set |
| `ErrNotModified` | an `APIError` for an edit that leaves the message unchanged |
| `ErrTransport` | Telegram cannot be reached, the connection breaks, or a file body breaks while `Download` reads it; a context that ends during the call also matches `context.Canceled` when it was cancelled, and `context.DeadlineExceeded` when it passed its deadline |
| `ErrWrite` | the `dst` of `Download` refuses bytes; the writer's own error is wrapped, so it matches too, and it never matches `ErrTransport` |
| `ErrBadResponse` | the reply is not the expected JSON, is longer than 8 MiB, or is a redirect the default client refused to follow; the text names the HTTP status, as `http 502`, and never the URL or the token |
| `ErrRequest` | the request cannot be built from the configuration |
| `ErrBadTimeout` | `GetUpdates` is given a timeout that is negative or over 50 s |
| `ErrBadLimit` | `Download` is given a limit that is not positive |
| `ErrBadFilePath` | `Download` refuses a path that is empty, absolute, not clean, or holds `?`, `#`, `\` or `%` |
| `ErrFileTooLarge` | the file holds more than `limit` bytes; the first `limit` bytes are written to `dst`, and no byte beyond them is |
| `ErrFileStatus` | the file server answers with a status other than 200, a redirect included |

`RetryAfter(err)` returns the delay a flood refusal (429) asks for, and whether
`err` holds one.

## Differences from a plain HTTP client

Three behaviours differ from `net/http` used directly. They are intended, and a
migration that moves a gateway onto this client can rely on them.

1. **Redirects are refused.** The default client does not follow a 3xx answer,
   because the token travels in the URL path and a redirect would resend it to
   another host. The 3xx is the reply: `ErrBadResponse` for a method call,
   `ErrFileStatus` for a download.
2. **A limit of zero or less is refused before any request.** `Download` returns
   `ErrBadLimit` without contacting Telegram.
3. **The limit is checked before the path.** `Download` tests the limit first and
   the path second, so a call with both wrong returns `ErrBadLimit`. Neither
   check sends a request.

## Out of scope

- The long poll loop, its backoff and offset bookkeeping.
- Dispatch of updates, commands, menus and callback routing.
- Staging of downloaded files, file naming and size policy beyond the cap of
  `Download`.
- Text formatting, entity offsets and message splitting.
- Webhooks, and Bot API methods the herrscher gateways do not call.

## License

MIT
