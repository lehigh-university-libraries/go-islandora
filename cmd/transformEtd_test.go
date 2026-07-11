package cmd

import "testing"

func TestParseCompletionYear(t *testing.T) {
	tests := []struct {
		name           string
		completionDate string
		want           string
		wantErr        bool
	}{
		{
			name:           "year and month",
			completionDate: "2009-05",
			want:           "2009",
		},
		{
			name:           "year only",
			completionDate: "2010",
			want:           "2010",
		},
		{
			name:           "trims whitespace",
			completionDate: " 2009 ",
			want:           "2009",
		},
		{
			name:           "invalid year only",
			completionDate: "20AB",
			wantErr:        true,
		},
		{
			name:           "invalid month",
			completionDate: "2009-13",
			wantErr:        true,
		},
		{
			name:           "empty",
			completionDate: "",
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseCompletionYear(tt.completionDate)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("parseCompletionYear(%q) expected error", tt.completionDate)
				}
				return
			}
			if err != nil {
				t.Fatalf("parseCompletionYear(%q) unexpected error: %v", tt.completionDate, err)
			}
			if got != tt.want {
				t.Fatalf("parseCompletionYear(%q) = %q, want %q", tt.completionDate, got, tt.want)
			}
		})
	}
}
