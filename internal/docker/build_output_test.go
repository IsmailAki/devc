package docker

import (
	"bytes"
	"strings"
	"testing"
)

func TestConsumeBuildOutputSuccess(t *testing.T) {
	input := strings.NewReader("{\"stream\":\"step 1\\n\"}\n{\"status\":\"pulling\",\"progress\":\"50%\"}\n")
	var output bytes.Buffer

	err := ConsumeBuildOutput(input, &output)
	if err != nil {
		t.Fatalf("ConsumeBuildOutput() error = %v", err)
	}

	got := output.String()
	if !strings.Contains(got, "step 1") {
		t.Fatalf("expected stream output, got %q", got)
	}
	if !strings.Contains(got, "pulling 50%") {
		t.Fatalf("expected status output, got %q", got)
	}
}

func TestConsumeBuildOutputFailure(t *testing.T) {
	input := strings.NewReader("{\"errorDetail\":{\"message\":\"exit code: 100\"},\"error\":\"build failed\"}\n")

	err := ConsumeBuildOutput(input, &bytes.Buffer{})
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !strings.Contains(err.Error(), "exit code: 100") {
		t.Fatalf("expected detailed build error, got %v", err)
	}
}
