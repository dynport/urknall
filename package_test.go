package urknall

import (
	"fmt"
	"testing"
	. "github.com/smartystreets/goconvey/convey"
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
	Convey("Given a package with a bool field that is required", t, func() {
		type pkg struct {
			Field bool `urknall:"required=true"`
			genericPkg
		}

		Convey("When an instance is created without value set", func() {
			pi := &pkg{}
			Convey("Then validation must return an error", func() {
				So(validateTemplate(pi), shouldBeError, `[package:pkg][field:Field] type "bool" doesn't support "required" tag`)
			})
		})
	})
}

func TestByteFields(t *testing.T) {
	Convey("Given a package with a []byte field with a default value", t, func() {
		type pkg struct {
			genericPkg
			Field []byte `urknall:"default='a test'"`
		}
		pi := &pkg{}
		So(validateTemplate(pi), ShouldBeNil)
		So(string(pi.Field), ShouldEqual, "a test")
	})

	Convey("Given a package with a []byte field with a required tag", t, func() {
		type pkg struct {
			genericPkg
			Field []byte `urknall:"required=true"`
		}
		pi := &pkg{}
		So(validateTemplate(pi), ShouldNotBeNil)
		pi.Field = []byte("hello world")
		So(validateTemplate(pi), ShouldBeNil)
	})
}

func TestBoolValidationDefault(t *testing.T) {
	Convey("Given a package with a bool field with a default value", t, func() {
		type pkg struct {
			genericPkg
			Field bool `urknall:"default=false"`
		}

		Convey("When an instance is created with value not set", func() {
			pi := &pkg{}
			Convey("Then validation must succeed", func() {
				So(validateTemplate(pi), ShouldBeNil)
			})
			Convey("Then value must be set to default", func() {
				So(pi.Field, ShouldEqual, false)
			})
		})

		Convey("When an instance is created with value set to false", func() {
			pi := &pkg{Field: false}
			Convey("Then validation must succeed", func() {
				So(validateTemplate(pi), ShouldBeNil)
			})
			Convey("Then value must be set to false", func() {
				So(pi.Field, ShouldEqual, false)
			})
		})

		Convey("When an instance is created with value set to true", func() {
			pi := &pkg{Field: true}
			Convey("Then validation must succeed", func() {
				So(validateTemplate(pi), ShouldBeNil)
			})
			Convey("Then value must be set to true", func() {
				So(pi.Field, ShouldEqual, true)
			})
		})
	})
}

func TestBoolValidationSize(t *testing.T) {
	Convey("Given a package with a size tag", t, func() {
		type pkg struct {
			genericPkg
			Field bool `urknall:"size=3"`
		}
		Convey("When an instance is created", func() {
			pi := &pkg{Field: true}
			Convey("Then validation must fail", func() {
				So(validateTemplate(pi), shouldBeError, `[package:pkg][field:Field] type "bool" doesn't support "size" tag`)
			})
		})
	})
}

func TestBoolValidationMin(t *testing.T) {
	Convey("Given a package with a min tag", t, func() {
		type pkg struct {
			genericPkg
			Field bool `urknall:"min=3"`
		}
		Convey("When an instance is created", func() {
			pi := &pkg{Field: true}
			Convey("Then validation must fail", func() {
				So(validateTemplate(pi), shouldBeError, `[package:pkg][field:Field] type "bool" doesn't support "min" tag`)
			})
		})
	})
}

func TestBoolValidationMax(t *testing.T) {
	Convey("Given a package with a max tag", t, func() {
		type pkg struct {
			genericPkg
			Field bool `urknall:"max=3"`
		}
		Convey("When an instance is created", func() {
			pi := &pkg{Field: true}
			Convey("Then validation must fail", func() {
				So(validateTemplate(pi), shouldBeError, `[package:pkg][field:Field] type "bool" doesn't support "max" tag`)
			})
		})
	})
}

func TestIntValidationRequired(t *testing.T) {
	Convey("Given a package with a int field that is required", t, func() {
		type pkg struct {
			genericPkg
			Field int `urknall:"required=true"`
		}

		Convey("When an instance is created", func() {
			pi := &pkg{}
			Convey("Then validation must return an error", func() {
				So(validateTemplate(pi), shouldBeError, `[package:pkg][field:Field] type "int" doesn't support "required" tag`)
			})
		})
	})
}

