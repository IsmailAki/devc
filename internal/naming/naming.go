package naming

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/big"
	"path/filepath"
	"regexp"
	"strings"
)

const (
	Prefix              = "devc-"
	RandomIDLength      = 5
	LocalNameHashLength = 8
	charset             = "abcdefghijklmnopqrstuvwxyz0123456789"
)

var invalidNameChars = regexp.MustCompile(`[^a-z0-9]+`)

func generateRandomID() string {
	b := make([]byte, RandomIDLength)
	for i := range b {
		n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		b[i] = charset[n.Int64()]
	}
	return string(b)
}

func sanitizeSegment(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	value = invalidNameChars.ReplaceAllString(value, "-")
	value = strings.Trim(value, "-")
	if value == "" {
		return "project"
	}
	return value
}

func GenerateContainerName(owner, repo string) string {
	randomID := generateRandomID()
	hasOwner := strings.TrimSpace(owner) != ""
	owner = sanitizeSegment(owner)
	repo = sanitizeSegment(repo)

	if hasOwner {
		return fmt.Sprintf("%s%s-%s-%s", Prefix, owner, repo, randomID)
	}

	return fmt.Sprintf("%s%s-%s", Prefix, repo, randomID)
}

func GenerateLocalContainerName(projectName, projectPath string) string {
	projectName = sanitizeSegment(projectName)
	cleanPath := filepath.Clean(projectPath)
	hash := sha256.Sum256([]byte(cleanPath))
	suffix := hex.EncodeToString(hash[:])[:LocalNameHashLength]
	return fmt.Sprintf("%s%s-%s", Prefix, projectName, suffix)
}

func GenerateVolumeName(containerName string) string {
	return containerName
}
