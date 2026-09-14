package awsprofile

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestAvailableFromFilesReturnsOnlySelectableProfiles(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	configFile := writeTestFile(t, dir, "config", `
[default]
region = us-east-1

[profile dsoadev]
sso_session = dsoadev

[sso-session dsoadev]
sso_start_url = https://example.awsapps.com/start

[services local-services]
endpoint_url = http://localhost:4566

[profile campstaging]
sso_session = campstaging

[profile ]
`)
	credentialsFile := writeTestFile(t, dir, "credentials", `
[legacy]
aws_access_key_id = key

[dsoadev]
aws_access_key_id = duplicate
`)

	got := availableFromFiles(configFile, credentialsFile)
	want := []string{"campstaging", "default", "dsoadev", "legacy"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("available profiles = %v, want %v", got, want)
	}
}

func TestAvailableFromFilesIgnoresMissingFiles(t *testing.T) {
	t.Parallel()
	missing := filepath.Join(t.TempDir(), "missing")
	if got := availableFromFiles(missing, missing); len(got) != 0 {
		t.Fatalf("available profiles = %v, want none", got)
	}
}

func writeTestFile(t *testing.T, dir, name, contents string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}
