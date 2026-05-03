package oodlehttp

import (
	"net/http"
	"net/http/httptest"
)

// newTestOodleAPIClient creates an OodleApiClient configured to use the given test server.
func newTestOodleAPIClient(server *httptest.Server) *OodleApiClient {
	return &OodleApiClient{
		HttpClient:    server.Client(),
		DeploymentUrl: server.URL,
		Instance:      "test-instance",
		Headers:       http.Header{},
	}
}

// deleteTestCase defines a test case for delete endpoint tests.
type deleteTestCase struct {
	name       string
	statusCode int
	wantErr    bool
}

// deleteStatusCodeTests returns the standard set of test cases for delete endpoints.
// All delete endpoints should accept both 200 OK and 204 No Content as success,
// and treat other status codes (4xx, 5xx) as errors.
func deleteStatusCodeTests() []deleteTestCase {
	return []deleteTestCase{
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
			name:       "500 Internal Server Error returns error",
			statusCode: http.StatusInternalServerError,
			wantErr:    true,
		},
	}
}
