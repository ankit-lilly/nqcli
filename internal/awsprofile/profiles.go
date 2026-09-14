package awsprofile

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Available returns the selectable profiles from the AWS shared config files.
func Available() []string {
	return availableFromFiles(configPath(), credentialsPath())
}

func availableFromFiles(configFile, credentialsFile string) []string {
	seen := make(map[string]struct{})
	for _, section := range iniSections(configFile) {
		if profile, ok := configProfile(section); ok {
			seen[profile] = struct{}{}
		}
	}
	for _, section := range iniSections(credentialsFile) {
		if profile := strings.TrimSpace(section); profile != "" {
			seen[profile] = struct{}{}
		}
	}

	profiles := make([]string, 0, len(seen))
	for profile := range seen {
		profiles = append(profiles, profile)
	}
	sort.Strings(profiles)
	return profiles
}

func configProfile(section string) (string, bool) {
	section = strings.TrimSpace(section)
	if section == "default" {
		return section, true
	}
	if profile, ok := strings.CutPrefix(section, "profile "); ok {
		profile = strings.TrimSpace(profile)
		return profile, profile != ""
	}
	return "", false
}

func configPath() string {
	if path := os.Getenv("AWS_CONFIG_FILE"); path != "" {
		return path
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".aws", "config")
}

func credentialsPath() string {
	if path := os.Getenv("AWS_SHARED_CREDENTIALS_FILE"); path != "" {
		return path
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".aws", "credentials")
}

func iniSections(path string) []string {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}

	var sections []string
	for line := range strings.SplitSeq(string(data), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			sections = append(sections, strings.TrimSpace(line[1:len(line)-1]))
		}
	}
	return sections
}
