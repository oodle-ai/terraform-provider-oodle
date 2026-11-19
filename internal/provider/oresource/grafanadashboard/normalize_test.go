package grafanadashboard

import "testing"

func TestConfigJSONForState(t *testing.T) {
	tests := []struct {
		name   string
		remote string
		prior  string
		uid    string
		want   string
	}{
		{
			name:   "id and version from API are ignored and prior is kept",
			remote: `{"id":12,"uid":"abc","title":"T","version":7,"panels":[]}`,
			prior:  "{\n  \"title\": \"T\",\n  \"uid\": \"abc\",\n  \"panels\": []\n}",
			uid:    "abc",
			want:   "{\n  \"title\": \"T\",\n  \"uid\": \"abc\",\n  \"panels\": []\n}",
		},
		{
			name:   "stale version in prior does not count as drift",
			remote: `{"id":12,"uid":"abc","title":"T","version":8}`,
			prior:  `{"uid":"abc","title":"T","version":3}`,
			uid:    "abc",
			want:   `{"uid":"abc","title":"T","version":3}`,
		},
		{
			name:   "server-assigned uid is not drift when config has none",
			remote: `{"id":12,"uid":"gen123","title":"T","version":1}`,
			prior:  `{"title":"T"}`,
			uid:    "gen123",
			want:   `{"title":"T"}`,
		},
		{
			name:   "real drift returns normalized remote",
			remote: `{"id":12,"uid":"abc","title":"Changed","version":9}`,
			prior:  `{"uid":"abc","title":"T"}`,
			uid:    "abc",
			want:   `{"title":"Changed","uid":"abc"}`,
		},
		{
			name:   "import with no prior returns normalized remote",
			remote: `{"id":12,"uid":"abc","title":"T","version":9}`,
			prior:  "",
			uid:    "abc",
			want:   `{"title":"T","uid":"abc"}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := configJSONForState(tt.remote, tt.prior, tt.uid)
			if err != nil {
				t.Fatal(err)
			}
			if got != tt.want {
				t.Errorf("got %s, want %s", got, tt.want)
			}
		})
	}
}
