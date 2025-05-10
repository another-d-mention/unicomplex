package dotenv

import (
	"bytes"
	"os"
	"strings"

	"github.com/another-d-mention/unicomplex/filesystem"
)

// Load reads the environment variables from a file and returns the variables in a map and sets them
// in the current process so they are accessible with os.Getenv.
func Load(filename string) (map[string]string, error) {
	f, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return parse(f)
}

// LoadFS reads the environment variables from a file and returns the variables in a map and sets them
// in the current process so they are accessible with os.Getenv.
func LoadFS(filesystem filesystem.FileSystem, filename string) (map[string]string, error) {
	data, err := filesystem.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return parse(data)
}

func parse(data []byte) (map[string]string, error) {
	lines := bytes.Split(data, []byte("\n"))
	env := make(map[string]string)
	for _, line := range lines {
		line = bytes.TrimSpace(line)
		if len(line) == 0 || line[0] == '#' {
			continue
		}
		parts := bytes.SplitN(line, []byte("="), 2)
		if len(parts) != 2 {
			continue
		}
		key := string(bytes.TrimSpace(parts[0]))
		value := string(bytes.TrimSpace(parts[1]))
		if strings.HasPrefix(value, `"`) && strings.HasSuffix(value, `"`) {
			value = strings.Trim(value, `"`)
		}
		env[key] = value
		if err := os.Setenv(key, value); err != nil {
			return nil, err
		}
	}
	return env, nil
}
