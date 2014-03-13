package urknall

import (
	"fmt"
	"reflect"
	"strconv"

	"github.com/dynport/dgtk/tagparse"
)

type Role interface {
	Apply(host *Host)
}

// A "Package" is an entity that packs commands into a runlist, taking into account their own configuration.
type Package interface {
	Package(rl *Runlist) // Add the package specific commands to the runlist.
}

// Compile the given package and return the generated runlist.
func CompilePackage(pkg Package) (*Runlist, error) {
	rl := &Runlist{pkg: pkg}
	return rl, rl.compile()
}

type anonymousPackage struct {
	cmds []interface{}
}

func (anon *anonymousPackage) Package(rl *Runlist) {
	for i := range anon.cmds {
		rl.Add(anon.cmds[i])
	}
}

// Create a package from a set of commands.
func NewPackage(cmds ...interface{}) Package {
	return &anonymousPackage{cmds: cmds}
}

func validatePackage(pkg interface{}) error {
	v := reflect.ValueOf(pkg)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return fmt.Errorf(`value is not a package, but of type "%T"`, pkg)
	}

	for i := 0; i < v.NumField(); i++ {
		e := validateField(v.Type().Field(i), v.Field(i))
		if e != nil {
			return fmt.Errorf("[package:%s]%s", v.Type().Name(), e.Error())
		}
	}
	return nil
}

func validateInt(field reflect.StructField, value int64, opts *validationOptions) (e error) {
	if opts.min != 0 && value < opts.min {
		return fmt.Errorf(`[field:%s] value "%d" smaller than the specified minimum "%d"`, field.Name, value, opts.min)
	}

	if opts.max != 0 && value > opts.max {
		return fmt.Errorf(`[field:%s] value "%d" greater than the specified maximum "%d"`, field.Name, value, opts.max)
	}

	return nil
}

func validateString(field reflect.StructField, value string, opts *validationOptions) (e error) {
	if opts.min != 0 && value != "" && (int64(len(value))) < opts.min {
		return fmt.Errorf(`[field:%s] length of value %q smaller than the specified minimum length "%d"`, field.Name, value, opts.min)
	}

	if opts.max != 0 && int64(len(value)) > opts.max {
		return fmt.Errorf(`[field:%s] length of value %q greater than the specified maximum length "%d"`, field.Name, value, opts.max)
	}

	if opts.size != 0 && value != "" && int64(len(value)) != opts.size {
		return fmt.Errorf(`[field:%s] length of value %q doesn't match the specified size "%d"`, field.Name, value, opts.size)
	}

	return nil
}

const (
	parse_INT_ERROR   = `failed to parse value (not an int) of tag %q: "%s"`
	parse_BOOL_ERROR  = `failed to parse value (neither "true" nor "false") of tag %q: "%s"`
	unknown_TAG_ERROR = `type %q doesn't support %q tag`
)

type validationOptions struct {
	required     bool
	defaultValue interface{}
	size         int64
	min          int64
	max          int64
}

func splitter(input string) (key, value string, e error) {
	switch input {
	case "required":
		return "required", "true", nil
	default:
		return "", "", fmt.Errorf("failed")
	}
}

func parseFieldValidationString(field reflect.StructField) (opts *validationOptions, e error) {
	opts = &validationOptions{}

	tagMap, e := tagparse.ParseCustom(field, "urknall", splitter)
	if e != nil {
		return nil, e
	}

	for key, value := range tagMap {
		switch key {
		case "required":
			switch field.Type.String() {
			case "string", "[]uint8":
				if value != "true" && value != "false" {
					return nil, fmt.Errorf(parse_BOOL_ERROR, key, value)
				}
				opts.required = value == "true"
			default:
				return nil, fmt.Errorf(unknown_TAG_ERROR, field.Type.String(), key)
			}
		case "default":
			switch field.Type.String() {
			case "string", "[]uint8":
				opts.defaultValue = value
			case "int":
				i, e := strconv.ParseInt(value, 10, 64)
				if e != nil {
					return nil, fmt.Errorf(parse_INT_ERROR, key, value)
				}
				opts.defaultValue = i
			case "bool":
				if value != "true" && value != "false" {
					return nil, fmt.Errorf(parse_BOOL_ERROR, key, value)
				}
				opts.defaultValue = value == "true"
			default:
				return nil, fmt.Errorf(unknown_TAG_ERROR, field.Type.String(), key)
			}
		case "size":
			switch field.Type.String() {
			case "string":
				i, e := strconv.ParseInt(value, 10, 64)
				if e != nil {
					return nil, fmt.Errorf(parse_INT_ERROR, key, value)
				}
				opts.size = i
			default:
				return nil, fmt.Errorf(unknown_TAG_ERROR, field.Type.String(), key)
			}
		case "min":
			switch field.Type.String() {
			case "string":
				fallthrough
			case "int":
				i, e := strconv.ParseInt(value, 10, 64)
				if e != nil {
					return nil, fmt.Errorf(parse_INT_ERROR, key, value)
				}
				opts.min = i
			default:
				return nil, fmt.Errorf(unknown_TAG_ERROR, field.Type.String(), key)
			}
		case "max":
			switch field.Type.String() {
			case "string":
				fallthrough
			case "int":
				i, e := strconv.ParseInt(value, 10, 64)
				if e != nil {
					return nil, fmt.Errorf(parse_INT_ERROR, key, value)
				}
				opts.max = i
			default:
				return nil, fmt.Errorf(unknown_TAG_ERROR, field.Type.String(), key)
			}
		default:
			return nil, fmt.Errorf(`tag %q unknown`, key)
		}
	}
	return opts, nil
}

func validateField(field reflect.StructField, value reflect.Value) error {
	opts, e := parseFieldValidationString(field)
	if e != nil {
		return fmt.Errorf("[field:%s] %s", field.Name, e.Error())
	}

	switch field.Type.String() {
	case "[]uint8":
		if opts.required && len(value.Bytes()) == 0 {
			return fmt.Errorf("[field:%s] required field not set", field.Name)
		}
		if opts.defaultValue != nil && len(value.Bytes()) == 0 {
			value.SetBytes([]byte(opts.defaultValue.(string)))
		}
		return nil
	case "string":
		if opts.required && value.String() == "" {
			return fmt.Errorf("[field:%s] required field not set", field.Name)
		}

		if opts.defaultValue != nil && value.String() == "" {
			value.SetString(opts.defaultValue.(string))
		}
		return validateString(field, value.String(), opts)
	case "int":
		if opts.defaultValue != nil && value.Int() == 0 {
			value.SetInt(opts.defaultValue.(int64))
		}
		return validateInt(field, value.Int(), opts)
	case "bool":
		if opts.defaultValue != nil && opts.defaultValue.(bool) {
			value.SetBool(true)
		}
	}
	return nil
}
