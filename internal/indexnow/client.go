package indexnow

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

// Request body spec: https://www.indexnow.org/documentation
type indexNowRequest struct {
	Host    string   `json:"host"`
	Key     string   `json:"key"`
	URLList []string `json:"urlList"`
}

// Submit multiple URLs with the same host
func SubmitURLs(rawURLs []string, key string) error {
	if len(rawURLs) == 0 {
		return fmt.Errorf("no URLs provided")
	}

	// Parse first URL to extract host
	parsed, err := url.Parse(rawURLs[0])
	if err != nil {
		return fmt.Errorf("invalid URL: %v", err)
	}
	if parsed.Scheme == "" || parsed.Host == "" {
		return fmt.Errorf("URL must include scheme and host, e.g. https://example.com/page")
	}

	// Ensure all URLs share the same host
	host := parsed.Host
	for _, u := range rawURLs {
		pu, err := url.Parse(u)
		if err != nil {
			return fmt.Errorf("invalid URL: %v", err)
		}
		if pu.Host != host {
			return fmt.Errorf("all URLs must have the same host (found %s and %s)", host, pu.Host)
		}
	}

	reqBody := indexNowRequest{
		Host:    host,
		Key:     key,
		URLList: rawURLs,
	}

	data, err := json.Marshal(reqBody)
	if err != nil {
		return err
	}

	endpoint := "https://api.indexnow.org/indexnow"
	client := &http.Client{Timeout: 10 * time.Second}

	resp, err := client.Post(endpoint, "application/json", bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("failed to submit URLs: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		return fmt.Errorf("unexpected status code %d from IndexNow", resp.StatusCode)
	}

	return nil
}
