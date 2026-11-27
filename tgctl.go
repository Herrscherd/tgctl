// Package tgctl is a Telegram Bot API client, the Telegram counterpart of
// github.com/Herrscherd/dctl. It owns the wire types the Bot API sends and
// takes, and one method per call the herrscher gateways make. It holds no
// state beyond its configuration, runs no long poll loop and logs nothing: the
// caller owns the loop, the logger and the bytes it writes. It depends on the
// standard library only.
//
//	bot := tgctl.New(token)
//	updates, err := bot.GetUpdates(ctx, 0)
//	if err != nil {
//		return err
//	}
//	for _, update := range updates {
//		if update.Message != nil {
//			_, err = bot.SendMessage(ctx, tgctl.OutgoingMessage{
//				ChatID: update.Message.Chat.ID,
//				Text:   "received",
//			})
//		}
//	}
package tgctl

import (
	"net/http"
	"time"
)

const (
	defaultBaseURL = "https://api.telegram.org"
	defaultPoll    = 50 * time.Second
	requestSlack   = 20 * time.Second
)

// Client calls the Bot API of one bot. Build it with [New]. It holds no
// per-call state, so one Client serves concurrent calls.
type Client struct {
	api *transport
}

// ClientOption configures a [Client] built by [New].
type ClientOption func(*clientConfig)

type clientConfig struct {
	baseURL    string
	httpClient *http.Client
}

// WithBaseURL points the client at another Bot API root, such as a local Bot
// API server or a test server. The default is https://api.telegram.org.
func WithBaseURL(baseURL string) ClientOption {
	return func(cfg *clientConfig) { cfg.baseURL = baseURL }
}

// WithHTTPClient replaces the default HTTP client. The default times out after
// 70 s, which covers the default 50 s poll of [Client.GetUpdates]; a
// replacement must allow at least as long for the polls it carries.
func WithHTTPClient(httpClient *http.Client) ClientOption {
	return func(cfg *clientConfig) { cfg.httpClient = httpClient }
}

// New builds a Client for the bot identified by token. The token is kept in
// memory and sent in request paths only; transport errors replace it with
// [redacted] before they are returned.
func New(token string, opts ...ClientOption) *Client {
	cfg := clientConfig{baseURL: defaultBaseURL}
	for _, apply := range opts {
		apply(&cfg)
	}
	if cfg.httpClient == nil {
		cfg.httpClient = &http.Client{Timeout: defaultPoll + requestSlack}
	}
	return &Client{api: newTransport(cfg.baseURL, token, cfg.httpClient)}
}
