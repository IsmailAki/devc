package baseimage

import (
	"strings"
	"testing"
)

func TestBaseDockerfileAllowsRootSSH(t *testing.T) {
	if !strings.Contains(BaseDockerfile, "PermitRootLogin prohibit-password") {
		t.Fatal("expected root SSH login to be enabled with key-only auth")
	}

	if !strings.Contains(BaseDockerfile, "AllowUsers dev root") {
		t.Fatal("expected both dev and root SSH users to be allowed")
	}
}
