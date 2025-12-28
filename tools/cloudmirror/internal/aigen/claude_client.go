package aigen

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"
)

// ClaudeClient is a client for the Anthropic Claude API.
type ClaudeClient struct {
	apiKey     string
	model      string
	httpClient *http.Client
	apiURL     string
}

// ClaudeRequest represents a request to the Claude API.
type ClaudeRequest struct {
	Model       string          `json:"model"`
	MaxTokens   int             `json:"max_tokens"`
	Messages    []ClaudeMessage `json:"messages"`
	System      string          `json:"system,omitempty"`
	Temperature float64         `json:"temperature,omitempty"`
}

// ClaudeMessage represents a message in the Claude API.
type ClaudeMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ClaudeResponse represents a response from the Claude API.
type ClaudeResponse struct {
	ID           string         `json:"id"`
	Type         string         `json:"type"`
	Role         string         `json:"role"`
	Content      []ContentBlock `json:"content"`
	Model        string         `json:"model"`
	StopReason   string         `json:"stop_reason"`
	StopSequence string         `json:"stop_sequence,omitempty"`
	Usage        ClaudeUsage    `json:"usage"`
}

// ContentBlock represents a content block in the response.
type ContentBlock struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// ClaudeUsage represents token usage.
type ClaudeUsage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

// ClaudeError represents an error from the Claude API.
type ClaudeError struct {
	Type    string `json:"type"`
	Error   struct {
		Type    string `json:"type"`
		Message string `json:"message"`
	} `json:"error"`
}

// NewClaudeClient creates a new Claude API client.
func NewClaudeClient(apiKey, model string) *ClaudeClient {
	return &ClaudeClient{
		apiKey: apiKey,
		model:  model,
		httpClient: &http.Client{
			Timeout: 5 * time.Minute, // Long timeout for code generation
		},
		apiURL: "https://api.anthropic.com/v1/messages",
	}
}

// Generate generates code using the Claude API.
func (c *ClaudeClient) Generate(ctx context.Context, prompt string) (string, error) {
	return c.GenerateWithSystem(ctx, "", prompt)
}

// GenerateWithSystem generates code using the Claude API with a system prompt.
func (c *ClaudeClient) GenerateWithSystem(ctx context.Context, systemPrompt, userPrompt string) (string, error) {
	req := ClaudeRequest{
		Model:     c.model,
		MaxTokens: 8192,
		Messages: []ClaudeMessage{
			{Role: "user", Content: userPrompt},
		},
		Temperature: 0.1, // Low temperature for consistent code generation
	}

	if systemPrompt != "" {
		req.System = systemPrompt
	}

	reqBody, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.apiURL, bytes.NewReader(reqBody))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", c.apiKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var apiErr ClaudeError
		if err := json.Unmarshal(body, &apiErr); err == nil && apiErr.Error.Message != "" {
			return "", fmt.Errorf("Claude API error: %s", apiErr.Error.Message)
		}
		return "", fmt.Errorf("Claude API returned status %d: %s", resp.StatusCode, string(body))
	}

	var claudeResp ClaudeResponse
	if err := json.Unmarshal(body, &claudeResp); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	// Extract text content
	var result strings.Builder
	for _, block := range claudeResp.Content {
		if block.Type == "text" {
			result.WriteString(block.Text)
		}
	}

	// Extract code from markdown code blocks if present
	text := result.String()
	code := extractCodeFromMarkdown(text)
	if code != "" {
		return code, nil
	}

	return text, nil
}

// extractCodeFromMarkdown extracts Go code from markdown code blocks.
func extractCodeFromMarkdown(text string) string {
	// Pattern to match ```go ... ``` or ``` ... ```
	re := regexp.MustCompile("(?s)```(?:go)?\\s*\\n(.+?)```")
	matches := re.FindAllStringSubmatch(text, -1)

	if len(matches) == 0 {
		return ""
	}

	// Combine all code blocks
	var code strings.Builder
	for i, match := range matches {
		if len(match) > 1 {
			if i > 0 {
				code.WriteString("\n\n")
			}
			code.WriteString(strings.TrimSpace(match[1]))
		}
	}

	return code.String()
}

// GenerateHandler generates a handler implementation.
func (c *ClaudeClient) GenerateHandler(ctx context.Context, prompt string) (*GeneratedCode, error) {
	systemPrompt := `You are an expert Go developer implementing AWS service handlers for the InfraSpec API emulator.
Follow these patterns exactly:
- Use the service's successResponse() and errorResponse() helpers
- Create Result types with XMLName for Query protocol
- Store state using consistent key patterns: <service>:<resource-type>:<id>
- Validate required parameters before processing
- Return AWS-compatible error codes

Output ONLY the Go code, wrapped in a single markdown code block. Do not include any explanations.`

	code, err := c.GenerateWithSystem(ctx, systemPrompt, prompt)
	if err != nil {
		return nil, err
	}

	return &GeneratedCode{
		HandlerCode: code,
	}, nil
}

// GenerateTests generates tests for a handler.
func (c *ClaudeClient) GenerateTests(ctx context.Context, prompt string) (string, error) {
	systemPrompt := `You are an expert Go developer writing tests for AWS service handlers.
Follow these patterns:
- Use testing.T and testify assertions
- Create helper functions for common setup
- Test success cases, missing parameters, and resource not found
- Use the emulator.AWSRequest structure for test inputs

Output ONLY the Go test code, wrapped in a single markdown code block. Do not include any explanations.`

	return c.GenerateWithSystem(ctx, systemPrompt, prompt)
}