func TestIntValidationDefault(t *testing.T) {
	Convey("Given a package with a int field that has an erroneous default tag", t, func() {
		type pkg struct {
			genericPkg
			Field int `urknall:"default=five"`
		}

		pi := &pkg{Field: 1}
		Convey("Then validation must fail", func() {
			So(validateTemplate(pi), shouldBeError, `[package:pkg][field:Field] failed to parse value (not an int) of tag "default": "five"`)
		})
	})

	Convey("Given a package with a int field that has a default value", t, func() {
		type pkg struct {
			genericPkg
			Field int `urknall:"default=5"`
		}

		Convey("When an instance is created without specifying a value", func() {
			pi := &pkg{}
			Convey("Then validation must set the value", func() {
				validateTemplate(pi)
				So(pi.Field, ShouldEqual, 5)
			})
		})

		Convey("When an instance is created with an empty value specified", func() {
			pi := &pkg{Field: 0}
			Convey("Then validation must set the value", func() {
				validateTemplate(pi)
				So(pi.Field, ShouldEqual, 5)
			})
		})

		Convey("When an instance is created with a value specified", func() {
			pi := &pkg{Field: 42}
			Convey("Then validation must not touch the set value", func() {
				validateTemplate(pi)
				So(pi.Field, ShouldEqual, 42)
			})
		})
	})
}

func TestIntValidationMin(t *testing.T) {
	Convey("Given a package with a int field that has a min value", t, func() {
		type pkg struct {
			genericPkg
			Field int `urknall:"min=5"`
		}

		Convey("When an instance is created without specifying a value", func() {
			pi := &pkg{}
			Convey("Then validation must succeed", func() {
				So(validateTemplate(pi), shouldBeError, `[package:pkg][field:Field] value "0" smaller than the specified minimum "5"`)
			})
		})

		Convey("When an instance is created with a value smaller than min value", func() {
			pi := &pkg{Field: 4}
			Convey("Then validation must fail", func() {
				So(validateTemplate(pi), shouldBeError, `[package:pkg][field:Field] value "4" smaller than the specified minimum "5"`)
			})
		})

		Convey("When an instance is created with a value equal to the min value", func() {
			pi := &pkg{Field: 5}
			Convey("Then validation must succeed", func() {
				So(validateTemplate(pi), ShouldBeNil)
			})
		})

		Convey("When an instance is created with a value greater than the min value", func() {
			pi := &pkg{Field: 6}
			Convey("Then validation must succeed", func() {
				So(validateTemplate(pi), ShouldBeNil)
			})
		})
	})
}

func TestIntValidationMax(t *testing.T) {
	Convey("Given a package with a int field that has a max value", t, func() {
		type pkg struct {
			genericPkg
			Field int `urknall:"max=5"`
		}

		Convey("When an instance is created without specifying a value", func() {
			pi := &pkg{}
			Convey("Then validation must succeed", func() {
				So(validateTemplate(pi), ShouldBeNil)
			})
		})

		Convey("When an instance is created with a value smaller than max value", func() {
			pi := &pkg{Field: 4}
			Convey("Then validation must succeed", func() {
				So(validateTemplate(pi), ShouldBeNil)
			})
		})

		Convey("When an instance is created with a value equal to the max value", func() {
			pi := &pkg{Field: 5}
			Convey("Then validation must succeed", func() {
				So(validateTemplate(pi), ShouldBeNil)
			})
		})

		Convey("When an instance is created with a value greater than the max value", func() {
			pi := &pkg{Field: 6}
			Convey("Then validation must fail", func() {
				So(validateTemplate(pi), shouldBeError, `[package:pkg][field:Field] value "6" greater than the specified maximum "5"`)
			})
		})
	})
}

func TestIntValidationSize(t *testing.T) {
	Convey("Given a package with a int field that has a size value", t, func() {
		type pkg struct {
			genericPkg
			Field int `urknall:"size=5"`
		}

		pi := &pkg{}
		Convey("Then validation must return an error", func() {
			So(validateTemplate(pi), shouldBeError, `[package:pkg][field:Field] type "int" doesn't support "size" tag`)
		})
	})
}

func TestStringValidationRequired(t *testing.T) {
	Convey("Given a package with a string field that has an invalid required annotation", t, func() {
		type pkg struct {
			genericPkg
			Field string `urknall:"required=tru"`
		}
		Convey("When an instance is created without specifying a value", func() {
			pi := &pkg{}
			Convey("Then validation must return an error", func() {
				So(validateTemplate(pi), shouldBeError, `[package:pkg][field:Field] failed to parse value (neither "true" nor "false") of tag "required": "tru"`)
			})
		})
	})

	Convey("Given a package with a string field that is required", t, func() {
		type pkg struct {
			genericPkg
			Field string `urknall:"required=true"`
		}
		Convey("When an instance is created without specifying a value", func() {
			pi := &pkg{}
			Convey("Then validation must return an error", func() {
				So(validateTemplate(pi), shouldBeError, "[package:pkg][field:Field] required field not set")
			})
		})

		Convey("When an instance is created with an empty value specified", func() {
			pi := &pkg{Field: ""}
			Convey("Then validation must return an error", func() {
				So(validateTemplate(pi), shouldBeError, "[package:pkg][field:Field] required field not set")
			})
		})

		Convey("When an instance is created with a value specified", func() {
			pi := &pkg{Field: "something"}
			Convey("Then validation must succeed", func() {
				So(validateTemplate(pi), ShouldBeNil)
			})
		})
	})
}

