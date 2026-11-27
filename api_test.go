package tgctl

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestGetMeDecodesTheBotAccount(t *testing.T) {
	bot := newFakeBot(t, okReply(`{"id":777,"is_bot":true,"first_name":"Herrscher","username":"herrscher_bot"}`))

	me, err := bot.client().GetMe(t.Context())
	if err != nil {
		t.Fatal(err)
	}

	if me.ID != 777 || !me.IsBot || me.Username != "herrscher_bot" {
		t.Errorf("GetMe = %+v", me)
	}
	if got := bot.recorded()[0]; got.path != "/bot"+testToken+"/getMe" || got.body != "{}" {
		t.Errorf("request = %+v", got)
	}
}

func TestGetUpdatesLongPollsWithTheRequestedWindow(t *testing.T) {
	cases := []struct {
		name     string
		offset   int64
		timeout  []time.Duration
		wantJSON string
	}{
		{
			name:     "default window, no offset",
			wantJSON: `{"limit":100,"timeout":50,"allowed_updates":["message","callback_query","stopped_message_generation"]}`,
		},
		{
			name:     "explicit offset and window",
			offset:   42,
			timeout:  []time.Duration{3 * time.Second},
			wantJSON: `{"offset":42,"limit":100,"timeout":3,"allowed_updates":["message","callback_query","stopped_message_generation"]}`,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			bot := newFakeBot(t, okReply(`[{"update_id":5}]`))

			updates, err := bot.client().GetUpdates(t.Context(), tc.offset, tc.timeout...)
			if err != nil {
				t.Fatal(err)
			}

			if len(updates) != 1 || updates[0].UpdateID != 5 {
				t.Errorf("updates = %+v", updates)
			}
			if got := bot.recorded()[0].body; got != tc.wantJSON {
				t.Errorf("body = %s\nwant %s", got, tc.wantJSON)
			}
		})
	}
}

func TestSendMessageEncodesEveryOptionalField(t *testing.T) {
	const wantBody = `{"chat_id":-100500,"message_thread_id":7,"text":"hi",` +
		`"reply_parameters":{"message_id":3,"allow_sending_without_reply":true},` +
		`"reply_markup":{"inline_keyboard":[[{"text":"ok","callback_data":"c:x:0"}]]},` +
		`"link_preview_options":{"is_disabled":true}}`
	bot := newFakeBot(t, okReply(`{"message_id":901,"chat":{"id":-100500,"type":"supergroup"}}`))

	sent, err := bot.client().SendMessage(t.Context(), OutgoingMessage{
		ChatID:          -100500,
		ThreadID:        7,
		Text:            "hi",
		ReplyParameters: &ReplyParameters{MessageID: 3, AllowSendingWithoutReply: true},
		ReplyMarkup: &InlineKeyboardMarkup{Rows: [][]InlineKeyboardButton{
			{{Text: "ok", CallbackData: "c:x:0"}},
		}},
		LinkPreview: &LinkPreviewOptions{IsDisabled: true},
	})
	if err != nil {
		t.Fatal(err)
	}

	if sent.MessageID != 901 || sent.Chat.Type != "supergroup" {
		t.Errorf("sent = %+v", sent)
	}
	if got := bot.recorded()[0].body; got != wantBody {
		t.Errorf("body = %s\nwant %s", got, wantBody)
	}
}

func TestSendMessageOmitsUnsetOptionalFields(t *testing.T) {
	bot := newFakeBot(t, okReply(`{"message_id":1,"chat":{"id":1,"type":"private"}}`))

	if _, err := bot.client().SendMessage(t.Context(), OutgoingMessage{ChatID: 1, Text: "hi"}); err != nil {
		t.Fatal(err)
	}

	if got := bot.recorded()[0].body; got != `{"chat_id":1,"text":"hi"}` {
		t.Errorf("body = %s", got)
	}
}

func TestEditMessageTextUnchangedMatchesErrNotModified(t *testing.T) {
	bot := newFakeBot(t, refusal(http.StatusBadRequest, "Bad Request: message is not modified: specified new message content"))

	err := bot.client().EditMessageText(t.Context(), EditMessageText{ChatID: 1, MessageID: 2, Text: "same"})

	if !errors.Is(err, ErrNotModified) {
		t.Fatalf("err = %v, want ErrNotModified", err)
	}
	if got := bot.recorded()[0].body; got != `{"chat_id":1,"message_id":2,"text":"same"}` {
		t.Errorf("body = %s", got)
	}
}

func TestEditMessageReplyMarkupEmptyKeyboardClearsIt(t *testing.T) {
	bot := newFakeBot(t, okReply(`true`))

	if err := bot.client().EditMessageReplyMarkup(t.Context(), EditMessageReplyMarkup{ChatID: 1, MessageID: 2}); err != nil {
		t.Fatal(err)
	}

	if got := bot.recorded()[0].body; got != `{"chat_id":1,"message_id":2,"reply_markup":{"inline_keyboard":null}}` {
		t.Errorf("body = %s", got)
	}
}

func TestSendMessageDraftCarriesTheStreamedText(t *testing.T) {
	bot := newFakeBot(t, okReply(`true`))

	err := bot.client().SendMessageDraft(t.Context(), MessageDraft{ChatID: 1, ThreadID: 5, DraftID: 9, Text: "part", CanStop: true})
	if err != nil {
		t.Fatal(err)
	}
	if got := bot.recorded()[0].body; got != `{"chat_id":1,"message_thread_id":5,"draft_id":9,"text":"part","can_stop":true}` {
		t.Errorf("body = %s", got)
	}
}

