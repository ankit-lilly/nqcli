package desktop

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func AvailableProfiles() []string {
	seen := map[string]bool{}
	for _, name := range parseINISections(configPath()) {
		name = strings.TrimPrefix(name, "profile ")
		if name != "" {
			seen[name] = true
		}
	}
	for _, name := range parseINISections(credentialsPath()) {
		if name != "" {
			seen[name] = true
		}
	}
	profiles := make([]string, 0, len(seen))
	for p := range seen {
		profiles = append(profiles, p)
	}
	sort.Strings(profiles)
	return profiles
}

func configPath() string {
	if p := os.Getenv("AWS_CONFIG_FILE"); p != "" {
		return p
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".aws", "config")
}

func credentialsPath() string {
	if p := os.Getenv("AWS_SHARED_CREDENTIALS_FILE"); p != "" {
		return p
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".aws", "credentials")
}

func parseINISections(path string) []string {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	var sections []string
	for line := range strings.SplitSeq(string(data), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			name := line[1 : len(line)-1]
			sections = append(sections, strings.TrimSpace(name))
		}
	}
	return sections
}
