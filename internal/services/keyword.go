package services

import (
	"fmt"
	"strings"

	"github.com/ugolbck/seofordev/internal/api"
	"github.com/ugolbck/seofordev/internal/config"
)

// KeywordService provides keyword functionality
type KeywordService struct {
	client *api.Client
}

// NewKeywordService creates a new keyword service
func NewKeywordService() (*KeywordService, error) {
	// Validate API key and show setup message if needed
	cfg, err := config.ValidateAPIKeyWithMessage()
	if err != nil {
		return nil, err
	}

	client := api.NewClient(cfg.GetEffectiveBaseURL(), cfg.APIKey)
	
	return &KeywordService{
		client: client,
	}, nil
}

// GenerateKeywords generates keywords for a seed keyword
func (s *KeywordService) GenerateKeywords(seedKeyword string) (*api.GenerateKeywordsResponse, error) {
	result, err := s.client.GenerateKeywords(seedKeyword)
	if err != nil {
		return nil, formatAPIError("generate keywords", err)
	}
	
	return result, nil
}

// GetHistory returns keyword generation history
func (s *KeywordService) GetHistory() (*api.KeywordHistoryResponse, error) {
	history, err := s.client.GetKeywordHistory()
	if err != nil {
		return nil, formatAPIError("get keyword history", err)
	}
	
	return history, nil
}

// GetKeywordGeneration returns detailed results of a specific keyword generation
func (s *KeywordService) GetKeywordGeneration(generationID string) (*api.GenerateKeywordsResponse, error) {
	// Get full history since the individual endpoint might not exist
	history, err := s.client.GetKeywordHistory()
	if err != nil {
		return nil, formatAPIError("get keyword history", err)
	}
	
	// Find the specific generation
	for _, gen := range history.Generations {
		if gen.ID == generationID {
			// Convert to GenerateKeywordsResponse format
			result := &api.GenerateKeywordsResponse{
				ID:           gen.ID,
				SeedKeyword:  gen.SeedKeyword,
				Keywords:     gen.Keywords,
				Status:       gen.Status,
				CreditsUsed:  gen.CreditsUsed,
				GeneratedAt:  gen.GeneratedAt,
				TotalResults: gen.TotalResults,
			}
			
			return result, nil
		}
	}
	
	return nil, fmt.Errorf("keyword generation with ID %s not found", generationID)
}

// GetCreditBalance returns the current credit balance
func (s *KeywordService) GetCreditBalance() (*api.CreditBalanceResponse, error) {
	balance, err := s.client.GetCreditBalance()
	if err != nil {
		return nil, formatAPIError("get credit balance", err)
	}
	
	return balance, nil
}

// formatAPIError provides user-friendly error messages for API errors
func formatAPIError(operation string, err error) error {
	if apiErr, ok := err.(*api.APIError); ok {
		switch apiErr.StatusCode {
		case 401:
			return fmt.Errorf("invalid API key - please check your API key configuration.\n\nGet your API key at: https://seofor.dev/dashboard\nSet it with: export SEO_API_KEY=your_key_here")
		case 402:
			return fmt.Errorf("insufficient credits to %s.\n\nManage your subscription: https://seofor.dev/dashboard", operation)
		case 403:
			return fmt.Errorf("access denied - please check your API key permissions")
		case 429:
			return fmt.Errorf("rate limit exceeded - please wait a moment and try again")
		case 500, 502, 503, 504:
			return fmt.Errorf("server error - please try again later or contact support")
		default:
			if strings.Contains(apiErr.Message, "API key") {
				return fmt.Errorf("API key issue: %s\n\nGet your API key at: https://seofor.dev/dashboard", apiErr.Message)
			}
			return fmt.Errorf("failed to %s: %s (HTTP %d)", operation, apiErr.Message, apiErr.StatusCode)
		}
	}
	
	return fmt.Errorf("failed to %s: %w", operation, err)
}