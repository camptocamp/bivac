package engines

import (
	"net/url"
	"testing"
)

func TestFormatURL(t *testing.T) {
	checkURL(t, "swift://foo", "swift:foo", []string{})
	checkURL(t, "s3://foo", "s3:foo", []string{})
	checkURL(t, "swift://foo/bar", "swift:foo/bar", []string{})
	checkURL(t, "s3://foo/bar", "s3:foo/bar", []string{})
	checkURL(t, "s3+http://foo", "s3:foo", []string{})
	checkURL(t, "s3+http://foo/bar", "s3:foo/bar", []string{})
	checkURL(t, "s3://sos.io.exo/foo", "s3:foo", []string{"AWS_ENDPOINT=sos.io.exo"})
	checkURL(t, "s3://sos.io.exo/foo/bar", "s3:foo/bar", []string{"AWS_ENDPOINT=sos.io.exo"})
	checkURL(t, "s3+http://sos.io.exo/foo", "s3:foo", []string{"AWS_ENDPOINT=sos.io.exo"})
	checkURL(t, "s3+http://sos.io.exo/foo/bar", "s3:foo/bar", []string{"AWS_ENDPOINT=sos.io.exo"})
}

func checkURL(t *testing.T, uri, expected string, expectedEnv []string) {
	u, err := url.Parse(uri)
	if err != nil {
		t.Fatalf("Failed to parse URL %s", uri)
	}

	env := formatURL(u)
	if u.String() != expected {
		t.Fatalf("Expected %s, got %s", expected, u.String())
	}

	if len(expectedEnv) == 0 {
		return
	}

	if len(env) != len(expectedEnv) {
		t.Fatalf("Expected size %v, got %v", len(expectedEnv), len(env))
	}

	if env[0] != expectedEnv[0] {
		t.Fatalf("Expected %v, got %v", expectedEnv, env)
	}
	return
}
