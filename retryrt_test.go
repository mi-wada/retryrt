package retryrt_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mi-wada/retryrt"
)

func TestMaxRetries(t *testing.T) {
	tests := []struct {
		name           string
		failCount      int
		maxRetries     int
		expectedStatus int
	}{
		{
			name:           "Success after two failures within retry limit",
			failCount:      2,
			maxRetries:     2,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Failure when retries are exhausted",
			failCount:      2,
			maxRetries:     1,
			expectedStatus: http.StatusServiceUnavailable,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			callCount := 0
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				callCount++
				if callCount <= tt.failCount {
					w.WriteHeader(http.StatusServiceUnavailable)
					return
				}
				w.WriteHeader(http.StatusOK)
			}))
			t.Cleanup(srv.Close)

			client := &http.Client{
				Transport: retryrt.New(
					http.DefaultTransport,
					retryrt.WithMaxRetries(tt.maxRetries),
					retryrt.WithBackoff(retryrt.DefaultBackoff(0, 0)),
				),
			}

			resp, err := client.Get(srv.URL)
			if err != nil {
				t.Fatalf("Request failed: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("got status %d, want %d", resp.StatusCode, tt.expectedStatus)
			}
		})
	}
}
