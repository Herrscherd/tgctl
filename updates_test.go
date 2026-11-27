package tgctl

import (
	"errors"
	"net/http"
	"testing"
	"time"
)

func TestPollSecondsRoundsUpAndBoundsTheWindow(t *testing.T) {
	cases := []struct {
		name    string
		timeout []time.Duration
		want    int
		wantErr error
	}{
		{"default window", nil, 50, nil},
		{"zero returns at once", []time.Duration{0}, 0, nil},
		{"one nanosecond rounds up", []time.Duration{time.Nanosecond}, 1, nil},
		{"whole seconds", []time.Duration{3 * time.Second}, 3, nil},
		{"fraction rounds up", []time.Duration{1500 * time.Millisecond}, 2, nil},
		{"longest window", []time.Duration{50 * time.Second}, 50, nil},
		{"past the longest window", []time.Duration{50*time.Second + time.Nanosecond}, 0, ErrBadTimeout},
		{"negative window", []time.Duration{-time.Second}, 0, ErrBadTimeout},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := pollSeconds(tc.timeout)

			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("err = %v, want %v", err, tc.wantErr)
			}
			if err == nil && got != tc.want {
				t.Errorf("seconds = %d, want %d", got, tc.want)
			}
		})
	}
}

func TestGetUpdatesRefusesAWindowOutsideTheBoundsBeforeAnyRequest(t *testing.T) {
	bot := newFakeBot(t, refusal(http.StatusInternalServerError, "never sent"))

	_, err := bot.client().GetUpdates(t.Context(), 0, 80*time.Second)

	if !errors.Is(err, ErrBadTimeout) {
		t.Fatalf("err = %v, want ErrBadTimeout", err)
	}
	if n := len(bot.recorded()); n != 0 {
		t.Errorf("%d requests sent for a refused window, want none", n)
	}
}
