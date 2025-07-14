package proxy

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// ModelsHandler handles /v1/models requests for OpenAI compatibility
type ModelsHandler struct {
	tokenProvider TokenProvider
	upstreamURL   string
	httpClient    *http.Client
}

// NewModelsHandler creates a new models handler
func NewModelsHandler(tokenProvider TokenProvider, upstreamURL string) *ModelsHandler {
	return &ModelsHandler{
		tokenProvider: tokenProvider,
		upstreamURL:   upstreamURL,
		httpClient:    &http.Client{Timeout: 30 * time.Second},
	}
}

// ServeHTTP handles the models endpoint
func (h *ModelsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Handle CORS
	if r.Method == "OPTIONS" {
		setCORSHeadersStandalone(w, r)
		w.WriteHeader(http.StatusNoContent)
		return
	}
	
	setCORSHeadersStandalone(w, r)
	
	// Anthropic's /v1/models endpoint doesn't support OAuth authentication
	// So we use a comprehensive static list of OAuth-accessible models
	models := h.getOAuthModels()
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(models)
}

// fetchModelsFromAnthropic fetches available models from Anthropic's API
func (h *ModelsHandler) fetchModelsFromAnthropic() (map[string]interface{}, error) {
	// Get access token
	accessToken, err := h.tokenProvider.GetAccessToken()
	if err != nil {
		return nil, fmt.Errorf("failed to get access token: %w", err)
	}
	
	// Create request to Anthropic's models endpoint
	req, err := http.NewRequest("GET", h.upstreamURL+"/v1/models", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	// Set headers
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))
	req.Header.Set("anthropic-version", "2023-06-01")
	req.Header.Set("anthropic-beta", "oauth-2025-04-20")
	req.Header.Set("Content-Type", "application/json")
	
	// Make request
	resp, err := h.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}
	
	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}
	
	// Parse Anthropic response
	var anthropicResponse map[string]interface{}
	if err := json.Unmarshal(body, &anthropicResponse); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	
	// Convert to OpenAI format
	return h.convertAnthropicModelsToOpenAI(anthropicResponse), nil
}

// convertAnthropicModelsToOpenAI converts Anthropic models response to OpenAI format
func (h *ModelsHandler) convertAnthropicModelsToOpenAI(anthropicResponse map[string]interface{}) map[string]interface{} {
	openAIModels := map[string]interface{}{
		"object": "list",
		"data":   []interface{}{},
	}
	
	// Extract models from Anthropic response
	if data, ok := anthropicResponse["data"].([]interface{}); ok {
		var models []interface{}
		
		for _, item := range data {
			if model, ok := item.(map[string]interface{}); ok {
				if modelID, ok := model["id"].(string); ok {
					// Convert to OpenAI format
					openAIModel := map[string]interface{}{
						"id":       modelID,
						"object":   "model",
						"created":  int(time.Now().Unix()),
						"owned_by": "anthropic",
						"permission": []interface{}{
							map[string]interface{}{
								"allow_create_engine":  false,
								"allow_fine_tuning":    false,
								"allow_logprobs":       false,
								"allow_sampling":       true,
								"allow_search_indices": false,
								"allow_view":           true,
								"created":              int(time.Now().Unix()),
								"group":                nil,
								"id":                   "modelperm-" + modelID,
								"is_blocking":          false,
								"object":               "model_permission",
								"organization":         "*",
							},
						},
					}
					models = append(models, openAIModel)
				}
			}
		}
		
		openAIModels["data"] = models
	}
	
	return openAIModels
}

