package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"pulsewatch-notification-service/internal/notifier"
)

type MailpitMessage struct {
	ID      string `json:"ID"`
	Subject string `json:"Subject"`
	To      []struct {
		Address string `json:"Address"`
	} `json:"To"`
}

type MailpitMessagesResponse struct {
	Total    int              `json:"total"`
	Messages []MailpitMessage `json:"messages"`
}

func TestEmailSender_Mailpit(t *testing.T) {
	sender := notifier.NewEmailSender()

	testRecipient := fmt.Sprintf("test-%d@example.com", time.Now().UnixNano())
	testSubject := "Test Email Verification"
	testText := "Hello, please verify your account."
	testHTML := "<p>Hello, please verify your account.</p>"

	err := sender.Send(testRecipient, testSubject, testText, testHTML)
	if err != nil {
		t.Fatalf("Failed to send email to Mailpit: %v", err)
	}

	// Verify message in Mailpit REST API
	resp, err := http.Get("http://localhost:8025/api/v1/messages")
	if err != nil {
		t.Fatalf("Failed to reach Mailpit API: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected 200 OK from Mailpit API, got %d", resp.StatusCode)
	}

	var mailpitData MailpitMessagesResponse
	if err := json.NewDecoder(resp.Body).Decode(&mailpitData); err != nil {
		t.Fatalf("Failed to decode Mailpit response: %v", err)
	}

	found := false
	for _, m := range mailpitData.Messages {
		if m.Subject == testSubject {
			for _, to := range m.To {
				if to.Address == testRecipient {
					found = true
					break
				}
			}
		}
	}

	if !found {
		t.Fatalf("Sent email with subject '%s' to '%s' was not found in Mailpit messages", testSubject, testRecipient)
	}
}
