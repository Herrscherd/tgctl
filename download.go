package tgctl

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// Download copies the file at filePath, a path from [Client.GetFile], into dst.
// At most limit bytes are accepted: a file that holds more returns
// [ErrFileTooLarge], and the first limit bytes stay in dst. A limit that is not
// positive returns [ErrBadLimit], and a path that is not relative returns
// [ErrBadFilePath]; both before any request is sent. A server answer other than
// 200, a redirect included, returns [ErrFileStatus]. A body that breaks while it
// is read returns [ErrTransport], and a dst that refuses bytes returns [ErrWrite].
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
	sink := &recordingWriter{dst: dst}
	_, err := io.Copy(sink, io.LimitReader(resp.Body, limit))
	if sink.err != nil {
		return fmt.Errorf("%w: %w", ErrWrite, sink.err)
	}
	if err != nil {
		return t.transportFailure(err)
	}
	return t.rejectExtraByte(resp.Body)
}

func (t *transport) rejectExtraByte(body io.Reader) error {
	var probe [1]byte
	count, err := io.ReadFull(body, probe[:])
	switch {
	case count == 1:
		return ErrFileTooLarge
	case errors.Is(err, io.EOF):
		return nil
	default:
		return t.transportFailure(err)
	}
}

type recordingWriter struct {
	dst io.Writer
	err error
}

func (w *recordingWriter) Write(chunk []byte) (int, error) {
	written, err := w.dst.Write(chunk)
	if err == nil && written < len(chunk) {
		err = io.ErrShortWrite
	}
	if err != nil {
		w.err = err
	}
	return written, err
}

func relativeFilePath(filePath string) bool {
	if filePath == "" || strings.HasPrefix(filePath, "/") || strings.ContainsAny(filePath, "?#\\%") {
		return false
	}
	for segment := range strings.SplitSeq(filePath, "/") {
		if segment == "" || segment == "." || segment == ".." {
			return false
		}
	}
	return true
}
