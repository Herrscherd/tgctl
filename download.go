package tgctl

import (
	"context"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"strings"
)

// Download copies the file at filePath, a path from [Client.GetFile], into dst.
// At most limit bytes are accepted: a file that holds more returns
// [ErrFileTooLarge], and the bytes read before that point stay in dst. A limit
// that is not positive returns [ErrBadLimit], and a path that is not relative
// returns [ErrBadFilePath]; both before any request is sent. A server answer
// other than 200, a redirect included, returns [ErrFileStatus].
func (c *Client) Download(ctx context.Context, filePath string, limit int64, dst io.Writer) error {
	return c.api.download(ctx, filePath, limit, dst)
}

func (t *transport) download(ctx context.Context, filePath string, limit int64, dst io.Writer) error {
	if limit <= 0 {
		return ErrBadLimit
	}
	if !relativeFilePath(filePath) {
		return ErrBadFilePath
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, t.base+"/file/bot"+t.token+"/"+filePath, http.NoBody)
	if err != nil {
		return ErrRequest
	}
	resp, err := t.client.Do(req)
	if err != nil {
		return t.transportFailure(err)
	}
	copyErr := t.copyCapped(resp, limit, dst)
	return errors.Join(copyErr, resp.Body.Close())
}

func (t *transport) copyCapped(resp *http.Response, limit int64, dst io.Writer) error {
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%w: http %d", ErrFileStatus, resp.StatusCode)
	}
	written, err := io.Copy(dst, io.LimitReader(resp.Body, readBound(limit)))
	if err != nil {
		return t.transportFailure(err)
	}
	if written > limit {
		return ErrFileTooLarge
	}
	return nil
}

func readBound(limit int64) int64 {
	return min(limit, math.MaxInt64-1) + 1
}

func relativeFilePath(filePath string) bool {
	if filePath == "" || strings.HasPrefix(filePath, "/") || strings.ContainsAny(filePath, "?#\\") {
		return false
	}
	for segment := range strings.SplitSeq(filePath, "/") {
		if segment == "" || segment == "." || segment == ".." {
			return false
		}
	}
	return true
}
