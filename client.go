package swipegames

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
)

// Client is the Swipe Games SDK client.
type Client struct {
	cid               string
	extCID            string
	apiKey            string
	integrationAPIKey string
	baseURL           string
	debug             bool
	httpClient        *http.Client
}

// NewClient creates a new Swipe Games client.
func NewClient(config ClientConfig) (*Client, error) {
	if config.CID == "" {
		return nil, &ValidationError{Message: "CID is required", Field: "CID"}
	}
	if config.ExtCID == "" {
		return nil, &ValidationError{Message: "ExtCID is required", Field: "ExtCID"}
	}
	if config.APIKey == "" {
		return nil, &ValidationError{Message: "APIKey is required", Field: "APIKey"}
	}
	if config.IntegrationAPIKey == "" {
		return nil, &ValidationError{Message: "IntegrationAPIKey is required", Field: "IntegrationAPIKey"}
	}

	baseURL := config.BaseURL
	if baseURL == "" {
		env := config.Env
		if env == "" {
			env = EnvStaging
		}
		u, ok := envURLs[env]
		if !ok {
			return nil, fmt.Errorf("unknown env: %s", string(env))
		}
		baseURL = u
	}

	httpClient := config.HTTPClient
	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	return &Client{
		cid:               config.CID,
		extCID:            config.ExtCID,
		apiKey:            config.APIKey,
		integrationAPIKey: config.IntegrationAPIKey,
		baseURL:           baseURL,
		debug:             config.Debug,
		httpClient:        httpClient,
	}, nil
}

func (c *Client) log(format string, args ...interface{}) {
	if c.debug {
		log.Printf(fmt.Sprintf("[SwipeGamesSDK] %s", format), args...)
	}
}

func (c *Client) logError(format string, args ...interface{}) {
	if c.debug {
		log.Printf(fmt.Sprintf("[SwipeGamesSDK] ERROR: %s", format), args...)
	}
}

