package tgctl

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const notModifiedText = "message is not modified"

// ErrTransport is returned when Telegram cannot be reached, or the connection
// breaks before a complete reply. A context that ends during the call wraps
// the context error as well, so [context.Canceled] and
// [context.DeadlineExceeded] still match.
var ErrTransport = errors.New("telegram unreachable")

// ErrRequest is returned when a request cannot be built from the client
// configuration, for instance a base URL that does not parse.
var ErrRequest = errors.New("telegram request could not be built")

// ErrBadResponse is returned when Telegram answers with a body that is not the
// expected JSON envelope or result.
var ErrBadResponse = errors.New("telegram answered with an unreadable response")

// ErrConflict matches an [APIError] with status 409: another caller polls the
// same bot with [Client.GetUpdates], or a webhook is set on it.
var ErrConflict = errors.New("telegram conflict: another poller or a webhook holds the bot")

// ErrNotModified matches an [APIError] for an edit that would leave the message
// unchanged. Callers that retry an edit treat it as success.
var ErrNotModified = errors.New("telegram message is not modified")

// ErrBadFilePath is returned by [Client.Download] for a path that is empty, is
// absolute, holds a query or fragment marker, a backslash, an empty segment, or
// a dot segment, before any request is sent.
var ErrBadFilePath = errors.New("telegram file path is not a relative path")

// ErrFileTooLarge is returned by [Client.Download] when the file holds more
// bytes than the limit. The bytes up to the limit have already been written to
// the destination.
var ErrFileTooLarge = errors.New("file is bigger than the cap")

// ErrFileStatus is returned by [Client.Download] when the file server answers
// with a status other than 200.
var ErrFileStatus = errors.New("telegram refused the file download")

// APIError is a refusal the Bot API returns in its JSON envelope with ok set to
// false. Code is the error_code of the reply and Description its text. RetryAfter
// is the delay a flood refusal asks for, and zero when it asks for none. Method
// is the Bot API method that was refused.
type APIError struct {
	Method      string
	Code        int
	Description string
	RetryAfter  time.Duration
}

func (e *APIError) Error() string {
	return "telegram " + e.Method + ": " + strconv.Itoa(e.Code) + " " + e.Description
}

// Is makes an APIError match [ErrConflict] when its code is 409, and
// [ErrNotModified] when it is the 400 that reports an unchanged message.
func (e *APIError) Is(target error) bool {
	switch {
	case errors.Is(target, ErrConflict):
		return e.Code == http.StatusConflict
	case errors.Is(target, ErrNotModified):
		return e.Code == http.StatusBadRequest && strings.Contains(e.Description, notModifiedText)
	default:
		return false
	}
}

// RetryAfter returns the delay a flood refusal asked for, and whether err holds
// one. A caller waits that long before it retries the call.
func RetryAfter(err error) (time.Duration, bool) {
	failure, ok := errors.AsType[*APIError](err)
	if !ok || failure.RetryAfter <= 0 {
		return 0, false
	}
	return failure.RetryAfter, true
}
