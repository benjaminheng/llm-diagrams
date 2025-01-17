package anthropic

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/pkg/errors"
)

const (
	baseURL             = "https://api.anthropic.com/v1"
	ModelClaude35Sonnet = "claude-3-5-sonnet-20241022"
	ModelClaude35Haiku  = "claude-3-5-haiku-20241022"
)

// Client represents an Anthropic API client
type Client struct {
	apiKey     string
	httpClient *http.Client
}

// New creates a new Anthropic API client
func New(apiKey string) *Client {
	return &Client{
		apiKey:     apiKey,
		httpClient: &http.Client{},
	}
}

// MessageRequest represents the request body for the /v1/messages API
type MessageRequest struct {
	Model       string      `json:"model"`
	Messages    []Message   `json:"messages"`
	MaxTokens   int         `json:"max_tokens"`
	Temperature float64     `json:"temperature,omitempty"`
	Tools       []Tool      `json:"tools,omitempty"`
	ToolChoice  *ToolChoice `json:"tool_choice,omitempty"`
}

// Message represents a single message in the conversation
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// Tool represents a tool that can be used by the model
type Tool struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	InputSchema json.RawMessage `json:"input_schema"`
}

type ToolChoice struct {
	Type string `json:"type"` // auto, any, tool
	Name string `json:"name"` // set if type=tool, force the use of this tool
}

// MessageResponse represents the response from the /v1/messages API
type MessageResponse struct {
	ID         string    `json:"id"`
	Type       string    `json:"type"`
	Role       string    `json:"role"`
	Content    []Content `json:"content"`
	Model      string    `json:"model"`
	StopReason string    `json:"stop_reason"`
	Usage      Usage     `json:"usage"`
}

// Content represents the content of a message in the response
type Content struct {
	Type string `json:"type"`
	Text string `json:"text"`

	// Fields when type=tool_use
	ID    string          `json:"id"`
	Name  string          `json:"name"`
	Input json.RawMessage `json:"input"`
}

// Usage represents the token usage information
type Usage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

// CreateMessage sends a request to the /v1/messages API
func (c *Client) CreateMessage(req MessageRequest) (*MessageResponse, error) {
	url := fmt.Sprintf("%s/messages", baseURL)

	jsonBody, err := json.Marshal(req)
	if err != nil {
		return nil, errors.Wrap(err, "marshal request")
	}

	httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, errors.Wrap(err, "create request")
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-Api-Key", c.apiKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, errors.Wrap(err, "send request")
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "read response body")
	}

	if resp.StatusCode != http.StatusOK {
		return nil, errors.Errorf("API request failed with status code %d: %s", resp.StatusCode, string(body))
	}

	var messageResp MessageResponse
	err = json.Unmarshal(body, &messageResp)
	if err != nil {
		return nil, errors.Wrap(err, "unmarshaling response")
	}

	return &messageResp, nil
}
