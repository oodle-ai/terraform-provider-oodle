package oodlehttp

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	jsoniter "github.com/json-iterator/go"

	"terraform-provider-oodle/internal/oodlehttp/clientmodels"
)

func testClient(t *testing.T, handler http.Handler) (*PromptClient, func()) {
	t.Helper()
	server := httptest.NewServer(handler)
	return NewPromptClient(&OodleApiClient{
		DeploymentUrl: server.URL,
		Instance:      "oodle_internal",
		Headers:       http.Header{},
		HttpClient:    server.Client(),
	}), server.Close
}

// A read that feeds a write must not resolve. The default
// expands a prompt's @@@oodlePrompt@@@ references into the
// child's text, and an expansion never equals the reference it
// was stored as -- so every plan would report drift and every
// apply would publish a version that changes nothing.
func TestGetDoesNotResolveReferences(t *testing.T) {
	var gotQuery url.Values
	client, done := testClient(t, http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			gotQuery = r.URL.Query()
			_ = jsoniter.NewEncoder(w).Encode(clientmodels.Prompt{
				Name:    "groot/sub-agent/logs",
				Prompt:  "@@@oodlePrompt:name=base|label=production@@@ logs",
				Version: 4,
			})
		},
	))
	defer done()

	got, err := client.Get(
		context.Background(), "groot/sub-agent/logs", "production",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotQuery.Get("resolve") != "false" {
		t.Errorf("resolve = %q, want false", gotQuery.Get("resolve"))
	}
	if gotQuery.Get("label") != "production" {
		t.Errorf("label = %q, want production", gotQuery.Get("label"))
	}
	if got.Version != 4 {
		t.Errorf("version = %d, want 4", got.Version)
	}
}

// A name routinely contains "/". Unencoded it splits into extra
// path segments and 404s against a route that does not exist.
func TestGetEncodesTheName(t *testing.T) {
	var gotPath string
	client, done := testClient(t, http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			gotPath = r.URL.EscapedPath()
			_ = jsoniter.NewEncoder(w).Encode(clientmodels.Prompt{})
		},
	))
	defer done()

	if _, err := client.Get(
		context.Background(), "groot/sub-agent/logs", "production",
	); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := "/v1/api/instance/oodle_internal/langfuse/api/public/" +
		"v2/prompts/groot%2Fsub-agent%2Flogs"
	if gotPath != want {
		t.Errorf("path = %q, want %q", gotPath, want)
	}
}

// Absent is how Terraform learns the prompt has to be created,
// or that something removed it outside the state. Reported as an
// error it would fail the plan instead.
func TestGetTreatsMissingAsAbsent(t *testing.T) {
	client, done := testClient(t, http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		},
	))
	defer done()

	got, err := client.Get(context.Background(), "nope", "production")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != nil {
		t.Errorf("got %+v, want nil", got)
	}
}

func TestGetReportsOtherFailures(t *testing.T) {
	client, done := testClient(t, http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		},
	))
	defer done()

	if _, err := client.Get(
		context.Background(), "x", "production",
	); err == nil {
		t.Error("want an error for a 500")
	}
}

// A version is immutable, so publishing a change is a POST to
// the collection rather than a PUT on a row.
func TestCreatePostsToTheCollection(t *testing.T) {
	var gotMethod, gotPath string
	var gotBody clientmodels.Prompt
	client, done := testClient(t, http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			gotMethod = r.Method
			gotPath = r.URL.EscapedPath()
			_ = jsoniter.NewDecoder(r.Body).Decode(&gotBody)
			w.WriteHeader(http.StatusCreated)
			_ = jsoniter.NewEncoder(w).Encode(clientmodels.Prompt{
				Name:    gotBody.Name,
				Version: 6,
			})
		},
	))
	defer done()

	published, err := client.Create(
		context.Background(),
		&clientmodels.Prompt{
			Name:          "team/reply",
			Type:          "text",
			Prompt:        "hello",
			Labels:        []string{"production"},
			CommitMessage: "abc1234 a change",
		},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotMethod != http.MethodPost {
		t.Errorf("method = %q, want POST", gotMethod)
	}
	if gotPath != "/v1/api/instance/oodle_internal/langfuse/api/"+
		"public/v2/prompts" {
		t.Errorf("path = %q", gotPath)
	}
	if gotBody.CommitMessage != "abc1234 a change" {
		t.Errorf("commitMessage = %q", gotBody.CommitMessage)
	}
	if published.Version != 6 {
		t.Errorf("version = %d, want 6", published.Version)
	}
}

// Already gone is the outcome a destroy wanted.
func TestDeleteAcceptsMissing(t *testing.T) {
	client, done := testClient(t, http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		},
	))
	defer done()

	if err := client.Delete(context.Background(), "gone"); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}