func TestSetMessageReactionEmptyEmojiClearsTheReaction(t *testing.T) {
	cases := []struct {
		name     string
		emoji    string
		wantJSON string
	}{
		{"clear", "", `{"chat_id":1,"message_id":2,"reaction":[]}`},
		{"set", "👍", `{"chat_id":1,"message_id":2,"reaction":[{"type":"emoji","emoji":"👍"}]}`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			bot := newFakeBot(t, okReply(`true`))

			if err := bot.client().SetMessageReaction(t.Context(), 1, 2, tc.emoji); err != nil {
				t.Fatal(err)
			}

			if got := bot.recorded()[0].body; got != tc.wantJSON {
				t.Errorf("body = %s\nwant %s", got, tc.wantJSON)
			}
		})
	}
}

func TestAnswerCallbackQuerySendsTheAcknowledgement(t *testing.T) {
	bot := newFakeBot(t, okReply(`true`))

	if err := bot.client().AnswerCallbackQuery(t.Context(), CallbackAnswer{CallbackQueryID: "q1", Text: "done"}); err != nil {
		t.Fatal(err)
	}

	if got := bot.recorded()[0].body; got != `{"callback_query_id":"q1","text":"done"}` {
		t.Errorf("body = %s", got)
	}
}

func TestGetFileReturnsThePathToDownload(t *testing.T) {
	bot := newFakeBot(t, okReply(`{"file_id":"f1","file_size":12,"file_path":"documents/a.txt"}`))

	file, err := bot.client().GetFile(t.Context(), "f1")
	if err != nil {
		t.Fatal(err)
	}

	if file.FilePath != "documents/a.txt" || file.FileSize != 12 {
		t.Errorf("file = %+v", file)
	}
	if got := bot.recorded()[0].body; got != `{"file_id":"f1"}` {
		t.Errorf("body = %s", got)
	}
}

func TestFloodRefusalCarriesItsRetryDelay(t *testing.T) {
	bot := newFakeBot(t, func(string, string) (int, string) {
		return http.StatusTooManyRequests, `{"ok":false,"error_code":429,"description":"Too Many Requests","parameters":{"retry_after":1}}`
	})

	_, err := bot.client().SendMessage(t.Context(), OutgoingMessage{ChatID: 1, Text: "hi"})

	wait, limited := RetryAfter(err)
	if !limited || wait != time.Second {
		t.Fatalf("RetryAfter = %v, %v from %v", wait, limited, err)
	}
	failure, ok := errors.AsType[*APIError](err)
	if !ok || failure.Code != http.StatusTooManyRequests || failure.Method != "sendMessage" {
		t.Errorf("APIError = %+v, %v", failure, ok)
	}
}

func TestConflictMatchesErrConflictOnly(t *testing.T) {
	bot := newFakeBot(t, refusal(http.StatusConflict, "Conflict: terminated by other getUpdates request"))

	_, err := bot.client().GetUpdates(t.Context(), 0)

	if !errors.Is(err, ErrConflict) {
		t.Fatalf("err = %v, want ErrConflict", err)
	}
	if errors.Is(err, ErrNotModified) {
		t.Errorf("a 409 must not match ErrNotModified: %v", err)
	}
	failure, ok := errors.AsType[*APIError](err)
	if !ok || failure.Code != http.StatusConflict || failure.Method != "getUpdates" {
		t.Errorf("APIError = %+v, %v", failure, ok)
	}
}

func TestUnreadableRepliesMatchErrBadResponse(t *testing.T) {
	cases := []struct {
		name  string
		reply string
	}{
		{"not json", `<html>gateway error</html>`},
		{"result of the wrong shape", `{"ok":true,"result":"not a user"}`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			bot := newFakeBot(t, func(string, string) (int, string) { return http.StatusOK, tc.reply })

			_, err := bot.client().GetMe(t.Context())

			if !errors.Is(err, ErrBadResponse) {
				t.Fatalf("err = %v, want ErrBadResponse", err)
			}
		})
	}
}

func TestUnreachableTelegramMatchesErrTransport(t *testing.T) {
	closed := httptest.NewServer(http.NotFoundHandler())
	closed.Close()

	_, err := New(testToken, WithBaseURL(closed.URL)).GetMe(t.Context())

	if !errors.Is(err, ErrTransport) {
		t.Fatalf("err = %v, want ErrTransport", err)
	}
}

func TestCancelledCallMatchesBothTransportAndContext(t *testing.T) {
	bot := newFakeBot(t, okReply(`{}`))
	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	_, err := bot.client().GetMe(ctx)

	if !errors.Is(err, ErrTransport) || !errors.Is(err, context.Canceled) {
		t.Fatalf("err = %v, want ErrTransport and context.Canceled", err)
	}
}

func TestErrorsNeverCarryTheToken(t *testing.T) {
	closed := httptest.NewServer(http.NotFoundHandler())
	closed.Close()
	client := New(testToken, WithBaseURL(closed.URL))

	_, getMeErr := client.GetMe(t.Context())
	downloadErr := client.Download(t.Context(), "documents/a.txt", 10, io.Discard)

	for name, err := range map[string]error{"GetMe": getMeErr, "Download": downloadErr} {
		if err == nil {
			t.Fatalf("%s: want an error against a closed server", name)
		}
		if strings.Contains(err.Error(), testToken) {
			t.Errorf("%s error leaks the token: %v", name, err)
		}
	}
}
