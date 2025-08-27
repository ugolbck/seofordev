package services

import (
	"fmt"
	"strings"
	"time"

	"github.com/ugolbck/seofordev/internal/api"
	"github.com/ugolbck/seofordev/internal/config"
)

// BriefService provides content brief functionality
type BriefService struct {
	client *api.Client
}

// NewBriefService creates a new brief service
func NewBriefService() (*BriefService, error) {
	// Validate API key and show setup message if needed
	cfg, err := config.ValidateAPIKeyWithMessage()
	if err != nil {
		return nil, err
	}

	client := api.NewClient(cfg.GetEffectiveBaseURL(), cfg.APIKey)
	
	return &BriefService{
		client: client,
	}, nil
}

// GenerateBrief generates a content brief for a keyword
func (s *BriefService) GenerateBrief(keyword string) (*api.GenerateBriefResponse, error) {
	result, err := s.client.GenerateBrief(keyword)
	if err != nil {
		return nil, formatBriefAPIError("generate brief", err)
	}
	
	return result, nil
}

// WaitForBrief waits for a brief to be completed and returns the result
func (s *BriefService) WaitForBrief(briefID string) (*api.BriefResponse, error) {
	for {
		status, err := s.client.GetBriefStatus(briefID)
		if err != nil {
			return nil, formatBriefAPIError("get brief status", err)
		}
		
		if status.Status == "completed" && status.Brief != nil {
			return status, nil
		}
		
		if status.Status == "failed" {
			return nil, fmt.Errorf("brief generation failed for brief ID: %s", briefID)
		}
		
		// Wait before polling again
		time.Sleep(2 * time.Second)
	}
}

// GetHistory returns brief generation history
func (s *BriefService) GetHistory() (*api.BriefHistoryResponse, error) {
	history, err := s.client.GetBriefHistory()
	if err != nil {
		return nil, formatBriefAPIError("get brief history", err)
	}
	
	return history, nil
}

// GetBriefStatus returns the status of a specific brief
func (s *BriefService) GetBriefStatus(briefID string) (*api.BriefResponse, error) {
	status, err := s.client.GetBriefStatus(briefID)
	if err != nil {
		return nil, formatBriefAPIError("get brief status", err)
	}
	
	return status, nil
}

// formatBriefAPIError provides user-friendly error messages for API errors
func formatBriefAPIError(operation string, err error) error {
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