package oodlehttp

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGrafanaFolderClientDelete(t *testing.T) {
	for _, tt := range deleteStatusCodeTests() {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
				w.Write([]byte("response body"))
			}))
			defer server.Close()

			client := NewGrafanaFolderClient(newTestOodleAPIClient(server))

			err := client.Delete(context.Background(), "test-uid")
			if tt.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("expected nil error, got: %v", err)
			}
		})
	}
}

func TestGrafanaDashboardClientDelete(t *testing.T) {
	for _, tt := range deleteStatusCodeTests() {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
				w.Write([]byte("response body"))
			}))
			defer server.Close()

			client := NewGrafanaDashboardClient(newTestOodleAPIClient(server))

			err := client.Delete(context.Background(), "test-uid")
			if tt.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("expected nil error, got: %v", err)
			}
		})
	}
}
