package urknall

import (
	"testing"
)

type testCommandCustomChecksum struct {
	*testCommand
}

func (c *testCommandCustomChecksum) Checksum() string {
	return "default checksum"
}

func TestUtils(t *testing.T) {
	c1 := &testCommand{}
	if checksum, err := commandChecksum(c1); err != nil {
		t.Errorf("didn't expect an error, got %q", err)
	} else if checksum != "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855" {
		t.Errorf("expected %q, got %q", "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855", checksum)
	}

	c2 := &testCommandCustomChecksum{c1}
	if checksum, err := commandChecksum(c2); err != nil {
		t.Errorf("didn't expect an error, got %q", err)
	} else if checksum != "default checksum" {
		t.Errorf("expected %q, got %q", "default checksum", checksum)
	}
}

func TestMidTruncate(t *testing.T) {
	tests := []struct {
		In  string
		Len int
		Out string
	}{
		{"test", 4, "test"},
		{"abcdef", 5, "a...f"},
		{"abcdefg", 6, "ab...g"},
	}

	for _, tst := range tests {
		if v, ex := midTrunc(tst.In, tst.Len), tst.Out; ex != v {
			t.Errorf("expected midTrunc of %q with len %d to be %q, was %q", tst.In, tst.Len, ex, v)
		}
	}
}
