package oodlehttp

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"terraform-provider-oodle/internal/oodlehttp/clientmodels"
)

const codeLibrariesPath = "/v1/api/instance/test-instance/langfuse/api/public/code-libraries"

// A delete the API refuses because the library is still imported
// names each importer, so the user knows what to change first.
func TestCodeLibraryDeleteInUseNamesImporters(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodDelete ||
				r.URL.Path != codeLibrariesPath+"/lib-1" {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			w.WriteHeader(http.StatusConflict)
			_, _ = w.Write([]byte(`{"message":"cannot delete","error":"Conflict",` +
				`"usedBy":[{"id":"t-1","name":"tone","kind":"template"},` +
				`{"id":"l-2","name":"helpers","kind":"library"}]}`))
		},
	))
	defer server.Close()

	client := NewGenAICodeLibraryClient(newTestOodleAPIClient(server))
	err := client.Delete(context.Background(), "lib-1")
	if err == nil {
		t.Fatal("expected an error")
	}
	for _, want := range []string{
		`template "tone" (id t-1)`, `library "helpers" (id l-2)`,
	} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("expected %q in %q", want, err.Error())
		}
	}
}

// Any other failure keeps the status and the body in the error.
func TestCodeLibraryDeleteOtherErrorKeepsBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(`{"message":"boom"}`))
		},
	))
	defer server.Close()

	client := NewGenAICodeLibraryClient(newTestOodleAPIClient(server))
	err := client.Delete(context.Background(), "lib-1")
	if err == nil || !strings.Contains(err.Error(), "boom") {
		t.Fatalf("expected the body in the error, got %v", err)
	}
}

func TestCodeLibraryDeleteStatusCodes(t *testing.T) {
	for _, tc := range deleteStatusCodeTests() {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(
				func(w http.ResponseWriter, _ *http.Request) {
					w.WriteHeader(tc.statusCode)
				},
			))
			defer server.Close()

			client := NewGenAICodeLibraryClient(newTestOodleAPIClient(server))
			err := client.Delete(context.Background(), "lib-1")
			if (err != nil) != tc.wantErr {
				t.Errorf("wantErr %v, got %v", tc.wantErr, err)
			}
		})
	}
}

// The update sends the description and the source, and never the name,
// which the API does not accept a change of.
func TestCodeLibraryUpdateSendsNoName(t *testing.T) {
	var sent map[string]any
	server := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPatch ||
				r.URL.Path != codeLibrariesPath+"/lib-1" {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			body, _ := io.ReadAll(r.Body)
			_ = json.Unmarshal(body, &sent)
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"id":"lib-1","name":"helpers",` +
				`"description":"","sourceCode":"x = 2","version":2}`))
		},
	))
	defer server.Close()

	description := ""
	source := "x = 2"
	client := NewGenAICodeLibraryClient(newTestOodleAPIClient(server))
	updated, err := client.Update(
		context.Background(),
		&clientmodels.GenAICodeLibrary{
			ID: "lib-1", Name: "helpers",
			Description: &description, SourceCode: &source,
		},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := sent["name"]; ok {
		t.Errorf("expected no name in the request, got %v", sent)
	}
	if got, ok := sent["description"]; !ok || got != "" {
		t.Errorf("expected an empty description to be sent, got %v", sent)
	}
	if updated.Version != 2 {
		t.Errorf("expected version 2, got %d", updated.Version)
	}
}
