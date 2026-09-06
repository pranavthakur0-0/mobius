package guides

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestScanWorkspace(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test files
	files := []string{
		"main.go",
		"utils.go",
		"go.mod",
		"Makefile",
		"Dockerfile",
		"LICENSE",
		"scripts/run.sh",
		"node_modules/package.json", 
		".git/config",              
	}

	for _, f := range files {
		fullPath := filepath.Join(tmpDir, f)
		err := os.MkdirAll(filepath.Dir(fullPath), 0755)
		if err != nil {
			t.Fatalf("failed to create dir: %v", err)
		}
		err = os.WriteFile(fullPath, []byte("content"), 0644)
		if err != nil {
			t.Fatalf("failed to create file: %v", err)
		}
	}

	manifests, exts, err := ScanWorkspace(tmpDir)
	if err != nil {
		t.Fatalf("ScanWorkspace returned error: %v", err)
	}

	// Verify manifests found
	expectedManifests := map[string]bool{
		"Dockerfile": true,
		"Makefile":   true,
		"go.mod":     true,
	}

	for _, m := range manifests {
		delete(expectedManifests, m)
	}
	if len(expectedManifests) > 0 {
		t.Errorf("missing expected manifests: %v", expectedManifests)
	}

	// Verify extensions found (.go, .sh)
	expectedExts := map[string]bool{
		".go": true,
		".sh": true,
	}
	for _, e := range exts {
		delete(expectedExts, e)
	}
	if len(expectedExts) > 0 {
		t.Errorf("missing expected extensions: %v", expectedExts)
	}
}

func TestBuildInitPrompt_Fallbacks(t *testing.T) {
	// Case 1: normal input
	p1 := BuildInitPrompt([]string{"go.mod"}, []string{".go"})
	if !strings.Contains(p1, "go.mod") || !strings.Contains(p1, ".go") {
		t.Errorf("expected go.mod and .go in prompt, got: %s", p1)
	}

	// Case 2: no extensions found
	p2 := BuildInitPrompt([]string{"Makefile"}, nil)
	if !strings.Contains(p2, "File extensions present: (none detected)") {
		t.Errorf("expected '(none detected)' fallback for extensions, got: %s", p2)
	}

	// Case 3: neither manifests nor extensions
	p3 := BuildInitPrompt(nil, nil)
	if !strings.Contains(p3, "Manifest/Config files found: (none detected)") {
		t.Errorf("expected '(none detected)' fallback for manifests, got: %s", p3)
	}
}

func TestCleanGeneratedMarkdown(t *testing.T) {
	inputWithTicks := "```markdown\n## Verification Commands\nBuild: go build ./...\n```"
	cleaned := cleanGeneratedMarkdown(inputWithTicks)
	expected := "## Verification Commands\nBuild: go build ./..."
	if cleaned != expected {
		t.Errorf("expected %q, got %q", expected, cleaned)
	}

	plainInput := "## Verification Commands\nBuild: make"
	if cleanGeneratedMarkdown(plainInput) != plainInput {
		t.Errorf("expected plain input preserved")
	}
}

func TestAppendOrUpdateVerificationCommands(t *testing.T) {
	newCmds := "## Verification Commands\nBuild: go build ./...\nTest: go test ./..."

	// Case 1: Empty file
	res1 := appendOrUpdateVerificationCommands("", newCmds)
	if !strings.HasPrefix(res1, newCmds) {
		t.Errorf("expected new commands at top of empty content, got %s", res1)
	}

	// Case 2: File with existing content but no verification commands
	existing := "# Project Info\nSome description."
	res2 := appendOrUpdateVerificationCommands(existing, newCmds)
	if !strings.Contains(res2, existing) || !strings.Contains(res2, newCmds) {
		t.Errorf("expected both sections in %s", res2)
	}

	// Case 3: File with existing verification commands gets updated idempotently
	existingWithCmds := "# Project Info\n\n## Verification Commands\nBuild: old build\n\n## Next Section\nMore stuff."
	res3 := appendOrUpdateVerificationCommands(existingWithCmds, newCmds)
	if strings.Contains(res3, "old build") {
		t.Errorf("expected 'old build' to be replaced, got %s", res3)
	}
	if !strings.Contains(res3, "go build ./...") || !strings.Contains(res3, "## Next Section") {
		t.Errorf("expected replacement to preserve subsequent sections, got %s", res3)
	}
}
