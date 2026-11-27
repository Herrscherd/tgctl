package tgctl

import (
	"testing"
)

func TestSentinelTextsMatchTheGateway(t *testing.T) {
	cases := []struct {
		name string
		got  error
		want string
	}{
		{"ErrRequest", ErrRequest, "telegram request could not be built"},
		{"ErrTransport", ErrTransport, "telegram unreachable"},
		{"ErrBadResponse", ErrBadResponse, "telegram answered with an unreadable response"},
		{"ErrBadFilePath", ErrBadFilePath, "telegram gave a file path that is not a relative path"},
		{"ErrFileTooLarge", ErrFileTooLarge, "file is bigger than the cap"},
		{"ErrFileStatus", ErrFileStatus, "telegram refused the file download"},
		{"ErrWrite", ErrWrite, "telegram file could not be written to the destination"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.got.Error() != tc.want {
				t.Errorf("%s = %q, want %q", tc.name, tc.got.Error(), tc.want)
			}
		})
	}
}
