package oodlehttp

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"terraform-provider-oodle/internal/oodlehttp/clientmodels"
)

func TestModelClientDelete(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		wantErr    bool
	}{
		{
			name:       "200 OK returns nil error",
			statusCode: http.StatusOK,
			wantErr:    false,
		},
		{
			name:       "204 No Content returns nil error",
			statusCode: http.StatusNoContent,
			wantErr:    false,
		},
		{
			name:       "400 Bad Request returns error",
			statusCode: http.StatusBadRequest,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
				w.Write([]byte("response body"))
			}))
			defer server.Close()

			apiClient := &OodleApiClient{
				HttpClient:    server.Client(),
				DeploymentUrl: server.URL,
				Instance:      "test-instance",
				Headers:       http.Header{},
			}

			client := NewModelClient[*clientmodels.Monitor](
				apiClient,
				"monitors",
				func() *clientmodels.Monitor { return &clientmodels.Monitor{} },
			)

			err := client.Delete(context.Background(), "test-id")
			if tt.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("expected nil error, got: %v", err)
			}
			if tt.wantErr && err != nil {
				if len(err.Error()) == 0 {
					t.Error("expected non-empty error message")
				}
			}
		})
	}
}