// getOAuthModels returns comprehensive list of OAuth-accessible models
func (h *ModelsHandler) getOAuthModels() map[string]interface{} {
	return map[string]interface{}{
		"object": "list",
		"data": []interface{}{
			// Claude 4 Series (Latest)
			map[string]interface{}{
				"id":       "claude-opus-4-20250514",
				"object":   "model",
				"created":  1747353600, // 2025-05-14
				"owned_by": "anthropic",
				"permission": []interface{}{
					map[string]interface{}{
						"allow_create_engine":  false,
						"allow_fine_tuning":    false,
						"allow_logprobs":       false,
						"allow_sampling":       true,
						"allow_search_indices": false,
						"allow_view":           true,
						"created":              int(time.Now().Unix()),
						"group":                nil,
						"id":                   "modelperm-claude-opus-4-20250514",
						"is_blocking":          false,
						"object":               "model_permission",
						"organization":         "*",
					},
				},
			},
			map[string]interface{}{
				"id":       "claude-sonnet-4-20250514",
				"object":   "model",
				"created":  1747353600, // 2025-05-14
				"owned_by": "anthropic",
				"permission": []interface{}{
					map[string]interface{}{
						"allow_create_engine":  false,
						"allow_fine_tuning":    false,
						"allow_logprobs":       false,
						"allow_sampling":       true,
						"allow_search_indices": false,
						"allow_view":           true,
						"created":              int(time.Now().Unix()),
						"group":                nil,
						"id":                   "modelperm-claude-sonnet-4-20250514",
						"is_blocking":          false,
						"object":               "model_permission",
						"organization":         "*",
					},
				},
			},
			// Claude 3.7 Series
			map[string]interface{}{
				"id":       "claude-3-7-sonnet-20250219",
				"object":   "model",
				"created":  1740009600, // 2025-02-19
				"owned_by": "anthropic",
				"permission": []interface{}{
					map[string]interface{}{
						"allow_create_engine":  false,
						"allow_fine_tuning":    false,
						"allow_logprobs":       false,
						"allow_sampling":       true,
						"allow_search_indices": false,
						"allow_view":           true,
						"created":              int(time.Now().Unix()),
						"group":                nil,
						"id":                   "modelperm-claude-3-7-sonnet-20250219",
						"is_blocking":          false,
						"object":               "model_permission",
						"organization":         "*",
					},
				},
			},
			// Claude 3.5 Sonnet (Latest)
			map[string]interface{}{
				"id":       "claude-3-5-sonnet-20241022",
				"object":   "model",
				"created":  1729555200, // 2024-10-22
				"owned_by": "anthropic",
				"permission": []interface{}{
					map[string]interface{}{
						"allow_create_engine":  false,
						"allow_fine_tuning":    false,
						"allow_logprobs":       false,
						"allow_sampling":       true,
						"allow_search_indices": false,
						"allow_view":           true,
						"created":              int(time.Now().Unix()),
						"group":                nil,
						"id":                   "modelperm-claude-3-5-sonnet-20241022",
						"is_blocking":          false,
						"object":               "model_permission",
						"organization":         "*",
					},
				},
			},
			// Claude 3.5 Sonnet (Previous)
			map[string]interface{}{
				"id":       "claude-3-5-sonnet-20240620",
				"object":   "model",
				"created":  1718841600, // 2024-06-20
				"owned_by": "anthropic",
				"permission": []interface{}{
					map[string]interface{}{
						"allow_create_engine":  false,
						"allow_fine_tuning":    false,
						"allow_logprobs":       false,
						"allow_sampling":       true,
						"allow_search_indices": false,
						"allow_view":           true,
						"created":              int(time.Now().Unix()),
						"group":                nil,
						"id":                   "modelperm-claude-3-5-sonnet-20240620",
						"is_blocking":          false,
						"object":               "model_permission",
						"organization":         "*",
					},
				},
			},
			// Claude 3.5 Haiku (Latest)
			map[string]interface{}{
				"id":       "claude-3-5-haiku-20241022",
				"object":   "model",
				"created":  1729555200, // 2024-10-22
				"owned_by": "anthropic",
				"permission": []interface{}{
					map[string]interface{}{
						"allow_create_engine":  false,
						"allow_fine_tuning":    false,
						"allow_logprobs":       false,
						"allow_sampling":       true,
						"allow_search_indices": false,
						"allow_view":           true,
						"created":              int(time.Now().Unix()),
						"group":                nil,
						"id":                   "modelperm-claude-3-5-haiku-20241022",
						"is_blocking":          false,
						"object":               "model_permission",
						"organization":         "*",
					},
				},
			},
			// Claude 3 Opus
			map[string]interface{}{
				"id":       "claude-3-opus-20240229",
				"object":   "model",
				"created":  1709251200, // 2024-02-29
				"owned_by": "anthropic",
				"permission": []interface{}{
					map[string]interface{}{
						"allow_create_engine":  false,
						"allow_fine_tuning":    false,
						"allow_logprobs":       false,
						"allow_sampling":       true,
						"allow_search_indices": false,
						"allow_view":           true,
						"created":              int(time.Now().Unix()),
						"group":                nil,
						"id":                   "modelperm-claude-3-opus-20240229",
						"is_blocking":          false,
						"object":               "model_permission",
						"organization":         "*",
					},
				},
			},
			// Claude 3 Sonnet
			map[string]interface{}{
				"id":       "claude-3-sonnet-20240229",
				"object":   "model",
				"created":  1709251200, // 2024-02-29
				"owned_by": "anthropic",
				"permission": []interface{}{
					map[string]interface{}{
						"allow_create_engine":  false,
						"allow_fine_tuning":    false,
						"allow_logprobs":       false,
						"allow_sampling":       true,
						"allow_search_indices": false,
						"allow_view":           true,
						"created":              int(time.Now().Unix()),
						"group":                nil,
						"id":                   "modelperm-claude-3-sonnet-20240229",
						"is_blocking":          false,
						"object":               "model_permission",
						"organization":         "*",
					},
				},
			},
			// Claude 3 Haiku
			map[string]interface{}{
				"id":       "claude-3-haiku-20240307",
				"object":   "model",
				"created":  1709769600, // 2024-03-07
				"owned_by": "anthropic",
				"permission": []interface{}{
					map[string]interface{}{
						"allow_create_engine":  false,
						"allow_fine_tuning":    false,
						"allow_logprobs":       false,
						"allow_sampling":       true,
						"allow_search_indices": false,
						"allow_view":           true,
						"created":              int(time.Now().Unix()),
						"group":                nil,
						"id":                   "modelperm-claude-3-haiku-20240307",
						"is_blocking":          false,
						"object":               "model_permission",
						"organization":         "*",
					},
				},
			},
		},
	}
}

// setCORSHeadersStandalone is a standalone CORS header setter
func setCORSHeadersStandalone(w http.ResponseWriter, r *http.Request) {
	origin := r.Header.Get("Origin")
	if origin == "" {
		origin = "*"
	}
	
	w.Header().Set("Access-Control-Allow-Origin", origin)
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.Header().Set("Access-Control-Max-Age", "3600")
}