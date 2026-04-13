package docker

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
)

type buildOutputMessage struct {
	Stream      string `json:"stream,omitempty"`
	Status      string `json:"status,omitempty"`
	Progress    string `json:"progress,omitempty"`
	Error       string `json:"error,omitempty"`
	ErrorDetail *struct {
		Message string `json:"message,omitempty"`
	} `json:"errorDetail,omitempty"`
}

func ConsumeBuildOutput(r io.Reader, w io.Writer) error {
	decoder := json.NewDecoder(r)

	for {
		var msg buildOutputMessage
		if err := decoder.Decode(&msg); err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			return fmt.Errorf("failed to decode build output: %w", err)
		}

		switch {
		case msg.Error != "":
			if msg.ErrorDetail != nil && msg.ErrorDetail.Message != "" {
				return fmt.Errorf("docker build failed: %s", msg.ErrorDetail.Message)
			}
			return fmt.Errorf("docker build failed: %s", msg.Error)
		case msg.Stream != "":
			if _, err := io.WriteString(w, msg.Stream); err != nil {
				return fmt.Errorf("failed to write build output: %w", err)
			}
		case msg.Status != "":
			line := msg.Status
			if msg.Progress != "" {
				line += " " + msg.Progress
			}
			line += "\n"
			if _, err := io.WriteString(w, line); err != nil {
				return fmt.Errorf("failed to write build output: %w", err)
			}
		}
	}
}
