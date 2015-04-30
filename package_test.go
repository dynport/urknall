package urknall

import (
	"fmt"
	"testing"
)

func shouldBeError(actual interface{}, expected ...interface{}) string {
	if actual == nil {
		return "expected error, got nil"
	}

	err, actIsError := actual.(error)
	msg, msgIsString := expected[0].(string)

	if !actIsError {
		return fmt.Sprintf("expected something of type error\nexpected: %s\n  actual: %s", msg, actual)
	}

	if !msgIsString {
		return fmt.Sprintf("expected value must be of type string: %s", expected[0])
	}

	if err.Error() != msg {
		return fmt.Sprintf("error message did not match\nexpected: %s\n  actual: %s", msg, actual)
	}

	return ""
}

type genericPkg struct {
}

func (p *genericPkg) Render(Package) {
}

func TestBoolValidationRequired(t *testing.T) {
	type pkg struct {
		Field bool `urknall:"required=true"`
		genericPkg
	}

	pi := &pkg{}
	err := validateTemplate(pi)
	if err == nil {
		t.Errorf("expected error, got none")
	} else if err.Error() != `[package:pkg][field:Field] type "bool" doesn't support "required" tag` {
		t.Errorf("got wrong error: %s", err)
	}
}

func TestByteFields(t *testing.T) {
	func() {
		type pkg struct {
			genericPkg
			Field []byte `urknall:"default='a test'"`
		}
		pi := &pkg{}
		err := validateTemplate(pi)
		if err != nil {
			t.Errorf("didn't expect an error, got %q", err)
		}
		if string(pi.Field) != "a test" {
			t.Errorf("expected field to be %q, got %q", "a test", pi.Field)
		}
	}()

	func() {
		type pkg struct {
			genericPkg
			Field []byte `urknall:"required=true"`
		}
		pi := &pkg{}
		if err := validateTemplate(pi); err == nil {
			t.Errorf("expected error, got none")
		}
		pi.Field = []byte("hello world")
		err := validateTemplate(pi)
		if err != nil {
			t.Errorf("didn't expect an error, got %q", err)
		}
	}()
}

func TestBoolValidationDefault(t *testing.T) {
	type pkg struct {
		genericPkg
		Field bool `urknall:"default=false"`
	}

	func() {
		pi := &pkg{}
		err := validateTemplate(pi)
		if err != nil {
			t.Errorf("didn't expect an error, got %q", err)
		}
		if pi.Field != false {
			t.Errorf("expected field to be %t, got %t", false, pi.Field)
		}
	}()

	func() {
		pi := &pkg{Field: false}
		err := validateTemplate(pi)
		if err != nil {
			t.Errorf("didn't expect an error, got %q", err)
		}
		if pi.Field != false {
			t.Errorf("expected field to be %t, got %t", false, pi.Field)
		}
	}()

	func() {
		pi := &pkg{Field: true}
		err := validateTemplate(pi)
		if err != nil {
			t.Errorf("didn't expect an error, got %q", err)
		}
		if pi.Field != true {
			t.Errorf("expected field to be %t, got %t", true, pi.Field)
		}
	}()
}

func TestBoolValidationSize(t *testing.T) {
	type pkg struct {
		genericPkg
		Field bool `urknall:"size=3"`
	}
	pi := &pkg{Field: true}
	err := validateTemplate(pi)

	if err == nil {
		t.Errorf("expected error, got none")
	} else if err.Error() != `[package:pkg][field:Field] type "bool" doesn't support "size" tag` {
		t.Errorf("got wrong error: %s", err)
	}
}

func TestBoolValidationMin(t *testing.T) {
	type pkg struct {
		genericPkg
		Field bool `urknall:"min=3"`
	}
	pi := &pkg{Field: true}
	err := validateTemplate(pi)

	if err == nil {
		t.Errorf("expected error, got none")
	} else if err.Error() != `[package:pkg][field:Field] type "bool" doesn't support "min" tag` {
		t.Errorf("got wrong error: %s", err)
	}
}

func TestBoolValidationMax(t *testing.T) {
	type pkg struct {
		genericPkg
		Field bool `urknall:"max=3"`
	}
	pi := &pkg{Field: true}
	err := validateTemplate(pi)

	if err == nil {
		t.Errorf("expected error, got none")
	} else if err.Error() != `[package:pkg][field:Field] type "bool" doesn't support "max" tag` {
		t.Errorf("got wrong error: %s", err)
	}
}

