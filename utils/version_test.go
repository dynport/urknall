package utils

import (
	"testing"
)

func TestParseVersion(t *testing.T) {
	v := &Version{}
	if err := v.Parse("1.2.3"); err != nil {
		t.Errorf("didn't expect an error, got %q", err)
	}

	if v.Major != 1 {
		t.Errorf("expected major version to be %d, got %d", 1, v.Major)
	}
	if v.Minor != 2 {
		t.Errorf("expected major version to be %d, got %d", 2, v.Minor)
	}
	if v.Patch != 3 {
		t.Errorf("expected major version to be %d, got %d", 3, v.Patch)
	}
}

func TestCompareSmallerVersion(t *testing.T) {
	a, err := ParseVersion("0.1.2")
	if err != nil {
		t.Errorf("didn't expect an error, got %q", err)
	}

	for _, v := range []string{"0.1.3", "0.2.0", "1.0.0"} {
		b, err := ParseVersion(v)
		if err != nil {
			t.Errorf("didn't expect an error, got %q", err)
		}
		if !a.Smaller(b) {
			t.Errorf("expected %q to be smaller than %q, wasn't!", a, b)
		}
	}
}

func TestCompareEqualVersions(t *testing.T) {
	a, err := ParseVersion("0.1.2")
	if err != nil {
		t.Errorf("didn't expect an error, got %q", err)
	}

	b, err := ParseVersion("0.1.2")
	if err != nil {
		t.Errorf("didn't expect an error, got %q", err)
	}
	if a.Smaller(b) {
		t.Errorf("expected %q to be smaller than %q, wasn't!", a, b)
	}
}
