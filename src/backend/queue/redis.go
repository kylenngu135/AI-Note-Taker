package queue

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type redisClient struct {
	url   string
	token string
}

func newRedisClient(url, token string) *redisClient {
	return &redisClient{url: url, token: token}
}

// do sends a Redis command encoded as a JSON array to the Upstash REST API.
func (c *redisClient) do(cmd []interface{}) (interface{}, error) {
	body, err := json.Marshal(cmd)
	if err != nil {
		return nil, fmt.Errorf("marshal redis command: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, c.url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create redis request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("redis HTTP: %w", err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read redis response: %w", err)
	}

	var result struct {
		Result interface{} `json:"result"`
		Error  string      `json:"error"`
	}
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, fmt.Errorf("decode redis response: %w", err)
	}
	if result.Error != "" {
		return nil, fmt.Errorf("redis: %s", result.Error)
	}
	return result.Result, nil
}

// lpush pushes value onto the left end of key.
func (c *redisClient) lpush(key, value string) error {
	_, err := c.do([]interface{}{"LPUSH", key, value})
	return err
}

// brpop blocks until an element is available on one of keys, or timeout seconds elapse.
// Returns ("", "", nil) when the timeout fires with no item.
func (c *redisClient) brpop(timeout int, keys ...string) (string, string, error) {
	args := make([]interface{}, 0, len(keys)+2)
	args = append(args, "BRPOP")
	for _, k := range keys {
		args = append(args, k)
	}
	args = append(args, timeout)

	result, err := c.do(args)
	if err != nil {
		return "", "", err
	}
	if result == nil {
		return "", "", nil
	}

	arr, ok := result.([]interface{})
	if !ok || len(arr) != 2 {
		return "", "", fmt.Errorf("unexpected BRPOP result shape: %v", result)
	}
	return arr[0].(string), arr[1].(string), nil
}
