package tgctl

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync"
	"testing"
)

const testToken = "123456:secret-token"

type recordedRequest struct {
	path string
	body string
}

type fakeBot struct {
	t      *testing.T
	server *httptest.Server
	reply  func(method, body string) (int, string)

	mu       sync.Mutex
	requests []recordedRequest
	files    map[string]string
}

func newFakeBot(t *testing.T, reply func(method, body string) (int, string)) *fakeBot {
	t.Helper()
	bot := &fakeBot{t: t, reply: reply, files: make(map[string]string)}
	bot.server = httptest.NewServer(http.HandlerFunc(bot.serve))
	t.Cleanup(bot.server.Close)
	return bot
}

func (f *fakeBot) client() *Client {
	return New(testToken, WithBaseURL(f.server.URL))
}

func (f *fakeBot) recorded() []recordedRequest {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]recordedRequest(nil), f.requests...)
}

func (f *fakeBot) serve(writer http.ResponseWriter, request *http.Request) {
	if path, ok := strings.CutPrefix(request.URL.Path, "/file/bot"+testToken+"/"); ok {
		f.serveFile(writer, path)
		return
	}
	method, ok := strings.CutPrefix(request.URL.Path, "/bot"+testToken+"/")
	if !ok {
		writer.WriteHeader(http.StatusNotFound)
		return
	}
	body, err := io.ReadAll(request.Body)
	if err != nil {
		f.t.Errorf("read the request: %v", err)
	}
	f.mu.Lock()
	f.requests = append(f.requests, recordedRequest{path: request.URL.Path, body: string(body)})
	f.mu.Unlock()
	status, answer := f.reply(method, string(body))
	writer.WriteHeader(status)
	if _, err := io.WriteString(writer, answer); err != nil {
		f.t.Errorf("write the reply: %v", err)
	}
}

func (f *fakeBot) serveFile(writer http.ResponseWriter, path string) {
	f.mu.Lock()
	content, found := f.files[path]
	f.mu.Unlock()
	if !found {
		writer.WriteHeader(http.StatusNotFound)
		return
	}
	if _, err := io.WriteString(writer, content); err != nil {
		f.t.Errorf("write the file: %v", err)
	}
}

func okReply(result string) func(method, body string) (int, string) {
	return func(string, string) (int, string) {
		return http.StatusOK, `{"ok":true,"result":` + result + `}`
	}
}

func refusal(status int, description string) func(method, body string) (int, string) {
	return func(string, string) (int, string) {
		return status, `{"ok":false,"error_code":` + strconv.Itoa(status) + `,"description":"` + description + `"}`
	}
}
