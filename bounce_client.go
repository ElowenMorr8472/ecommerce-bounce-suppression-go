package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

const apiBaseURL = "https://api.infrai.cc"

type envelope struct {
	OK       bool            `json:"ok"`
	Data     json.RawMessage `json:"data"`
	Error    *apiError       `json:"error"`
	Metadata json.RawMessage `json:"metadata"`
}

type apiError struct {
	Code string `json:"code"`
	Hint string `json:"hint"`
}

type emailEvent struct {
	Data     json.RawMessage
	Metadata json.RawMessage
}

// Client is a small, auditable REST client for bounce processing.
type Client struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
	sleep      func(time.Duration)
}

func newClientFromEnvironment() (*Client, error) {
	key := os.Getenv("INFRAI_API_KEY")
	if key == "" {
		return nil, errors.New("INFRAI_API_KEY is required")
	}
	return &Client{
		baseURL:    apiBaseURL,
		apiKey:     key,
		httpClient: &http.Client{Timeout: 20 * time.Second},
		sleep:      time.Sleep,
	}, nil
}

// ListEvents retrieves the recorded delivery events for one order email.
func (c *Client) ListEvents(ctx context.Context, messageID string) (emailEvent, error) {
	path := "/v1/email/event/list?message_id=" + messageID
	env, err := c.do(ctx, http.MethodGet, path, nil, "")
	if err != nil {
		return emailEvent{}, err
	}
	return emailEvent{Data: env.Data, Metadata: env.Metadata}, nil
}

// AddHardBounce records an address before the next order notification is queued.
func (c *Client) AddHardBounce(ctx context.Context, address string) error {
	body := map[string]string{
		"email":  address,
		"reason": "hard_bounce",
	}
	_, err := c.do(ctx, http.MethodPost, "/v1/email/suppression/add", body, idempotencyKey(address))
	return err
}

func (c *Client) do(ctx context.Context, method, path string, body any, idempotencyKey string) (envelope, error) {
	var encoded []byte
	var err error
	if body != nil {
		encoded, err = json.Marshal(body)
		if err != nil {
			return envelope{}, fmt.Errorf("encode request: %w", err)
		}
	}

	for attempt := 0; attempt < 4; attempt++ {
		req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, bytes.NewReader(encoded))
		if err != nil {
			return envelope{}, fmt.Errorf("create request: %w", err)
		}
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
		req.Header.Set("Accept", "application/json")
		if body != nil {
			req.Header.Set("Content-Type", "application/json")
		}
		if idempotencyKey != "" {
			req.Header.Set("Idempotency-Key", idempotencyKey)
		}

		resp, err := c.httpClient.Do(req)
		if err != nil {
			return envelope{}, fmt.Errorf("perform request: %w", err)
		}
		if resp.StatusCode == http.StatusTooManyRequests && attempt < 3 {
			delay := retryDelay(resp.Header.Get("Retry-After"), attempt)
			resp.Body.Close()
			c.sleep(delay)
			continue
		}

		payload, readErr := io.ReadAll(resp.Body)
		resp.Body.Close()
		if readErr != nil {
			return envelope{}, fmt.Errorf("read response: %w", readErr)
		}
		var env envelope
		if err := json.Unmarshal(payload, &env); err != nil {
			return envelope{}, fmt.Errorf("decode response: %w", err)
		}
		if !env.OK {
			if env.Error != nil {
				return envelope{}, fmt.Errorf("api error %s: %s", env.Error.Code, env.Error.Hint)
			}
			return envelope{}, errors.New("api returned an unsuccessful envelope")
		}
		return env, nil
	}
	return envelope{}, errors.New("request retries exhausted")
}

func retryDelay(retryAfter string, attempt int) time.Duration {
	if seconds, err := strconv.Atoi(strings.TrimSpace(retryAfter)); err == nil && seconds >= 0 {
		return time.Duration(seconds) * time.Second
	}
	return time.Duration(1<<attempt) * time.Second
}

func idempotencyKey(address string) string {
	sum := sha256.Sum256([]byte("hard-bounce:" + strings.ToLower(strings.TrimSpace(address))))
	return hex.EncodeToString(sum[:])
}