func TestIntValidationRequired(t *testing.T) {
	type pkg struct {
		genericPkg
		Field int `urknall:"required=true"`
	}

	pi := &pkg{}
	err := validateTemplate(pi)

	if err == nil {
		t.Errorf("expected error, got none")
	} else if err.Error() != `[package:pkg][field:Field] type "int" doesn't support "required" tag` {
		t.Errorf("got wrong error: %s", err)
	}
}

func TestIntValidationDefault(t *testing.T) {
	func() {
		type pkg struct {
			genericPkg
			Field int `urknall:"default=five"`
		}

		pi := &pkg{Field: 1}
		err := validateTemplate(pi)

		if err == nil {
			t.Errorf("expected error, got none")
		} else if err.Error() != `[package:pkg][field:Field] failed to parse value (not an int) of tag "default": "five"` {
			t.Errorf("got wrong error: %s", err)
		}
	}()

	func() {
		type pkg struct {
			genericPkg
			Field int `urknall:"default=5"`
		}

		func() {
			pi := &pkg{}
			validateTemplate(pi)
			if pi.Field != 5 {
				t.Errorf("expected field to be %d, got %d", 5, pi.Field)
			}
		}()

		func() {
			pi := &pkg{Field: 0}
			validateTemplate(pi)
			if pi.Field != 5 {
				t.Errorf("expected field to be %d, got %d", 5, pi.Field)
			}
		}()

		func() {
			pi := &pkg{Field: 42}
			validateTemplate(pi)
			if pi.Field != 42 {
				t.Errorf("expected field to be %d, got %d", 42, pi.Field)
			}
		}()
	}()
}

func TestIntValidationMin(t *testing.T) {
	type pkg struct {
		genericPkg
		Field int `urknall:"min=5"`
	}

	pi := &pkg{}
	err := validateTemplate(pi)

	if err == nil {
		t.Errorf("expected error, got none")
	} else if err.Error() != `[package:pkg][field:Field] value "0" smaller than the specified minimum "5"` {
		t.Errorf("got wrong error: %s", err)
	}

	pi.Field = 4
	err = validateTemplate(pi)
	if err == nil {
		t.Errorf("expected error, got none")
	} else if err.Error() != `[package:pkg][field:Field] value "4" smaller than the specified minimum "5"` {
		t.Errorf("got wrong error: %s", err)
	}

	pi.Field = 5
	err = validateTemplate(pi)
	if err != nil {
		t.Errorf("didn't expect an error, got %q", err)
	}
	if pi.Field != 5 {
		t.Errorf("expected field to be %d, got %d", 5, pi.Field)
	}

	pi.Field = 6
	err = validateTemplate(pi)
	if err != nil {
		t.Errorf("didn't expect an error, got %q", err)
	}
	if pi.Field != 6 {
		t.Errorf("expected field to be %d, got %d", 6, pi.Field)
	}
}

func TestIntValidationMax(t *testing.T) {
	type pkg struct {
		genericPkg
		Field int `urknall:"max=5"`
	}

	pi := &pkg{}
	err := validateTemplate(pi)
	if err != nil {
		t.Errorf("didn't expect an error, got %q", err)
	}

	pi.Field = 4
	err = validateTemplate(pi)
	if err != nil {
		t.Errorf("didn't expect an error, got %q", err)
	}
	if pi.Field != 4 {
		t.Errorf("expected field to be %d, got %d", 4, pi.Field)
	}
	err = validateTemplate(pi)

	pi.Field = 5
	err = validateTemplate(pi)
	if err != nil {
		t.Errorf("didn't expect an error, got %q", err)
	}
	if pi.Field != 5 {
		t.Errorf("expected field to be %d, got %d", 5, pi.Field)
	}

	pi.Field = 6
	err = validateTemplate(pi)
	if err == nil {
		t.Errorf("expected error, got none")
	} else if err.Error() != `[package:pkg][field:Field] value "6" greater than the specified maximum "5"` {
		t.Errorf("got wrong error: %s", err)
	}

}

func TestIntValidationSize(t *testing.T) {
	type pkg struct {
		genericPkg
		Field int `urknall:"size=5"`
	}

	pi := &pkg{}
	err := validateTemplate(pi)
	if err == nil {
		t.Errorf("expected error, got none")
	} else if err.Error() != `[package:pkg][field:Field] type "int" doesn't support "size" tag` {
		t.Errorf("got wrong error: %s", err)
	}
}

