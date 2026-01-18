package gateway

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/furuya-3150/fam-diary-log/pkg/errors"
	httputil "github.com/furuya-3150/fam-diary-log/pkg/http"
)

const (
	KouseiEndpoint = "https://jlp.yahooapis.jp/KouseiService/V1/kousei"
)

type YahooNLPGateway struct {
	appID  string
	client *httputil.Client
}

// NewYahooNLPGateway creates a new YahooNLPGateway
func NewYahooNLPGateway(appID string) *YahooNLPGateway {
	return &YahooNLPGateway{
		appID:  appID,
		client: httputil.NewClient(),
	}
}

type kouseiRequest struct {
	ID      string `json:"id"`
	JSONRPC string `json:"jsonrpc"`
	Method  string `json:"method"`
	Params  struct {
		Q string `json:"q"`
	} `json:"params"`
}

type kouseiResponse struct {
	Result struct {
		Suggestions []struct {
			Offset      int    `json:"offset"`
			Length      int    `json:"length"`
			Message     string `json:"message"`
			Suggestion  string `json:"suggestion"`
			SurfaceForm string `json:"surface_form"`
		} `json:"suggestions"`
	} `json:"result"`
}

// CheckAccuracy checks the accuracy of the given text using Yahoo Proofreading API
// Returns the number of suggestions found (errors)
func (g *YahooNLPGateway) CheckAccuracy(ctx context.Context, text string) (int, error) {
	if g.appID == "" {
		return 0, fmt.Errorf("Yahoo AppID not configured")
	}

	if text == "" {
		return 0, fmt.Errorf("text is empty")
	}

	request := kouseiRequest{
		ID:      "1",
		JSONRPC: "2.0",
		Method:  "jlp.kouseiservice.kousei",
	}
	request.Params.Q = text

	body, err := json.Marshal(request)
	if err != nil {
		return 0, err
	}

	// Create request
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, KouseiEndpoint, bytes.NewReader(body))
	if err != nil {
		return 0, err
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Yahoo AppID: "+g.appID)

	// Execute request with retry logic
	resp, err := g.client.Do(ctx, req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return 0, &errors.ExternalAPIError{
			Message: fmt.Sprintf("yahoo kousei api error: status=%d", resp.StatusCode),
			Cause:   fmt.Errorf("body=%s", string(respBody)),
		}
	}

	// Parse response
	var result kouseiResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, err
	}

	// Return suggestion count only (score calculation is domain responsibility)
	return len(result.Result.Suggestions), nil
}
