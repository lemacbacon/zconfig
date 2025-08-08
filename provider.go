package zconfig

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

// A Provider that implements the repository.Provider interface.
type ArgsProvider struct {
	Args map[string]string
}

// NewArgsProvider lookup keys based on the command-line string.
func NewArgsProvider() (p *ArgsProvider) {
	p = new(ArgsProvider)

	// Initialize the flags map.
	p.Args = make(map[string]string, len(os.Args))

	// For each argument, check if it starts with two dashes. If it does,
	// trim it, split around the first equal sign and set the flag value.
	// If there is no equal sign, and the next argument starts with a
	// double-dash, the flag is added without value, which allows to
	// differentiate between an empty and a non-existing flag.
	for i := 0; i < len(os.Args); i++ {
		arg := os.Args[i]

		if !strings.HasPrefix(arg, "--") {
			continue
		}

		arg = strings.TrimPrefix(arg, "--")
		parts := strings.SplitN(arg, "=", 2)

		if len(parts) == 1 && i+1 < len(os.Args) && !strings.HasPrefix(os.Args[i+1], "--") {
			parts = append(parts, os.Args[i+1])
			i += 1
		}

		parts = append(parts, "") // Avoid out-of-bound errors.
		p.Args[parts[0]] = parts[1]
	}

	return p
}

// Retrieve will return the value from the parsed command-line arguments.
// Arguments are parsed the first time the method is called. Arguments are
// expected to be in the form `--key=value` exclusively (for now).
func (p *ArgsProvider) Retrieve(key string) (value interface{}, found bool, err error) {
	value, found = p.Args[key]
	return value, found, nil
}

// Name of the provider.
func (ArgsProvider) Name() string {
	return "args"
}

// Priority of the provider.
func (ArgsProvider) Priority() int {
	return 1
}

// A Provider that implements the repository.Provider interface.
type EnvProvider struct{}

// NewEnvProvider returns a provider that will lookup keys in the environment
// variables.
func NewEnvProvider() (p EnvProvider) {
	return p
}

// Retrieve will return the value from the parsed environment variables.
// Variables are parsed the first time the method is called.
func (p EnvProvider) Retrieve(key string) (value interface{}, found bool, err error) {
	value, found = os.LookupEnv(p.FormatKey(key))
	return value, found, nil
}

// Name of the provider.
func (EnvProvider) Name() string {
	return "env"
}

// Priority of the provider.
func (EnvProvider) Priority() int {
	return 2
}

func (EnvProvider) FormatKey(key string) (env string) {
	env = strings.ToUpper(key)
	env = strings.Replace(env, ".", "_", -1)
	return strings.Replace(env, "-", "_", -1)
}

// A Provider that implements the repository.Provider interface for dotenv files.
type DotenvProvider struct {
	vars map[string]string
}

// NewDotenvProvider creates a provider that loads environment variables from a .env file
// in the current directory. If the file doesn't exist, the provider will be empty but still functional.
func NewDotenvProvider() *DotenvProvider {
	return NewDotenvProviderWithPath(".env")
}

// NewDotenvProviderWithPath creates a provider that loads environment variables from the specified .env file.
// If the file doesn't exist, the provider will be empty but still functional.
func NewDotenvProviderWithPath(path string) *DotenvProvider {
	p := &DotenvProvider{
		vars: make(map[string]string),
	}

	// Make path absolute to avoid issues with working directory changes
	if !filepath.IsAbs(path) {
		if abs, err := filepath.Abs(path); err == nil {
			path = abs
		}
	}

	p.loadFile(path)
	return p
}

// loadFile loads variables from the specified dotenv file.
func (p *DotenvProvider) loadFile(path string) {
	file, err := os.Open(path)
	if err != nil {
		// File doesn't exist or can't be opened, but that's okay
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Parse KEY=VALUE format
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// Remove surrounding quotes if present
		if len(value) >= 2 {
			if (value[0] == '"' && value[len(value)-1] == '"') ||
				(value[0] == '\'' && value[len(value)-1] == '\'') {
				value = value[1 : len(value)-1]
			}
		}

		p.vars[key] = value
	}
}

// Retrieve will return the value from the loaded dotenv variables.
func (p *DotenvProvider) Retrieve(key string) (value interface{}, found bool, err error) {
	// Use the same key formatting as EnvProvider for consistency
	envKey := p.FormatKey(key)
	value, found = p.vars[envKey]
	return value, found, nil
}

// Name of the provider.
func (p *DotenvProvider) Name() string {
	return "dotenv"
}

// Priority of the provider. Set to 3 so it comes after args (1) and env (2).
func (p *DotenvProvider) Priority() int {
	return 3
}

// FormatKey formats the configuration key to match environment variable naming conventions.
func (p *DotenvProvider) FormatKey(key string) string {
	env := strings.ToUpper(key)
	env = strings.Replace(env, ".", "_", -1)
	return strings.Replace(env, "-", "_", -1)
}
