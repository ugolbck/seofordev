package indexnow

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

func generateKey() (string, error) {
	bytes := make([]byte, 16) // 128-bit
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func Setup() error {
	key, err := generateKey()
	if err != nil {
		return err
	}

	fmt.Printf("✅ IndexNow key generated: %s\n\n", key)
	fmt.Println("👉 To finish setup:")
	fmt.Printf("1. Create a file named %s.txt at the root of your site.\n", key)
	fmt.Printf("2. File contents must contain only this key (no spaces, no newlines).\n")
	fmt.Printf("   Example: https://yourdomain.com/%s.txt\n", key)
	fmt.Println("3. Once deployed, you can submit pages using:")
	fmt.Printf("   seo index submit <url> --key %s\n\n", key)

	return nil
}
