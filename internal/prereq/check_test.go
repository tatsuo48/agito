package prereq_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/moneyforward/figaro/internal/git"
	"github.com/moneyforward/figaro/internal/prereq"
)

func newOllamaServer(model string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `{"models":[{"name":%q}]}`, model)
	}))
}

func TestCheck_AllPass(t *testing.T) {
	srv := newOllamaServer("gemma4:latest")
	defer srv.Close()

	fake := &git.FakeRunner{Output: "true"}
	cfg := prereq.Config{
		OllamaHost: srv.URL,
		Model:      "gemma4:latest",
	}
	if err := prereq.Check(fake, cfg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCheck_NoChanges(t *testing.T) {
	srv := newOllamaServer("gemma4:latest")
	defer srv.Close()

	callCount := 0
	fake := &git.FuncRunner{
		RunFn: func(name string, args ...string) (string, error) {
			callCount++
			// git status --porcelain returns empty = no changes
			if name == "git" && len(args) > 0 && args[0] == "status" {
				return "", nil
			}
			return "ok", nil
		},
	}
	cfg := prereq.Config{OllamaHost: srv.URL, Model: "gemma4:latest"}
	err := prereq.Check(fake, cfg)
	if err == nil {
		t.Fatal("expected error for no changes")
	}
}
