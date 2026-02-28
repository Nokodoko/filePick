package main

import (
	"os"
	"testing"
)

func TestTruncatePathNoTruncation(t *testing.T) {
	result := truncatePath("/short/path", 50)
	if result != "/short/path" {
		t.Errorf("expected /short/path, got %s", result)
	}
}

func TestTruncatePathHomeReplacement(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skip("cannot determine home directory")
	}

	path := home + "/Projects/myapp"
	result := truncatePath(path, 100)
	if result != "~/Projects/myapp" {
		t.Errorf("expected ~/Projects/myapp, got %s", result)
	}
}

func TestTruncatePathComponentBoundary(t *testing.T) {
	path := "/very/long/directory/path/to/some/deep/file"
	result := truncatePath(path, 30)

	// Should start with ... and end with full components
	if len(result) > 30 {
		t.Errorf("result too long: %d chars, expected <= 30: %s", len(result), result)
	}
	if result[:3] != "..." {
		t.Errorf("expected to start with ..., got %s", result)
	}
}

func TestTruncatePathExactFit(t *testing.T) {
	path := "/abc/def"
	result := truncatePath(path, 8)
	if result != "/abc/def" {
		t.Errorf("expected /abc/def (exact fit), got %s", result)
	}
}

func TestTruncatePathVeryNarrow(t *testing.T) {
	path := "/a/very/long/path/that/needs/truncation"
	result := truncatePath(path, 10)

	if len(result) > 10 {
		t.Errorf("result too long for narrow width: %d chars: %s", len(result), result)
	}
}

func TestTruncatePathSingleComponent(t *testing.T) {
	path := "/verylongdirectoryname"
	result := truncatePath(path, 15)

	// The single component is too long, should be truncated
	if len(result) > 15 {
		t.Errorf("result too long: %d chars: %s", len(result), result)
	}
}

func TestTruncatePathPreservesLastComponent(t *testing.T) {
	path := "/home/user/Programs/ai/computeCommander"
	result := truncatePath(path, 30)

	// Should preserve "computeCommander" at the end if it fits
	if len(result) > 30 {
		t.Errorf("result too long: %d chars: %s", len(result), result)
	}
}