// CreateNewGame creates a new game session and returns the launcher URL.
func (c *Client) CreateNewGame(ctx context.Context, params CreateNewGameParams) (*CreateNewGameResponse, error) {
	body := c.buildCreateNewGameBody(params)

	var result CreateNewGameResponse
	if err := c.doPost(ctx, "/create-new-game", body, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetGames returns information about all supported games.
func (c *Client) GetGames(ctx context.Context) ([]GameInfo, error) {
	queryParams := map[string]string{
		"cID":    c.cid,
		"extCID": c.extCID,
	}

	var result []GameInfo
	if err := c.doGet(ctx, "/games", queryParams, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// CreateFreeRounds creates a new free rounds campaign.
func (c *Client) CreateFreeRounds(ctx context.Context, params CreateFreeRoundsParams) (*CreateFreeRoundsResponse, error) {
	body := c.buildCreateFreeRoundsBody(params)

	var result CreateFreeRoundsResponse
	if err := c.doPost(ctx, "/free-rounds", body, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// CancelFreeRounds cancels an existing free rounds campaign.
func (c *Client) CancelFreeRounds(ctx context.Context, params CancelFreeRoundsParams) error {
	if params.ID == "" && params.ExtID == "" {
		return &ValidationError{Message: "One of id or extID must be provided"}
	}

	body := map[string]interface{}{
		"cID":    c.cid,
		"extCID": c.extCID,
	}
	if params.ID != "" {
		body["id"] = params.ID
	}
	if params.ExtID != "" {
		body["extID"] = params.ExtID
	}

	return c.doDelete(ctx, "/free-rounds", body)
}

func (c *Client) buildCreateNewGameBody(params CreateNewGameParams) map[string]interface{} {
	body := map[string]interface{}{
		"cID":      c.cid,
		"extCID":   c.extCID,
		"gameID":   params.GameID,
		"demo":     params.Demo,
		"platform": params.Platform,
		"currency": params.Currency,
		"locale":   params.Locale,
	}
	if params.SessionID != "" {
		body["sessionID"] = params.SessionID
	}
	if params.ReturnURL != "" {
		body["returnURL"] = params.ReturnURL
	}
	if params.DepositURL != "" {
		body["depositURL"] = params.DepositURL
	}
	if params.InitDemoBalance != "" {
		body["initDemoBalance"] = params.InitDemoBalance
	}
	if params.User != nil {
		body["user"] = params.User
	}
	return body
}

func (c *Client) buildCreateFreeRoundsBody(params CreateFreeRoundsParams) map[string]interface{} {
	body := map[string]interface{}{
		"cID":       c.cid,
		"extCID":    c.extCID,
		"extID":     params.ExtID,
		"currency":  params.Currency,
		"quantity":  params.Quantity,
		"betLine":   params.BetLine,
		"validFrom": params.ValidFrom,
	}
	if len(params.GameIDs) > 0 {
		body["gameIDs"] = params.GameIDs
	}
	if len(params.UserIDs) > 0 {
		body["userIDs"] = params.UserIDs
	}
	if params.ValidUntil != "" {
		body["validUntil"] = params.ValidUntil
	}
	return body
}

func (c *Client) doPost(ctx context.Context, path string, body map[string]interface{}, result interface{}) error {
	resp, err := c.doRequest(ctx, http.MethodPost, path, body)
	if err != nil {
		return err
	}
	defer func() {
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
	}()

	return json.NewDecoder(resp.Body).Decode(result)
}

func (c *Client) doGet(ctx context.Context, path string, queryParams map[string]string, result interface{}) error {
	u, err := url.Parse(c.baseURL + path)
	if err != nil {
		return fmt.Errorf("failed to parse URL: %w", err)
	}

	q := u.Query()
	for k, v := range queryParams {
		q.Set(k, v)
	}
	u.RawQuery = q.Encode()

	sig, err := createQueryParamsSignature(queryParams, c.apiKey)
	if err != nil {
		return fmt.Errorf("failed to sign query params: %w", err)
	}

	fullURL := u.String()
	c.log("GET %s", fullURL)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fullURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("X-REQUEST-SIGN", sig)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer func() {
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
	}()

	c.log("GET %s -> %d", fullURL, resp.StatusCode)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return c.parseAPIError(resp, fmt.Sprintf("GET %s", fullURL))
	}

	return json.NewDecoder(resp.Body).Decode(result)
}

func (c *Client) doDelete(ctx context.Context, path string, body map[string]interface{}) error {
	resp, err := c.doRequest(ctx, http.MethodDelete, path, body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// read and discard body
	io.Copy(io.Discard, resp.Body)
	return nil
}

func (c *Client) doRequest(ctx context.Context, method, path string, body map[string]interface{}) (*http.Response, error) {
	fullURL := c.baseURL + path

	canonical, err := canonicalizeJSON(body)
	if err != nil {
		return nil, fmt.Errorf("failed to canonicalize body: %w", err)
	}

	sig, err := createSignatureFromCanonical(canonical, c.apiKey)
	if err != nil {
		return nil, fmt.Errorf("failed to sign request: %w", err)
	}

	c.log("%s %s", method, fullURL)
	c.log("Body: %s", canonical)

	req, err := http.NewRequestWithContext(ctx, method, fullURL, bytes.NewReader([]byte(canonical)))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-REQUEST-SIGN", sig)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	c.log("%s %s -> %d", method, fullURL, resp.StatusCode)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		defer func() {
			io.Copy(io.Discard, resp.Body)
			resp.Body.Close()
		}()
		return nil, c.parseAPIError(resp, fmt.Sprintf("%s %s", method, fullURL))
	}

	return resp, nil
}

func (c *Client) parseAPIError(resp *http.Response, label string) error {
	var errBody ErrorResponse
	if err := json.NewDecoder(resp.Body).Decode(&errBody); err != nil {
		errBody = ErrorResponse{Message: resp.Status}
	}
	c.logError("%s error: %+v", label, errBody)
	apiErr := &APIError{
		StatusCode: resp.StatusCode,
		Message:    errBody.Message,
	}
	if errBody.Code != nil {
		apiErr.Code = string(*errBody.Code)
	}
	if errBody.Details != nil {
		apiErr.Details = *errBody.Details
	}
	return apiErr
}
