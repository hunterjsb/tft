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
	defer os.RemoveAll(tmpDir)

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
		os.Unsetenv(tt.key)
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
	defer os.RemoveAll(tmpDir)

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
	defer os.RemoveAll(tmpDir)

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

	os.Unsetenv("TEST_COMMENT")
}

func TestLoad_InvalidLines(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "dotenv_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

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

	os.Unsetenv("VALID_KEY")
	os.Unsetenv("ANOTHER_VALID")
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
	defer os.RemoveAll(tmpDir)

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

	os.Unsetenv("DEFAULT_TEST")
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
	defer os.RemoveAll(tmpDir)

	err = os.Chdir(tmpDir)
	if err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	err = LoadDefault()
	if err == nil {
		t.Error("Expected error when .env file doesn't exist")
	}
}
