package engines

import (
	"net/url"
	"testing"
)

func TestFormatURL(t *testing.T) {
	checkURL(t, "swift://foo", "swift:foo")
	checkURL(t, "s3://foo", "s3:foo")
	checkURL(t, "swift://foo/bar", "swift:foo/bar")
	checkURL(t, "s3://foo/bar", "s3:foo/bar")
	checkURL(t, "s3+http://foo", "s3:foo")
	checkURL(t, "s3+http://foo/bar", "s3:foo/bar")
	checkURL(t, "s3://sos.io.exo/foo", "s3:foo")
	checkURL(t, "s3://sos.io.exo/foo/bar", "s3:foo/bar")
	checkURL(t, "s3+http://sos.io.exo/foo", "s3:foo")
	checkURL(t, "s3+http://sos.io.exo/foo/bar", "s3:foo/bar")
}

func checkURL(t *testing.T, uri, expected string) {
	u, err := url.Parse(uri)
	if err != nil {
		t.Fatalf("Failed to parse URL %s", uri)
	}

	formatURL(u)
	if u.String() != expected {
		t.Fatalf("Expected %s, got %s", expected, u.String())
	}
	return
}
