package oodlehttp

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestMutingRuleClientGetFallsBackToTheCollection covers a recurring
// rule, which the API's read-by-id answers 404 for however live it
// is. Without the fallback Terraform reads every recurring rule as
// deleted and creates a second copy on the next apply.
func TestMutingRuleClientGetFallsBackToTheCollection(t *testing.T) {
	const id = "5a6e8a44-7046-418d-9a81-5470d5ea46bf"

	var listed bool
	server := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/v1/api/instance/test-instance/muting-rules" {
				listed = true
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(
					`[{"id":"` + id + `","name":"recurring",` +
						`"scheduleIds":["schedule-1"]}]`,
				))

				return
			}

			w.WriteHeader(http.StatusNotFound)
		},
	))
	defer server.Close()

	client := NewMutingRuleClient(newTestOodleAPIClient(server))
	rule, err := client.Get(context.Background(), id)
	if err != nil {
		t.Fatalf("expected the rule, got error: %v", err)
	}

	if !listed {
		t.Error("expected the collection to be consulted")
	}

	if rule.ID != id {
		t.Errorf("expected rule %q, got %q", id, rule.ID)
	}

	if len(rule.ScheduleIDs) != 1 || rule.ScheduleIDs[0] != "schedule-1" {
		t.Errorf("expected the schedule ids, got %v", rule.ScheduleIDs)
	}
}

// TestMutingRuleClientGetReportsAMissingRule guards the fallback from
// turning a genuinely deleted rule into one Terraform keeps in state.
func TestMutingRuleClientGetReportsAMissingRule(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/v1/api/instance/test-instance/muting-rules" {
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`[]`))

				return
			}

			w.WriteHeader(http.StatusNotFound)
		},
	))
	defer server.Close()

	client := NewMutingRuleClient(newTestOodleAPIClient(server))
	_, err := client.Get(context.Background(), "gone")
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got: %v", err)
	}
}

// TestMutingRuleClientGetUsesTheDirectRead pins that a one-off rule,
// which reads by id perfectly well, is not put through the listing.
func TestMutingRuleClientGetUsesTheDirectRead(t *testing.T) {
	const id = "f9be1c43-7041-4ba4-bbd2-6886787a6539"

	server := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path ==
				"/v1/api/instance/test-instance/muting-rules/"+id {
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"id":"` + id + `"}`))

				return
			}

			t.Errorf("unexpected request to %q", r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		},
	))
	defer server.Close()

	client := NewMutingRuleClient(newTestOodleAPIClient(server))
	rule, err := client.Get(context.Background(), id)
	if err != nil {
		t.Fatalf("expected the rule, got error: %v", err)
	}

	if rule.ID != id {
		t.Errorf("expected rule %q, got %q", id, rule.ID)
	}
}
