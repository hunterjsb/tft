package dotenv

import (
	"bufio"
	"os"
	"strings"
	"unicode"
)

// Load reads a .env file and loads the key-value pairs into the environment
func Load(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer func() { _ = file.Close() }()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Split on first = sign
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		value = stripInlineComment(value)
		if key == "" {
			// Skip entries with an empty key (e.g., lines like "=value")
			continue
		}

		// Remove quotes if present (supports both " and ')
		if len(value) >= 2 {
			if (value[0] == '"' && value[len(value)-1] == '"') || (value[0] == '\'' && value[len(value)-1] == '\'') {
				value = value[1 : len(value)-1]
			}
		}

		if err := os.Setenv(key, value); err != nil {
			return err
		}
	}

	return scanner.Err()
}

// stripInlineComment trims inline comments starting with '#' that are not inside quotes.
// A '#' starts a comment when it appears at the beginning of the value or when preceded by whitespace.
// Quoted sections ("..."/'...') are respected.
func stripInlineComment(s string) string {
	inSingle := false
	inDouble := false
	prevIsSpace := true
	for i, r := range s {
		switch r {
		case '\'':
			if !inDouble {
				inSingle = !inSingle
			}
		case '"':
			if !inSingle {
				inDouble = !inDouble
			}
		case '#':
			if !inSingle && !inDouble && (i == 0 || prevIsSpace) {
				return strings.TrimSpace(s[:i])
			}
		}
		prevIsSpace = unicode.IsSpace(r)
	}
	return strings.TrimSpace(s)
}

// LoadDefault loads .env file from the current directory
func LoadDefault() error {
	return Load(".env")
}
