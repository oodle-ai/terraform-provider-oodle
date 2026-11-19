package mutingrule

import (
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestKeepStartsAt(t *testing.T) {
	now := time.Date(2026, 11, 19, 21, 30, 0, 0, time.UTC)
	str := types.StringValue

	tests := []struct {
		name   string
		prior  types.String
		server types.String
		want   types.String
	}{
		{
			name:   "past start moved forward to save time keeps prior",
			prior:  str("2026-11-19T00:00:00Z"),
			server: str("2026-11-19T21:17:42.886Z"),
			want:   str("2026-11-19T00:00:00Z"),
		},
		{
			name:   "same instant in another format keeps prior",
			prior:  str("2026-12-01T00:00:00Z"),
			server: str("2026-12-01T00:00:00.000Z"),
			want:   str("2026-12-01T00:00:00Z"),
		},
		{
			name:   "server moved start into the future is drift",
			prior:  str("2026-11-19T00:00:00Z"),
			server: str("2026-12-01T00:00:00Z"),
			want:   str("2026-12-01T00:00:00Z"),
		},
		{
			name:   "server moved start earlier is drift",
			prior:  str("2026-12-01T00:00:00Z"),
			server: str("2026-11-01T00:00:00Z"),
			want:   str("2026-11-01T00:00:00Z"),
		},
		{
			name:   "unset in config takes server value",
			prior:  types.StringUnknown(),
			server: str("2026-11-19T21:17:42.886Z"),
			want:   str("2026-11-19T21:17:42.886Z"),
		},
		{
			name:   "import with no prior takes server value",
			prior:  types.StringNull(),
			server: str("2026-11-19T21:17:42.886Z"),
			want:   str("2026-11-19T21:17:42.886Z"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := keepStartsAt(tt.prior, tt.server, now)
			if !got.Equal(tt.want) {
				t.Errorf("got %s, want %s", got, tt.want)
			}
		})
	}
}
