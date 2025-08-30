package dotenv

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad_Success(t *testing.T) {
	// Create a temporary .env file
	tmpDir, err := os.MkdirTemp("", "dotenv_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() {
		if err := os.RemoveAll(tmpDir); err != nil {
			t.Fatalf("Failed to remove temp dir: %v", err)
		}
	}()

	envFile := filepath.Join(tmpDir, ".env")
	content := `# This is a comment
TEST_KEY1=value1
TEST_KEY2="quoted value"
TEST_KEY3=value with spaces
# Another comment

TEST_KEY4=value4`

	err = os.WriteFile(envFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to write test .env file: %v", err)
	}

	// Load the .env file
	err = Load(envFile)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	// Check that environment variables were set correctly
	tests := []struct {
		key      string
		expected string
	}{
		{"TEST_KEY1", "value1"},
		{"TEST_KEY2", "quoted value"},
		{"TEST_KEY3", "value with spaces"},
		{"TEST_KEY4", "value4"},
	}

	for _, tt := range tests {
		actual := os.Getenv(tt.key)
		if actual != tt.expected {
			t.Errorf("Expected %s=%s, got %s", tt.key, tt.expected, actual)
		}
	}

	// Clean up environment variables
	for _, tt := range tests {
		if err := os.Unsetenv(tt.key); err != nil {
			t.Fatalf("Failed to unset %s: %v", tt.key, err)
		}
	}
}

func TestLoad_FileNotFound(t *testing.T) {
	err := Load("nonexistent.env")
	if err == nil {
		t.Error("Expected error for non-existent file")
	}
}

func TestLoad_EmptyFile(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "dotenv_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() {
		if err := os.RemoveAll(tmpDir); err != nil {
			t.Fatalf("Failed to remove temp dir: %v", err)
		}
	}()

	envFile := filepath.Join(tmpDir, "empty.env")
	err = os.WriteFile(envFile, []byte(""), 0644)
	if err != nil {
		t.Fatalf("Failed to write empty .env file: %v", err)
	}

	err = Load(envFile)
	if err != nil {
		t.Errorf("Load should handle empty file without error: %v", err)
	}
}

func TestLoad_CommentsAndEmptyLines(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "dotenv_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() {
		if err := os.RemoveAll(tmpDir); err != nil {
			t.Fatalf("Failed to remove temp dir: %v", err)
		}
	}()

	envFile := filepath.Join(tmpDir, ".env")
	content := `# Comment at start

# Another comment
TEST_COMMENT=value
# Comment at end`

	err = os.WriteFile(envFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to write test .env file: %v", err)
	}

	err = Load(envFile)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if os.Getenv("TEST_COMMENT") != "value" {
		t.Error("Expected TEST_COMMENT=value")
	}

	if err := os.Unsetenv("TEST_COMMENT"); err != nil {
		t.Fatalf("Failed to unset TEST_COMMENT: %v", err)
	}
}

func TestLoad_InvalidLines(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "dotenv_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() {
		if err := os.RemoveAll(tmpDir); err != nil {
			t.Fatalf("Failed to remove temp dir: %v", err)
		}
	}()

	envFile := filepath.Join(tmpDir, ".env")
	content := `VALID_KEY=value
invalid_line_no_equals
ANOTHER_VALID=another_value
=invalid_empty_key`

	err = os.WriteFile(envFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to write test .env file: %v", err)
	}

	err = Load(envFile)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	// Should only load valid lines
	if os.Getenv("VALID_KEY") != "value" {
		t.Error("Expected VALID_KEY=value")
	}
	if os.Getenv("ANOTHER_VALID") != "another_value" {
		t.Error("Expected ANOTHER_VALID=another_value")
	}

	if err := os.Unsetenv("VALID_KEY"); err != nil {
		t.Fatalf("Failed to unset VALID_KEY: %v", err)
	}
	if err := os.Unsetenv("ANOTHER_VALID"); err != nil {
		t.Fatalf("Failed to unset ANOTHER_VALID: %v", err)
	}
}

func TestLoadDefault_Success(t *testing.T) {
	// Save current working directory
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}
	defer func() {
		_ = os.Chdir(originalWd)
	}()

	// Create temporary directory and change to it
	tmpDir, err := os.MkdirTemp("", "dotenv_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() {
		if err := os.RemoveAll(tmpDir); err != nil {
			t.Fatalf("Failed to remove temp dir: %v", err)
		}
	}()

	err = os.Chdir(tmpDir)
	if err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	// Create .env file in current directory
	content := "DEFAULT_TEST=success"
	err = os.WriteFile(".env", []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to write .env file: %v", err)
	}

	err = LoadDefault()
	if err != nil {
		t.Fatalf("LoadDefault failed: %v", err)
	}

	if os.Getenv("DEFAULT_TEST") != "success" {
		t.Error("Expected DEFAULT_TEST=success")
	}

	if err := os.Unsetenv("DEFAULT_TEST"); err != nil {
		t.Fatalf("Failed to unset DEFAULT_TEST: %v", err)
	}
}

func TestLoadDefault_FileNotFound(t *testing.T) {
	// Save current working directory
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}
	defer func() {
		_ = os.Chdir(originalWd)
	}()

	// Create temporary directory with no .env file
	tmpDir, err := os.MkdirTemp("", "dotenv_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() {
		if err := os.RemoveAll(tmpDir); err != nil {
			t.Fatalf("Failed to remove temp dir: %v", err)
		}
	}()

	err = os.Chdir(tmpDir)
	if err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	err = LoadDefault()
	if err == nil {
		t.Error("Expected error when .env file doesn't exist")
	}
}

func TestLoad_InlineComments(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "dotenv_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() {
		if err := os.RemoveAll(tmpDir); err != nil {
			t.Fatalf("Failed to remove temp dir: %v", err)
		}
	}()

	envFile := filepath.Join(tmpDir, ".env")
	content := `IN_UNQUOTED=value # trailing comment
IN_QUOTED="value # not a comment"
SINGLE_QUOTED='v # not a comment'
LEADING_HASH=   # only comment
MIXED_SPACES=  spaced   # comment
ONLY_COMMENT=    # another`

	err = os.WriteFile(envFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to write test .env file: %v", err)
	}

	err = Load(envFile)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	cases := []struct {
		key, want string
	}{
		{"IN_UNQUOTED", "value"},
		{"IN_QUOTED", "value # not a comment"},
		{"SINGLE_QUOTED", "v # not a comment"},
		{"MIXED_SPACES", "spaced"},
	}

	for _, c := range cases {
		got := os.Getenv(c.key)
		if got != c.want {
			t.Errorf("%s: want %q, got %q", c.key, c.want, got)
		}
	}

	for _, c := range cases {
		if err := os.Unsetenv(c.key); err != nil {
			t.Fatalf("Failed to unset %s: %v", c.key, err)
		}
	}
}
