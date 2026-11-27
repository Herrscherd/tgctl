package tgctl

// Update is one item of a getUpdates reply. The optional fields are set
// according to the kind of update Telegram sends.
type Update struct {
	UpdateID                 int64              `json:"update_id"`
	Message                  *Message           `json:"message,omitempty"`
	CallbackQuery            *CallbackQuery     `json:"callback_query,omitempty"`
	StoppedMessageGeneration *GenerationStopped `json:"stopped_message_generation,omitempty"`
}

// Message is a Telegram message as the Bot API returns it. Only the fields the
// herrscher gateways read or write are modelled.
type Message struct {
	MessageID       int64       `json:"message_id"`
	ThreadID        int64       `json:"message_thread_id,omitempty"`
	From            *User       `json:"from,omitempty"`
	Chat            Chat        `json:"chat"`
	Text            string      `json:"text,omitempty"`
	Entities        []Entity    `json:"entities,omitempty"`
	Caption         string      `json:"caption,omitempty"`
	CaptionEntities []Entity    `json:"caption_entities,omitempty"`
	ReplyTo         *Message    `json:"reply_to_message,omitempty"`
	MediaGroupID    string      `json:"media_group_id,omitempty"`
	IsTopicMessage  bool        `json:"is_topic_message,omitempty"`
	Document        *Document   `json:"document,omitempty"`
	Photo           []PhotoSize `json:"photo,omitempty"`
}

// User is a Telegram account, the bot itself included.
type User struct {
	ID           int64  `json:"id"`
	IsBot        bool   `json:"is_bot"`
	FirstName    string `json:"first_name"`
	Username     string `json:"username,omitempty"`
	LanguageCode string `json:"language_code,omitempty"`
}

// Chat is the conversation a message belongs to.
type Chat struct {
	ID      int64  `json:"id"`
	Type    string `json:"type"`
	IsForum bool   `json:"is_forum,omitempty"`
}

// Entity marks a span of a message text, such as a command, a link or a code
// block. Offset and Length count UTF-16 code units, as the Bot API does.
type Entity struct {
	Type     string `json:"type"`
	Offset   int    `json:"offset"`
	Length   int    `json:"length"`
	URL      string `json:"url,omitempty"`
	Language string `json:"language,omitempty"`
	User     *User  `json:"user,omitempty"`
}

// Document is a file sent as a document, with the metadata Telegram keeps.
type Document struct {
	FileID   string `json:"file_id"`
	FileName string `json:"file_name,omitempty"`
	MimeType string `json:"mime_type,omitempty"`
	FileSize int64  `json:"file_size,omitempty"`
}

// PhotoSize is one resolution of a photo. A message carries several, smallest
// first.
type PhotoSize struct {
	FileID   string `json:"file_id"`
	Width    int    `json:"width"`
	Height   int    `json:"height"`
	FileSize int64  `json:"file_size,omitempty"`
}

// CallbackQuery is what a user sends by pressing an inline keyboard button.
type CallbackQuery struct {
	ID      string   `json:"id"`
	From    User     `json:"from"`
	Message *Message `json:"message,omitempty"`
	Data    string   `json:"data,omitempty"`
}

// GenerationStopped reports that a user stopped the generation of a streamed
// draft.
type GenerationStopped struct {
	Chat     Chat  `json:"chat"`
	ThreadID int64 `json:"message_thread_id,omitempty"`
	DraftID  int64 `json:"draft_id"`
}

// File is the metadata of a file held by Telegram. FilePath is the relative
// path [Client.Download] reads, and it is empty until the file is requested
// with [Client.GetFile].
type File struct {
	FileID   string `json:"file_id"`
	FileSize int64  `json:"file_size,omitempty"`
	FilePath string `json:"file_path,omitempty"`
}

// OutgoingMessage is the body of a sendMessage call.
type OutgoingMessage struct {
	ChatID          int64                 `json:"chat_id"`
	ThreadID        int64                 `json:"message_thread_id,omitempty"`
	Text            string                `json:"text"`
	Entities        []Entity              `json:"entities,omitempty"`
	ReplyParameters *ReplyParameters      `json:"reply_parameters,omitempty"`
	ReplyMarkup     *InlineKeyboardMarkup `json:"reply_markup,omitempty"`
	LinkPreview     *LinkPreviewOptions   `json:"link_preview_options,omitempty"`
}

// ReplyParameters names the message a new message answers.
type ReplyParameters struct {
	MessageID                int64 `json:"message_id"`
	AllowSendingWithoutReply bool  `json:"allow_sending_without_reply"`
}

// LinkPreviewOptions controls the link preview of a message. IsDisabled true
// hides it.
type LinkPreviewOptions struct {
	IsDisabled bool `json:"is_disabled"`
}

// InlineKeyboardMarkup is the keyboard attached under a message, one slice of
// buttons per row.
type InlineKeyboardMarkup struct {
	Rows [][]InlineKeyboardButton `json:"inline_keyboard"`
}

// InlineKeyboardButton is one button of an inline keyboard. A press sends
// CallbackData back in a [CallbackQuery].
type InlineKeyboardButton struct {
	Text         string `json:"text"`
	CallbackData string `json:"callback_data"`
}

// EditMessageText is the body of an editMessageText call.
type EditMessageText struct {
	ChatID      int64                 `json:"chat_id"`
	MessageID   int64                 `json:"message_id"`
	Text        string                `json:"text"`
	Entities    []Entity              `json:"entities,omitempty"`
	LinkPreview *LinkPreviewOptions   `json:"link_preview_options,omitempty"`
	ReplyMarkup *InlineKeyboardMarkup `json:"reply_markup,omitempty"`
}

// EditMessageReplyMarkup is the body of an editMessageReplyMarkup call. An
// empty markup removes the keyboard.
type EditMessageReplyMarkup struct {
	ChatID      int64                `json:"chat_id"`
	MessageID   int64                `json:"message_id"`
	ReplyMarkup InlineKeyboardMarkup `json:"reply_markup"`
}

// MessageDraft is the body of a sendMessageDraft call, which streams partial
// text into a draft that a later call replaces. CanStop lets the user stop the
// generation.
type MessageDraft struct {
	ChatID   int64    `json:"chat_id"`
	ThreadID int64    `json:"message_thread_id,omitempty"`
	DraftID  int64    `json:"draft_id"`
	Text     string   `json:"text"`
	Entities []Entity `json:"entities,omitempty"`
	CanStop  bool     `json:"can_stop"`
}

// ReactionType is one reaction a message carries. The gateways use the emoji
// kind only.
type ReactionType struct {
	Type  string `json:"type"`
	Emoji string `json:"emoji"`
}

// CallbackAnswer is the body of an answerCallbackQuery call. Text, when set, is
// shown to the user who pressed the button.
type CallbackAnswer struct {
	CallbackQueryID string `json:"callback_query_id"`
	Text            string `json:"text,omitempty"`
}
