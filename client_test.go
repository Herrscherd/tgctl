package tgctl

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"slices"
	"sync"
	"sync/atomic"
	"testing"
)

func TestClientsOfTwoBotsCallTheirOwnPath(t *testing.T) {
	var mu sync.Mutex
	var paths []string
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		mu.Lock()
		paths = append(paths, request.URL.Path)
		mu.Unlock()
		if _, err := writer.Write([]byte(`{"ok":true,"result":{"id":1,"is_bot":true,"first_name":"bot"}}`)); err != nil {
			t.Errorf("write the reply: %v", err)
		}
	}))
	t.Cleanup(server.Close)
	first := New("111:first-token", WithBaseURL(server.URL))
	second := New("222:second-token", WithBaseURL(server.URL))

	if _, err := first.GetMe(t.Context()); err != nil {
		t.Fatal(err)
	}
	if _, err := second.GetMe(t.Context()); err != nil {
		t.Fatal(err)
	}

	want := []string{"/bot111:first-token/getMe", "/bot222:second-token/getMe"}
	if !slices.Equal(paths, want) {
		t.Errorf("paths = %v, want %v", paths, want)
	}
}

func TestDefaultClientRefusesRedirectsSoTheTokenStaysHome(t *testing.T) {
	var followed atomic.Int32
	target := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		followed.Add(1)
	}))
	t.Cleanup(target.Close)
	redirect := httptest.NewServer(http.RedirectHandler(target.URL+"/bot"+testToken+"/getMe", http.StatusTemporaryRedirect))
	t.Cleanup(redirect.Close)

	_, err := New(testToken, WithBaseURL(redirect.URL)).GetMe(t.Context())

	if !errors.Is(err, ErrBadResponse) {
		t.Fatalf("err = %v, want ErrBadResponse for the refused 307", err)
	}
	if n := followed.Load(); n != 0 {
		t.Errorf("the redirect target was called %d times, want none", n)
	}
}
