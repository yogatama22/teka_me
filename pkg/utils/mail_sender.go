package utils

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
)

// SendEmail mengirim email via Sender.net API v2 (/v2/message/send)
func SendEmail(to, subject, body string) error {
	fromEmail := os.Getenv("SENDER_FROM_EMAIL")
	fromName := os.Getenv("SENDER_FROM_NAME")
	apiKey := os.Getenv("SENDER_API_KEY")

	if fromEmail == "" || apiKey == "" {
		return errors.New("SENDER_FROM_EMAIL or SENDER_API_KEY not set in ENV")
	}

	payload := map[string]interface{}{
		"from": map[string]string{
			"email": fromEmail,
			"name":  fromName,
		},
		"to": map[string]string{ // <-- HARUS BUKAN ARRAY
			"email": to,
		},
		"subject": subject,
		"html":    body,
	}

	jsonBody, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", "https://api.sender.net/v2/message/send", bytes.NewBuffer(jsonBody))
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		return errors.New("sender.net failed send email: " + string(respBody))
	}

	return nil
}