func TestStringValidationDefault(t *testing.T) {
	Convey("Given a package with a string field that has a default value", t, func() {
		type pkg struct {
			genericPkg
			Field string `urknall:"default='the 'default' value'"`
		}
		Convey("When an instance is created without specifying a value", func() {
			pi := &pkg{}
			Convey("Then validation must set the default value", func() {
				validateTemplate(pi)
				So(pi.Field, ShouldEqual, "the 'default' value")
			})
		})

		Convey("When an instance is created with an empty value specified", func() {
			pi := &pkg{Field: ""}
			Convey("Then validation must set the default value", func() {
				validateTemplate(pi)
				So(pi.Field, ShouldEqual, "the 'default' value")
			})
		})

		Convey("When an instance is created with a value specified", func() {
			pi := &pkg{Field: "some other value"}
			Convey("Then validation must not touch the set value", func() {
				validateTemplate(pi)
				So(pi.Field, ShouldEqual, "some other value")
			})
		})
	})
}

func TestStringValidationMinMax(t *testing.T) {
	Convey("Given a package with a string field that has minimum and maximum length specified", t, func() {
		type pkg struct {
			genericPkg
			Field string `urknall:"min=3 max=4"`
		}
		Convey("When an instance is created without specifying a value", func() {
			pi := &pkg{}
			Convey("Then validation must succeed", func() {
				So(validateTemplate(pi), ShouldBeNil)
			})
		})

		Convey("When an instance is created with a value smaller than the minimum length", func() {
			pi := &pkg{Field: "ab"}
			Convey("Then validation must fail with an according error", func() {
				So(validateTemplate(pi), shouldBeError, `[package:pkg][field:Field] length of value "ab" smaller than the specified minimum length "3"`)
			})
		})

		Convey("When an instance is created with a value equal to the minimum length", func() {
			pi := &pkg{Field: "abc"}
			Convey("Then validation must succeed", func() {
				So(validateTemplate(pi), ShouldBeNil)
			})
		})

		Convey("When an instance is created with a value equal to the maximum length", func() {
			pi := &pkg{Field: "abcd"}
			Convey("Then validation must succeed", func() {
				So(validateTemplate(pi), ShouldBeNil)
			})
		})

		Convey("When an instance is created with a value longer than the maximum length", func() {
			pi := &pkg{Field: "abcde"}
			Convey("Then validation must fail with an according error", func() {
				So(validateTemplate(pi), shouldBeError, `[package:pkg][field:Field] length of value "abcde" greater than the specified maximum length "4"`)
			})
		})
	})
}

func TestStringValidationSize(t *testing.T) {
	Convey("Given a package with a string field that has a size set", t, func() {
		type pkg struct {
			genericPkg
			Field string `urknall:"size=3"`
		}
		Convey("When an instance is created without specifying a value", func() {
			pi := &pkg{}
			Convey("Then validation must succeed", func() {
				So(validateTemplate(pi), ShouldBeNil)
			})
		})

		Convey("When an instance is created with a value smaller than the size", func() {
			pi := &pkg{Field: "ab"}
			Convey("Then validation must fail with an according error", func() {
				So(validateTemplate(pi), shouldBeError, `[package:pkg][field:Field] length of value "ab" doesn't match the specified size "3"`)
			})
		})

		Convey("When an instance is created with a value equal to the size", func() {
			pi := &pkg{Field: "abc"}
			Convey("Then validation must succeed", func() {
				So(validateTemplate(pi), ShouldBeNil)
			})
		})

		Convey("When an instance is created with a value greater than the size", func() {
			pi := &pkg{Field: "abcd"}
			Convey("Then validation must fail with an according error", func() {
				So(validateTemplate(pi), shouldBeError, `[package:pkg][field:Field] length of value "abcd" doesn't match the specified size "3"`)
			})
		})
	})
}