func TestStringValidationRequired(t *testing.T) {
	func() {
		type pkg struct {
			genericPkg
			Field string `urknall:"required=tru"`
		}
		pi := &pkg{}
		err := validateTemplate(pi)
		if err == nil {
			t.Errorf("expected error, got none")
		} else if err.Error() != `[package:pkg][field:Field] failed to parse value (neither "true" nor "false") of tag "required": "tru"` {
			t.Errorf("got wrong error: %s", err)
		}
	}()

	func() {
		type pkg struct {
			genericPkg
			Field string `urknall:"required=true"`
		}
		pi := &pkg{}
		err := validateTemplate(pi)
		if err == nil {
			t.Errorf("expected error, got none")
		} else if err.Error() != "[package:pkg][field:Field] required field not set" {
			t.Errorf("got wrong error: %s", err)
		}

		pi = &pkg{Field: ""}
		err = validateTemplate(pi)
		if err == nil {
			t.Errorf("expected error, got none")
		} else if err.Error() != "[package:pkg][field:Field] required field not set" {
			t.Errorf("got wrong error: %s", err)
		}

		pi = &pkg{Field: "something"}
		err = validateTemplate(pi)
		if err != nil {
			t.Errorf("didn't expect an error, got %q", err)
		}
	}()
}

func TestStringValidationDefault(t *testing.T) {
	type pkg struct {
		genericPkg
		Field string `urknall:"default='the 'default' value'"`
	}

	pi := &pkg{}
	err := validateTemplate(pi)
	if err != nil {
		t.Errorf("didn't expect an error, got %q", err)
	} else if pi.Field != "the 'default' value" {
		t.Errorf("did expect field to be %q, got %q", "the 'default' value", pi.Field)
	}

	pi = &pkg{Field: ""}
	err = validateTemplate(pi)
	if err != nil {
		t.Errorf("didn't expect an error, got %q", err)
	} else if pi.Field != "the 'default' value" {
		t.Errorf("did expect field to be %q, got %q", "the 'default' value", pi.Field)
	}

	pi = &pkg{Field: "some other value"}
	err = validateTemplate(pi)
	if err != nil {
		t.Errorf("didn't expect an error, got %q", err)
	} else if pi.Field != "some other value" {
		t.Errorf("did expect field to be %q, got %q", "some other value", pi.Field)
	}
}

func TestStringValidationMinMax(t *testing.T) {
	type pkg struct {
		genericPkg
		Field string `urknall:"min=3 max=4"`
	}
	pi := &pkg{}
	err := validateTemplate(pi)
	if err != nil {
		t.Errorf("didn't expect an error, got %q", err)
	}

	pi = &pkg{Field: "ab"}
	err = validateTemplate(pi)
	if err == nil {
		t.Errorf("expected error, got none")
	} else if err.Error() != `[package:pkg][field:Field] length of value "ab" smaller than the specified minimum length "3"` {
		t.Errorf("got wrong error: %s", err)
	}

	pi = &pkg{Field: "abc"}
	err = validateTemplate(pi)
	if err != nil {
		t.Errorf("didn't expect an error, got %q", err)
	}

	pi = &pkg{Field: "abcd"}
	err = validateTemplate(pi)
	if err != nil {
		t.Errorf("didn't expect an error, got %q", err)
	}

	pi = &pkg{Field: "abcde"}
	err = validateTemplate(pi)
	if err == nil {
		t.Errorf("expected error, got none")
	} else if err.Error() != `[package:pkg][field:Field] length of value "abcde" greater than the specified maximum length "4"` {
		t.Errorf("got wrong error: %s", err)
	}
}

func TestStringValidationSize(t *testing.T) {
	type pkg struct {
		genericPkg
		Field string `urknall:"size=3"`
	}

	pi := &pkg{}
	err := validateTemplate(pi)
	if err != nil {
		t.Errorf("didn't expect an error, got %q", err)
	}

	pi = &pkg{Field: "ab"}
	err = validateTemplate(pi)
	if err == nil {
		t.Errorf("expected error, got none")
	} else if err.Error() != `[package:pkg][field:Field] length of value "ab" doesn't match the specified size "3"` {
		t.Errorf("got wrong error: %s", err)
	}

	pi = &pkg{Field: "abc"}
	err = validateTemplate(pi)
	if err != nil {
		t.Errorf("didn't expect an error, got %q", err)
	}

	pi = &pkg{Field: "abcd"}
	err = validateTemplate(pi)
	if err == nil {
		t.Errorf("expected error, got none")
	} else if err.Error() != `[package:pkg][field:Field] length of value "abcd" doesn't match the specified size "3"` {
		t.Errorf("got wrong error: %s", err)
	}
}

