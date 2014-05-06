package ssh

import (
	"testing"
)

func shouldEqualError(t *testing.T, a, b interface{}) {
	t.Errorf("expected %q to equal %q", a, b)
}

func TestNewHost(t *testing.T) {
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
		host, e := New(address)
		if e != nil {
			t.Fatalf("failed to parse address: %s", e)
		}

		if host.user != expectation.user {
			shouldEqualError(t, host.user, expectation.user)
		}

		if host.port != expectation.port {
			shouldEqualError(t, host.port, expectation.port)
		}

		if host.address != expectation.address {
			shouldEqualError(t, host.address, expectation.address)
		}
	}
}

func TestNewHostFailing(t *testing.T) {
	data := map[string]string{
		"":                        "empty address given for host",
		":22":                     "empty address given for host",
		"root@:22":                "empty address given for host",
		"example.com:foobar":      `port must be given as integer, got "foobar"`,
		"example.com:22:23":       `port must be given as integer, got "22:23"`,
		"root@foobar@example.com": "expected host address of the form '<user>@<host>', but was given: root@foobar@example.com",
	}

	for address, expectedError := range data {
		_, e := New(address)
		if e == nil {
			t.Fatalf("address %q should've invoked error %q, but didn't", address, expectedError)
		}

		if e.Error() != expectedError {
			shouldEqualError(t, e.Error(), expectedError)
		}
	}
}
