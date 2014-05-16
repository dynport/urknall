package target

import (
	"testing"
)

func shouldEqualError(t *testing.T, a, b interface{}) {
	t.Errorf("expected %q to equal %q", a, b)
}

func TestNewTarget(t *testing.T) {
	data := map[string]struct {
		user, address string
		port          int
	}{
		"example.com":              {"root", "example.com", 22},
		"example.com:22":           {"root", "example.com", 22},
		"root@example.com":         {"root", "example.com", 22},
		"root@example.com:22":      {"root", "example.com", 22},
		"foo@example.com":          {"foo", "example.com", 22},
		"foo@example.com:22":       {"foo", "example.com", 22},
		"foo@example.com:29":       {"foo", "example.com", 29},
		"foo@wunderscale.com:2222": {"foo", "wunderscale.com", 2222},
	}

	for address, expectation := range data {
		target, e := NewSshTarget(address)
		if e != nil {
			t.Fatalf("failed to parse address: %s", e)
		}

		if target.user != expectation.user {
			shouldEqualError(t, target.user, expectation.user)
		}

		if target.port != expectation.port {
			shouldEqualError(t, target.port, expectation.port)
		}

		if target.address != expectation.address {
			shouldEqualError(t, target.address, expectation.address)
		}
	}
}

func TestNewTargetFailing(t *testing.T) {
	data := map[string]string{
		"":                        "empty address given for target",
		":22":                     "empty address given for target",
		"root@:22":                "empty address given for target",
		"example.com:foobar":      `port must be given as integer, got "foobar"`,
		"example.com:22:23":       `port must be given as integer, got "22:23"`,
		"root@foobar@example.com": "expected target address of the form '<user>@<host>', but was given: root@foobar@example.com",
	}

	for address, expectedError := range data {
		_, e := NewSshTarget(address)
		if e == nil {
			t.Fatalf("address %q should've invoked error %q, but didn't", address, expectedError)
		}

		if e.Error() != expectedError {
			shouldEqualError(t, e.Error(), expectedError)
		}
	}
}
