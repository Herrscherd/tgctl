package tgctl

import (
	"bytes"
	"errors"
	"testing"
)

func TestDownloadCopiesTheFileWithinTheLimit(t *testing.T) {
	const content = "0123456789"
	cases := []struct {
		name    string
		limit   int64
		wantErr error
	}{
		{"exactly at the limit", int64(len(content)), nil},
		{"one byte over the limit", int64(len(content)) - 1, ErrFileTooLarge},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			bot := newFakeBot(t, okReply(`true`))
			bot.files["documents/a.txt"] = content
			var dst bytes.Buffer

			err := bot.client().Download(t.Context(), "documents/a.txt", tc.limit, &dst)

			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("err = %v, want %v", err, tc.wantErr)
			}
			if tc.wantErr == nil && dst.String() != content {
				t.Errorf("dst = %q, want %q", dst.String(), content)
			}
		})
	}
}

func TestDownloadRefusesPathsOutsideTheFileTree(t *testing.T) {
	bot := newFakeBot(t, okReply(`true`))

	for _, path := range []string{"", "/etc/passwd", "documents/../../x", "a//b", "a?b", "a#b", `a\b`, "./a"} {
		if err := bot.client().Download(t.Context(), path, 1, &bytes.Buffer{}); !errors.Is(err, ErrBadFilePath) {
			t.Errorf("Download(%q) = %v, want ErrBadFilePath", path, err)
		}
	}
	if n := len(bot.recorded()); n != 0 {
		t.Errorf("%d requests sent for refused paths, want none", n)
	}
}

func TestDownloadMissingFileMatchesErrFileStatus(t *testing.T) {
	bot := newFakeBot(t, okReply(`true`))

	err := bot.client().Download(t.Context(), "documents/gone.txt", 10, &bytes.Buffer{})

	if !errors.Is(err, ErrFileStatus) {
		t.Fatalf("err = %v, want ErrFileStatus", err)
	}
}
