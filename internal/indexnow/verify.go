package indexnow

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// VerifyKey checks if the given key file is accessible and valid on the domain
func VerifyKey(domain, key string) error {
	if !strings.HasPrefix(domain, "http://") && !strings.HasPrefix(domain, "https://") {
		domain = "https://" + domain
	}

	url := fmt.Sprintf("%s/%s.txt", strings.TrimRight(domain, "/"), key)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return fmt.Errorf("could not reach %s: %v", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status %d when fetching %s", resp.StatusCode, url)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %v", err)
	}

	content := strings.TrimSpace(string(body))
	if content != key {
		return fmt.Errorf("file found at %s but contents do not match key", url)
	}

	return nil
}
