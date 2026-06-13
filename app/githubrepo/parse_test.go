package githubrepo

import "testing"

func TestParseGitHubRepositoryURL(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want Repository
	}{
		{
			name: "https",
			raw:  "https://github.com/openai/codex",
			want: Repository{Owner: "openai", Repo: "codex", URL: "https://github.com/openai/codex"},
		},
		{
			name: "https git suffix",
			raw:  "https://github.com/openai/codex.git",
			want: Repository{Owner: "openai", Repo: "codex", URL: "https://github.com/openai/codex"},
		},
		{
			name: "ssh",
			raw:  "git@github.com:openai/codex.git",
			want: Repository{Owner: "openai", Repo: "codex", URL: "https://github.com/openai/codex"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Parse(tt.raw)
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}
			if got != tt.want {
				t.Fatalf("Parse() = %#v, want %#v", got, tt.want)
			}
		})
	}
}

func TestParseRejectsNonGitHubURL(t *testing.T) {
	if _, err := Parse("https://gitlab.com/openai/codex"); err == nil {
		t.Fatal("Parse() expected an error")
	}
}