func TestValidationRequiredInvalid(t *testing.T) {
	type pkg struct {
		genericPkg
		Field string `urknall:"required=aberja"`
	}
	pi := &pkg{}
	err := validateTemplate(pi)
	if err == nil {
		t.Errorf("expected error, got none")
	} else if err.Error() != `[package:pkg][field:Field] failed to parse value (neither "true" nor "false") of tag "required": "aberja"` {
		t.Errorf("got wrong error: %s", err)
	}
}

func TestValidationMinInvalid(t *testing.T) {
	type pkg struct {
		genericPkg
		Field string `urknall:"min=..3"`
	}
	pi := &pkg{}
	err := validateTemplate(pi)
	if err == nil {
		t.Errorf("expected error, got none")
	} else if err.Error() != `[package:pkg][field:Field] failed to parse value (not an int) of tag "min": "..3"` {
		t.Errorf("got wrong error: %s", err)
	}
}

func TestValidationMaxInvalid(t *testing.T) {
	type pkg struct {
		genericPkg
		Field string `urknall:"max=4a"`
	}
	pi := &pkg{}
	err := validateTemplate(pi)
	if err == nil {
		t.Errorf("expected error, got none")
	} else if err.Error() != `[package:pkg][field:Field] failed to parse value (not an int) of tag "max": "4a"` {
		t.Errorf("got wrong error: %s", err)
	}
}

func TestValidationSizeInvalid(t *testing.T) {
	type pkg struct {
		genericPkg
		Field string `urknall:"size=4a"`
	}
	pi := &pkg{}
	err := validateTemplate(pi)
	if err == nil {
		t.Errorf("expected error, got none")
	} else if err.Error() != `[package:pkg][field:Field] failed to parse value (not an int) of tag "size": "4a"` {
		t.Errorf("got wrong error: %s", err)
	}
}

func TestMultiTags(t *testing.T) {
	type pkg struct {
		genericPkg
		Field string `urknall:"default='foo' min=3 max=4"`
	}

	pi := &pkg{}
	err := validateTemplate(pi)
	if err != nil {
		t.Errorf("didn't expect an error, got %q", err)
	}
	if pi.Field != "foo" {
		t.Errorf("expect field to be set to %q, got %q", "foo", pi.Field)
	}

	pi = &pkg{Field: "ab"}
	err = validateTemplate(pi)
	if err == nil {
		t.Errorf("expected error, got none")
	} else if err.Error() != `[package:pkg][field:Field] length of value "ab" smaller than the specified minimum length "3"` {
		t.Errorf("got wrong error: %s", err)
	}
}

func TestTagParsing(t *testing.T) {
	func() {
		type pkg struct {
			genericPkg
			Field string `urknall:"required='abc"`
		}

		pi := &pkg{Field: "asd"}
		err := validateTemplate(pi)
		if err == nil {
			t.Errorf("expected error, got none")
		} else if err.Error() != `[package:pkg][field:Field] failed to parse tag due to erroneous quotes` {
			t.Errorf("got wrong error: %s", err)
		}
	}()

	func() {
		type pkg struct {
			genericPkg
			Field string `urknall:"required='ab'c'"`
		}

		pi := &pkg{Field: "asd"}
		err := validateTemplate(pi)
		if err == nil {
			t.Errorf("expected error, got none")
		} else if err.Error() != `[package:pkg][field:Field] failed to parse tag due to erroneous quotes` {
			t.Errorf("got wrong error: %s", err)
		}
	}()

	func() {
		type pkg struct {
			genericPkg
			Field string `urknall:"default"`
		}

		pi := &pkg{Field: "asd"}
		err := validateTemplate(pi)
		if err == nil {
			t.Errorf("expected error, got none")
		} else if err.Error() != `[package:pkg][field:Field] failed to parse annotation (value missing): "default"` {
			t.Errorf("got wrong error: %s", err)
		}
	}()

	func() {
		type pkg struct {
			genericPkg
			Field string `urknall:"defaul='asdf'"`
		}

		pi := &pkg{Field: "asd"}
		err := validateTemplate(pi)
		if err == nil {
			t.Errorf("expected error, got none")
		} else if err.Error() != `[package:pkg][field:Field] tag "defaul" unknown` {
			t.Errorf("got wrong error: %s", err)
		}
	}()
}
