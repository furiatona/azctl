package templatex

import (
	"context"
	"testing"

	"github.com/furiatona/azctl/internal/config"
)

func TestRenderEnv(t *testing.T) {
	// init config with env var in process
	t.Setenv("FOO", "bar")
	// minimal init
	_ = config.Init(context.TODO(), "")
	cfg := config.Current()

	input := `{"x":"{{ env "FOO" }}"}`
	out, err := RenderEnv(input, cfg)
	if err != nil {
		t.Fatalf("render failed: %v", err)
	}
	want := `{"x":"bar"}`
	if out != want {
		t.Fatalf("got %q want %q", out, want)
	}
}