func TestValidationRequiredInvalid(t *testing.T) {
	Convey("Given a package with a string field that has the required flag set to a wrong value", t, func() {
		type pkg struct {
			genericPkg
			Field string `urknall:"required=aberja"`
		}
		Convey("When an instance is created", func() {
			pi := &pkg{}
			Convey("Then validation must fail with an according error", func() {
				So(validateTemplate(pi), shouldBeError, `[package:pkg][field:Field] failed to parse value (neither "true" nor "false") of tag "required": "aberja"`)
			})
		})
	})
}
func TestValidationMinInvalid(t *testing.T) {
	Convey("Given a package with a string field that has minimum length specified with an invalid value", t, func() {
		type pkg struct {
			genericPkg
			Field string `urknall:"min=..3"`
		}
		Convey("When an instance is created", func() {
			pi := &pkg{}
			Convey("Then validation must fail with an according error", func() {
				So(validateTemplate(pi), shouldBeError, `[package:pkg][field:Field] failed to parse value (not an int) of tag "min": "..3"`)
			})
		})
	})
}

func TestValidationMaxInvalid(t *testing.T) {
	Convey("Given a package with a string field that has maximum length specified with an invalid value", t, func() {
		type pkg struct {
			genericPkg
			Field string `urknall:"max=4a"`
		}
		Convey("When an instance is created", func() {
			pi := &pkg{}
			Convey("Then validation must fail with an according error", func() {
				So(validateTemplate(pi), shouldBeError, `[package:pkg][field:Field] failed to parse value (not an int) of tag "max": "4a"`)
			})
		})
	})
}

func TestValidationSizeInvalid(t *testing.T) {
	Convey("Given a package with a string field that has size specified with an invalid value", t, func() {
		type pkg struct {
			genericPkg
			Field string `urknall:"size=4a"`
		}
		Convey("When an instance is created", func() {
			pi := &pkg{}
			Convey("Then validation must fail with an according error", func() {
				So(validateTemplate(pi), shouldBeError, `[package:pkg][field:Field] failed to parse value (not an int) of tag "size": "4a"`)
			})
		})
	})
}

func TestMultiTags(t *testing.T) {
	Convey("Given a package with multiple tags set on a field", t, func() {
		type pkg struct {
			genericPkg
			Field string `urknall:"default='foo' min=3 max=4"`
		}

		Convey("When an instance is create without a value set", func() {
			pi := &pkg{}
			e := validateTemplate(pi)
			Convey("Then validation must succeed", func() {
				So(e, ShouldBeNil)
			})
			Convey("Then the instances value must be set properly", func() {
				So(pi.Field, ShouldEqual, "foo")
			})
		})

		Convey("When an instance is create without a erroneous valiue set", func() {
			pi := &pkg{Field: "ab"}
			Convey("Then validation must fail", func() {
				So(validateTemplate(pi), shouldBeError, `[package:pkg][field:Field] length of value "ab" smaller than the specified minimum length "3"`)
			})
		})
	})
}

func TestTagParsing(t *testing.T) {
	Convey("Given a package with a missing single quote", t, func() {
		type pkg struct {
			genericPkg
			Field string `urknall:"required='abc"`
		}

		Convey("Then parsing should fail", func() {
			pi := &pkg{Field: "asd"}
			Convey("Then validation must fail", func() {
				So(validateTemplate(pi), shouldBeError, `[package:pkg][field:Field] failed to parse tag due to erroneous quotes`)
			})
		})
	})

	Convey("Given a package with a to many single quotes", t, func() {
		type pkg struct {
			genericPkg
			Field string `urknall:"required='ab'c'"`
		}

		Convey("Then parsing should fail", func() {
			pi := &pkg{Field: "asd"}
			Convey("Then validation must fail", func() {
				So(validateTemplate(pi), shouldBeError, `[package:pkg][field:Field] failed to parse tag due to erroneous quotes`)
			})
		})
	})

	Convey("Given a package with a key without value", t, func() {
		type pkg struct {
			genericPkg
			Field string `urknall:"default"`
		}

		Convey("Then parsing should fail", func() {
			pi := &pkg{Field: "asd"}
			Convey("Then validation must fail", func() {
				So(validateTemplate(pi), shouldBeError, `[package:pkg][field:Field] failed to parse annotation (value missing): "default"`)
			})
		})
	})

	Convey("Given a package with an invalid key", t, func() {
		type pkg struct {
			genericPkg
			Field string `urknall:"defaul='asdf'"`
		}

		Convey("Then parsing should fail", func() {
			pi := &pkg{Field: "asd"}
			Convey("Then validation must fail", func() {
				So(validateTemplate(pi), shouldBeError, `[package:pkg][field:Field] tag "defaul" unknown`)
			})
		})
	})
}
