package claude

import "os"

// ReadRawJSONL reads a JSONL file and returns the raw bytes.
func ReadRawJSONL(path string) ([]byte, error) {
	return os.ReadFile(path)
}
