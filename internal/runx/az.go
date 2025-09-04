package runx

import (
	"context"
	"fmt"
	"os"
	"os/exec"
)

func AZ(ctx context.Context, args ...string) error {
	cmd := exec.CommandContext(ctx, "az", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("az command failed: %w", err)
	}
	return nil
}

// AZOutput runs az command and returns the output as a string
func AZOutput(ctx context.Context, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, "az", args...)
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("az command failed: %w", err)
	}
	return string(output), nil
}
