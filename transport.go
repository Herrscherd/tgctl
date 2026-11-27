package tgctl

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const responseMaxBytes = 8 << 20

type apiResponse struct {
	OK          bool            `json:"ok"`
	Result      json.RawMessage `json:"result"`
	ErrorCode   int             `json:"error_code"`
	Description string          `json:"description"`
	Parameters  *struct {
		RetryAfter int `json:"retry_after"`
	} `json:"parameters"`
}

type transport struct {
	base   string
	token  string
	client *http.Client
}

func newTransport(base, token string, client *http.Client) *transport {
	return &transport{base: strings.TrimRight(base, "/"), token: token, client: client}
}

func (t *transport) call(ctx context.Context, method string, params, result any) error {
	body, err := json.Marshal(params)
	if err != nil {
		return fmt.Errorf("telegram %s: encode: %w", method, err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, t.base+"/bot"+t.token+"/"+method, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("telegram %s: %w", method, ErrRequest)
	}
	req.Header.Set("Content-Type", "application/json")
	data, err := t.fetch(req)
	if err != nil {
		return fmt.Errorf("telegram %s: %w", method, err)
	}
	var envelope apiResponse
	if err := json.Unmarshal(data, &envelope); err != nil {
		return fmt.Errorf("telegram %s: %w", method, ErrBadResponse)
	}
	if !envelope.OK {
		return envelope.failure(method)
	}
	if result == nil {
		return nil
	}
	if err := json.Unmarshal(envelope.Result, result); err != nil {
		return fmt.Errorf("telegram %s: %w", method, ErrBadResponse)
	}
	return nil
}

func (t *transport) fetch(req *http.Request) ([]byte, error) {
	resp, err := t.client.Do(req)
	if err != nil {
		return nil, t.transportFailure(err)
	}
	data, err := io.ReadAll(io.LimitReader(resp.Body, responseMaxBytes))
	err = errors.Join(err, resp.Body.Close())
	if err != nil {
		return nil, t.transportFailure(err)
	}
	return data, nil
}

func (t *transport) transportFailure(err error) error {
	var urlErr *url.Error
	if errors.As(err, &urlErr) {
		err = urlErr.Err
	}
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return fmt.Errorf("%w: %w", ErrTransport, err)
	}
	return fmt.Errorf("%w: %s", ErrTransport, t.redact(err.Error()))
}

func (t *transport) redact(text string) string {
	if t.token == "" {
		return text
	}
	return strings.ReplaceAll(text, t.token, "[redacted]")
}

func (r apiResponse) failure(method string) error {
	failure := &APIError{Method: method, Code: r.ErrorCode, Description: r.Description}
	if r.Parameters != nil && r.Parameters.RetryAfter > 0 {
		failure.RetryAfter = time.Duration(r.Parameters.RetryAfter) * time.Second
	}
	return failure
}
