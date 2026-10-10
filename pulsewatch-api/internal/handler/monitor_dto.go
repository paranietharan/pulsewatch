package handler

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/google/uuid"
)

type CreateMonitorRequest struct {
	Name               string            `json:"name"`
	URL                string            `json:"url"`
	Method             string            `json:"method"`
	IntervalSeconds    int               `json:"interval_seconds"`
	TimeoutSeconds     int               `json:"timeout_seconds"`
	ExpectedStatusCode int               `json:"expected_status_code"`
	Headers            map[string]string `json:"headers"`
	Body               string            `json:"body"`
	SSLCheckEnabled    *bool             `json:"ssl_check_enabled"`
	FailureThreshold   int               `json:"failure_threshold"`
	RecoveryThreshold  int               `json:"recovery_threshold"`
	AlertChannelIDs    []uuid.UUID       `json:"alert_channel_ids"`
}

func (r *CreateMonitorRequest) Validate() error {
	r.Name = strings.TrimSpace(r.Name)
	if r.Name == "" {
		return fmt.Errorf("name is required")
	}
	if len(r.Name) > 255 {
		return fmt.Errorf("name cannot exceed 255 characters")
	}

	r.URL = strings.TrimSpace(r.URL)
	if r.URL == "" {
		return fmt.Errorf("url is required")
	}
	parsedURL, err := url.ParseRequestURI(r.URL)
	if err != nil || (parsedURL.Scheme != "http" && parsedURL.Scheme != "https") {
		return fmt.Errorf("url must be a valid http or https URL")
	}

	if r.Method == "" {
		r.Method = "GET"
	} else {
		r.Method = strings.ToUpper(strings.TrimSpace(r.Method))
		validMethods := map[string]bool{
			"GET": true, "POST": true, "PUT": true, "PATCH": true, "DELETE": true, "HEAD": true, "OPTIONS": true,
		}
		if !validMethods[r.Method] {
			return fmt.Errorf("unsupported HTTP method: %s", r.Method)
		}
	}

	if r.IntervalSeconds == 0 {
		r.IntervalSeconds = 60
	} else if r.IntervalSeconds < 10 || r.IntervalSeconds > 86400 {
		return fmt.Errorf("interval_seconds must be between 10 and 86400")
	}

	if r.TimeoutSeconds == 0 {
		r.TimeoutSeconds = 10
	} else if r.TimeoutSeconds < 1 || r.TimeoutSeconds > 60 {
		return fmt.Errorf("timeout_seconds must be between 1 and 60")
	}
	if r.TimeoutSeconds > r.IntervalSeconds {
		return fmt.Errorf("timeout_seconds cannot be greater than interval_seconds")
	}

	if r.ExpectedStatusCode == 0 {
		r.ExpectedStatusCode = 200
	} else if r.ExpectedStatusCode < 100 || r.ExpectedStatusCode > 599 {
		return fmt.Errorf("expected_status_code must be between 100 and 599")
	}

	if r.FailureThreshold == 0 {
		r.FailureThreshold = 3
	} else if r.FailureThreshold < 1 || r.FailureThreshold > 10 {
		return fmt.Errorf("failure_threshold must be between 1 and 10")
	}

	if r.RecoveryThreshold == 0 {
		r.RecoveryThreshold = 2
	} else if r.RecoveryThreshold < 1 || r.RecoveryThreshold > 10 {
		return fmt.Errorf("recovery_threshold must be between 1 and 10")
	}

	if r.Headers == nil {
		r.Headers = make(map[string]string)
	}

	return nil
}

type UpdateMonitorRequest struct {
	Name               *string           `json:"name"`
	URL                *string           `json:"url"`
	Method             *string           `json:"method"`
	IntervalSeconds    *int              `json:"interval_seconds"`
	TimeoutSeconds     *int              `json:"timeout_seconds"`
	ExpectedStatusCode *int              `json:"expected_status_code"`
	Headers            map[string]string `json:"headers"`
	Body               *string           `json:"body"`
	SSLCheckEnabled    *bool             `json:"ssl_check_enabled"`
	FailureThreshold   *int              `json:"failure_threshold"`
	RecoveryThreshold  *int              `json:"recovery_threshold"`
}

func (r *UpdateMonitorRequest) Validate() error {
	if r.Name != nil {
		trimmed := strings.TrimSpace(*r.Name)
		if trimmed == "" {
			return fmt.Errorf("name cannot be empty")
		}
		if len(trimmed) > 255 {
			return fmt.Errorf("name cannot exceed 255 characters")
		}
		*r.Name = trimmed
	}

	if r.URL != nil {
		trimmed := strings.TrimSpace(*r.URL)
		if trimmed == "" {
			return fmt.Errorf("url cannot be empty")
		}
		parsedURL, err := url.ParseRequestURI(trimmed)
		if err != nil || (parsedURL.Scheme != "http" && parsedURL.Scheme != "https") {
			return fmt.Errorf("url must be a valid http or https URL")
		}
		*r.URL = trimmed
	}

	if r.Method != nil {
		upper := strings.ToUpper(strings.TrimSpace(*r.Method))
		validMethods := map[string]bool{
			"GET": true, "POST": true, "PUT": true, "PATCH": true, "DELETE": true, "HEAD": true, "OPTIONS": true,
		}
		if !validMethods[upper] {
			return fmt.Errorf("unsupported HTTP method: %s", upper)
		}
		*r.Method = upper
	}

	if r.IntervalSeconds != nil && (*r.IntervalSeconds < 10 || *r.IntervalSeconds > 86400) {
		return fmt.Errorf("interval_seconds must be between 10 and 86400")
	}

	if r.TimeoutSeconds != nil && (*r.TimeoutSeconds < 1 || *r.TimeoutSeconds > 60) {
		return fmt.Errorf("timeout_seconds must be between 1 and 60")
	}

	if r.ExpectedStatusCode != nil && (*r.ExpectedStatusCode < 100 || *r.ExpectedStatusCode > 599) {
		return fmt.Errorf("expected_status_code must be between 100 and 599")
	}

	if r.FailureThreshold != nil && (*r.FailureThreshold < 1 || *r.FailureThreshold > 10) {
		return fmt.Errorf("failure_threshold must be between 1 and 10")
	}

	if r.RecoveryThreshold != nil && (*r.RecoveryThreshold < 1 || *r.RecoveryThreshold > 10) {
		return fmt.Errorf("recovery_threshold must be between 1 and 10")
	}

	return nil
}

type CreateAlertChannelRequest struct {
	Type   string          `json:"type"`
	Name   string          `json:"name"`
	Config json.RawMessage `json:"config"`
}

func (r *CreateAlertChannelRequest) Validate() error {
	r.Type = strings.ToUpper(strings.TrimSpace(r.Type))
	if r.Type != "EMAIL" && r.Type != "WEBHOOK" {
		return fmt.Errorf("type must be either EMAIL or WEBHOOK")
	}

	r.Name = strings.TrimSpace(r.Name)
	if r.Name == "" {
		return fmt.Errorf("name is required")
	}

	if len(r.Config) == 0 || string(r.Config) == "{}" {
		return fmt.Errorf("config is required")
	}

	return nil
}

type PaginatedResponse struct {
	Data       interface{} `json:"data"`
	Total      int         `json:"total"`
	Page       int         `json:"page"`
	Limit      int         `json:"limit"`
	TotalPages int         `json:"total_pages"`
}
