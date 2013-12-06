package zwo

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// A 'Compiler' is an entity (lets call it a package) that adds commands to a given runlist, taking into account their
// own configuration.
type Compiler interface {
	Compile(rl *Runlist) // Add the package specific commands to the runlist.
}

// The 'CompileNamer' interface is used to specify an explicit name for an package (if interface is not implemented on
// the package then the package's struct name will be used).
type CompileNamer interface {
	CompileName() string
}

func validatePackage(pkg interface{}) error {
	v := reflect.ValueOf(pkg)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
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
		return fmt.Errorf(`[field:%s] value "%d" smaller than specified minimum "%d"`, field.Name, value, opts.min)
	}

	if opts.max != 0 && value > opts.max {
		return fmt.Errorf(`[field:%s] value "%d" greater than specified maximum "%d"`, field.Name, value, opts.max)
	}

	return nil
}

func validateString(field reflect.StructField, value string, opts *validationOptions) (e error) {
	if opts.min != 0 && value != "" && (int64(len(value))) < opts.min {
		return fmt.Errorf(`[field:%s] length of value %q smaller than specified minimum length "%d"`, field.Name, value, opts.min)
	}

	if opts.max != 0 && int64(len(value)) > opts.max {
		return fmt.Errorf(`[field:%s] length of value %q greater than specified maximum length "%d"`, field.Name, value, opts.max)
	}

	if opts.size != 0 && value != "" && int64(len(value)) != opts.size {
		return fmt.Errorf(`[field:%s] length of value %q doesn't match specified size "%d"`, field.Name, value, opts.size)
	}

	return nil
}

type validationOptions struct {
	required     bool
	defaultValue interface{}
	size         int64
	min          int64
	max          int64
}

func parseFieldValidationString(field reflect.StructField) (opts *validationOptions, e error) {
	opts = &validationOptions{}
	tagString := field.Tag.Get("zwo")

	fields := []string{}
	idxStart := 0
	sA := false
	for i, c := range tagString {
		if c == '\'' {
			sA = !sA
		}

		if (c == ' ' || i+1 == len(tagString)) && !sA {
			fields = append(fields, tagString[idxStart:i+1])
			idxStart = i + 1
		}
	}
	if sA {
		return nil, fmt.Errorf("failed to parse tag due to erroneous quote levels")
	}

	for fIdx := range fields {
		kvList := strings.SplitN(fields[fIdx], "=", 2)
		if len(kvList) != 2 {
			return nil, fmt.Errorf("failed to parse key value pair %q", fields[fIdx])
		}
		key := strings.TrimSpace(kvList[0])
		value := strings.Trim(kvList[1], " '")

		switch key {
		case "required":
			switch field.Type.String() {
			case "string":
				if value != "true" && value != "false" {
					return nil, fmt.Errorf("wrong value for 'required' validation field: %q", value)
				}
				opts.required = value == "true"
			default:
				return nil, fmt.Errorf(`type %q doesn't support "required" tag`, field.Type.String())
			}
		case "default":
			switch field.Type.String() {
			case "string":
				opts.defaultValue = value
			case "int":
				i, e := strconv.ParseInt(value, 10, 64)
				if e != nil {
					return nil, fmt.Errorf("failed to parse value of field 'default': %s", e.Error())
				}
				opts.defaultValue = i
			case "bool":
				if value != "true" && value != "false" {
					return nil, fmt.Errorf("wrong value for field 'default': only 'true' and 'false' are allowed (is %q)", value)
				}
				opts.defaultValue = value == "true"
			default:
				return nil, fmt.Errorf("type %q not supported yet", field.Type.String())
			}
		case "size":
			switch field.Type.String() {
			case "string":
				i, e := strconv.ParseInt(value, 10, 64)
				if e != nil {
					return nil, fmt.Errorf("failed to parse value of field 'size': %s", e.Error())
				}
				opts.size = i
			default:
				return nil, fmt.Errorf(`type %q doesn't support "size" tag`, field.Type.String())
			}
		case "min":
			switch field.Type.String() {
			case "string":
				fallthrough
			case "int":
				i, e := strconv.ParseInt(value, 10, 64)
				if e != nil {
					return nil, fmt.Errorf("failed to parse value of field 'min': %s", e.Error())
				}
				opts.min = i
			default:
				return nil, fmt.Errorf(`type %q doesn't support "min" tag`, field.Type.String())
			}
		case "max":
			switch field.Type.String() {
			case "string":
				fallthrough
			case "int":
				i, e := strconv.ParseInt(value, 10, 64)
				if e != nil {
					return nil, fmt.Errorf("failed to parse value of field 'max': %s", e.Error())
				}
				opts.max = i
			default:
				return nil, fmt.Errorf(`type %q doesn't support "max" tag`, field.Type.String())
			}
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
