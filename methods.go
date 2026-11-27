package tgctl

import (
	"context"
	"time"
)

const (
	updatesPerPoll = 100
	reactionEmoji  = "emoji"
)

type updatesRequest struct {
	Offset         int64    `json:"offset,omitempty"`
	Limit          int      `json:"limit"`
	Timeout        int      `json:"timeout"`
	AllowedUpdates []string `json:"allowed_updates"`
}

type fileRequest struct {
	FileID string `json:"file_id"`
}

type reactionRequest struct {
	ChatID    int64          `json:"chat_id"`
	MessageID int64          `json:"message_id"`
	Reaction  []ReactionType `json:"reaction"`
}

// GetMe returns the account of the bot, as Telegram sees it.
func (c *Client) GetMe(ctx context.Context) (User, error) {
	var account User
	err := c.api.call(ctx, "getMe", struct{}{}, &account)
	return account, err
}

// GetUpdates long polls Telegram for the updates after offset. Pass the last
// update id plus one as the next offset, so an update is not delivered twice.
// The call waits up to timeout for an update to arrive, 50 s when none is
// given; a zero timeout returns at once. Only the first timeout is read. Only
// one caller may poll a bot at a time: a second one, or a webhook, fails with
// [ErrConflict]. The loop that calls it belongs to the caller.
func (c *Client) GetUpdates(ctx context.Context, offset int64, timeout ...time.Duration) ([]Update, error) {
	var updates []Update
	request := updatesRequest{
		Offset:         offset,
		Limit:          updatesPerPoll,
		Timeout:        pollSeconds(timeout),
		AllowedUpdates: []string{"message", "callback_query", "stopped_message_generation"},
	}
	err := c.api.call(ctx, "getUpdates", request, &updates)
	return updates, err
}

func pollSeconds(timeout []time.Duration) int {
	if len(timeout) == 0 {
		return int(defaultPoll / time.Second)
	}
	return max(int(timeout[0]/time.Second), 0)
}

// SendMessage sends a message and returns it as Telegram stored it.
func (c *Client) SendMessage(ctx context.Context, message OutgoingMessage) (Message, error) {
	var sent Message
	err := c.api.call(ctx, "sendMessage", message, &sent)
	return sent, err
}

// EditMessageText replaces the text and keyboard of a message the bot sent. An
// edit that leaves the message unchanged fails with [ErrNotModified].
func (c *Client) EditMessageText(ctx context.Context, edit EditMessageText) error {
	return c.api.call(ctx, "editMessageText", edit, nil)
}

// EditMessageReplyMarkup replaces the keyboard of a message the bot sent. An
// edit that leaves the message unchanged fails with [ErrNotModified].
func (c *Client) EditMessageReplyMarkup(ctx context.Context, edit EditMessageReplyMarkup) error {
	return c.api.call(ctx, "editMessageReplyMarkup", edit, nil)
}

// SendMessageDraft streams partial text into a draft. A later call with the
// same DraftID replaces the text shown.
func (c *Client) SendMessageDraft(ctx context.Context, draft MessageDraft) error {
	return c.api.call(ctx, "sendMessageDraft", draft, nil)
}

// SetMessageReaction sets the reaction of the bot on a message. An empty emoji
// clears the reaction.
func (c *Client) SetMessageReaction(ctx context.Context, chatID, messageID int64, emoji string) error {
	request := reactionRequest{ChatID: chatID, MessageID: messageID, Reaction: []ReactionType{}}
	if emoji != "" {
		request.Reaction = []ReactionType{{Type: reactionEmoji, Emoji: emoji}}
	}
	return c.api.call(ctx, "setMessageReaction", request, nil)
}

// AnswerCallbackQuery acknowledges a button press, which stops the loading
// state of the button on the user's side.
func (c *Client) AnswerCallbackQuery(ctx context.Context, answer CallbackAnswer) error {
	return c.api.call(ctx, "answerCallbackQuery", answer, nil)
}

// GetFile asks Telegram for the path of a file, which [Client.Download] reads.
// The Bot API serves files up to 20 MB, so a larger one has no path to read.
func (c *Client) GetFile(ctx context.Context, fileID string) (File, error) {
	var file File
	err := c.api.call(ctx, "getFile", fileRequest{FileID: fileID}, &file)
	return file, err
}
