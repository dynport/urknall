package zwo

import (
	"fmt"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func shouldBeError(actual interface{}, expected ...interface{}) string {
	if actual == nil {
		return "expected error, got nil"
	}

	err, actIsError := actual.(error)
	msg, msgIsString := expected[0].(string)

	if !actIsError {
		return fmt.Sprintf("expected something of type error\nexpected: %s\nactual: %s", msg, actual)
	}

	if !msgIsString {
		return fmt.Sprintf("expected value must be of type string: %s", expected[0])
	}

	if err.Error() != msg {
		return fmt.Sprintf("error message did not match\nexpected: %s\nactual: %s", msg, actual)
	}

	return ""
}

func TestBoolValidation(t *testing.T) {
	Convey("Given a package with a bool field that is required", t, func() {
		type pkg struct {
			Field bool `zwo:"required:true"`
		}
		Convey("When an instance is created without specifying a value", func () {
			instance := &pkg{}
			Convey("Then validation must return an error", func() {
				So(validatePackage(instance), shouldBeError, "[package:pkg][field:Field] required field not set")
			})
		})
		Convey("When an instance is created with value set to false", func () {
			instance := &pkg{Field: false}
			Convey("Then validation must succeed", func() {
				So(validatePackage(instance), ShouldBeNil)
			})
		})
		Convey("When an instance is created with value set to true", func () {
			instance := &pkg{Field: true}
			Convey("Then validation must succeed", func() {
				So(validatePackage(instance), ShouldBeNil)
			})
		})
	})
}

func TestStringValidation(t *testing.T) {
	Convey("Given a package with a string field that is required", t, func() {
		type pkg struct {
			Field string `zwo:"required:true"`
		}
		Convey("When an instance is created without specifying a value", func() {
			instance := &pkg{}
			Convey("Then validation must return an error", func() {
				So(validatePackage(instance), shouldBeError, "[package:pkg][field:Field] required field not set")
			})
		})

		Convey("When an instance is created with an empty value specified", func() {
			instance := &pkg{Field: ""}
			Convey("Then validation must return an error", func() {
				So(validatePackage(instance), shouldBeError, "[package:pkg][field:Field] required field not set")
			})
		})

		Convey("When an instance is created with a value specified", func() {
			instance := &pkg{Field: "something"}
			Convey("Then validation must succeed", func() {
				So(validatePackage(instance), ShouldBeNil)
			})
		})
	})

	Convey("Given a package with a string field that has a default value", t, func() {
		type pkg struct {
			Field string `zwo:"default:'the value'"`
		}
		Convey("When an instance is created without specifying a value", func() {
			instance := &pkg{}
			Convey("Then validation must set the value", func() {
				validatePackage(instance)
				So(instance.Field, ShouldEqual, "the value")
			})
		})

		Convey("When an instance is created with an empty value specified", func() {
			instance := &pkg{Field: ""}
			Convey("Then validation must set the value", func() {
				validatePackage(instance)
				So(instance.Field, ShouldEqual, "the value")
			})
		})

		Convey("When an instance is created with a value specified", func() {
			instance := &pkg{Field: "some other value"}
			Convey("Then validation must not touch the set value", func() {
				validatePackage(instance)
				So(instance.Field, ShouldEqual, "some other value")
			})
		})
	})

	Convey("Given a package with a string field that has minimum and maximum length specified", t, func() {
		type pkg struct {
			Field string `zwo:"min:3, max:4"`
		}
		Convey("When an instance is created without specifying a value", func() {
			instance := &pkg{}
			Convey("Then validation must succeed", func() {
				So(validatePackage(instance), ShouldBeNil)
			})
		})

		Convey("When an instance is created with a value smaller than the minimum length", func() {
			instance := &pkg{Field: "ab"}
			Convey("Then validation must fail with an according error", func() {
				So(validatePackage(instance), shouldBeError, "[package:pkg][field:Field] length of value 'ab' smaller than specified minimum length '3'")
			})
		})

		Convey("When an instance is created with a value equal to the minimum length", func() {
			instance := &pkg{Field: "abc"}
			Convey("Then validation must succeed", func() {
				So(validatePackage(instance), ShouldBeNil)
			})
		})

		Convey("When an instance is created with a value equal to the maximum length", func() {
			instance := &pkg{Field: "abcd"}
			Convey("Then validation must succeed", func() {
				So(validatePackage(instance), ShouldBeNil)
			})
		})

		Convey("When an instance is created with a value longer than the maximum length", func() {
			instance := &pkg{Field: "abcde"}
			Convey("Then validation must fail with an according error", func() {
				So(validatePackage(instance), shouldBeError, "[package:pkg][field:Field] length of value 'abcde' greater than specified maximum length '4'")
			})
		})
	})

	Convey("Given a package with a string field that has a size set", t, func() {
		type pkg struct {
			Field string `zwo:"size:3"`
		}
		Convey("When an instance is created without specifying a value", func() {
			instance := &pkg{}
			Convey("Then validation must succeed", func() {
				So(validatePackage(instance), ShouldBeNil)
			})
		})

		Convey("When an instance is created with a value smaller than the size", func() {
			instance := &pkg{Field: "ab"}
			Convey("Then validation must fail with an according error", func() {
				So(validatePackage(instance), shouldBeError, "[package:pkg][field:Field] length of value 'ab' doesn't match specified size '3'")
			})
		})

		Convey("When an instance is created with a value equal to the size", func() {
			instance := &pkg{Field: "abc"}
			Convey("Then validation must succeed", func() {
				So(validatePackage(instance), ShouldBeNil)
			})
		})

		Convey("When an instance is created with a value larger than the size", func() {
			instance := &pkg{Field: "abcd"}
			Convey("Then validation must fail with an according error", func() {
				So(validatePackage(instance), shouldBeError, "[package:pkg][field:Field] length of value 'abcd' doesn't match specified size '3'")
			})
		})
	})

	Convey("Given a package with a string field that has minimum length specified with an invalid value", t, func() {
		type pkg struct {
			Field string `zwo:"min:..3"`
		}
		Convey("When an instance is created", func() {
			instance := &pkg{}
			Convey("Then validation must fail with an according error", func() {
				So(validatePackage(instance), shouldBeError, "[package:pkg][field:Field] failed to parse value of field 'min': strconv.ParseInt: parsing \"..3\": invalid syntax")
			})
		})
	})

	Convey("Given a package with a string field that has maximum length specified with an invalid value", t, func() {
		type pkg struct {
			Field string `zwo:"max:4a"`
		}
		Convey("When an instance is created", func() {
			instance := &pkg{}
			Convey("Then validation must fail with an according error", func() {
				So(validatePackage(instance), shouldBeError, "[package:pkg][field:Field] failed to parse value of field 'max': strconv.ParseInt: parsing \"4a\": invalid syntax")
			})
		})
	})
}
