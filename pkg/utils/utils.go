package utils

import (
	"bytes"
	"fmt"
	"github.com/google/triage-party/pkg/models"
	"io"
	"net/url"
	"reflect"
	"strings"
)

// parseRepo returns provider, organization and project for a URL
// rawURL should be a valid url with host like https://github.com/org/repo
func ParseRepo(rawURL string) (r models.Repo, err error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return
	}
	if u.Host == "" {
		err = fmt.Errorf("Provided string %s is not a valid URL", rawURL)
		return
	}
	parts := strings.Split(u.Path, "/")
	if len(parts) != 3 {
		err = fmt.Errorf("expected 2 repository parts, got %d: %v", len(parts), parts)
		return
	}
	r = models.Repo{
		Host:         u.Host,
		Organization: parts[1],
		Project:      parts[2],
	}
	return
}

var timestampType = reflect.TypeOf(models.Timestamp{})

// Stringify attempts to create a reasonable string representation of types in
// the GitHub library. It does things like resolve pointers to their values
// and omits struct fields with nil values.
func Stringify(message interface{}) string {
	var buf bytes.Buffer
	v := reflect.ValueOf(message)
	stringifyValue(&buf, v)
	return buf.String()
}

// stringifyValue was heavily inspired by the goprotobuf library.

func stringifyValue(w io.Writer, val reflect.Value) {
	if val.Kind() == reflect.Ptr && val.IsNil() {
		w.Write([]byte("<nil>"))
		return
	}

	v := reflect.Indirect(val)

	switch v.Kind() {
	case reflect.String:
		fmt.Fprintf(w, `"%s"`, v)
	case reflect.Slice:
		w.Write([]byte{'['})
		for i := 0; i < v.Len(); i++ {
			if i > 0 {
				w.Write([]byte{' '})
			}

			stringifyValue(w, v.Index(i))
		}

		w.Write([]byte{']'})
		return
	case reflect.Struct:
		if v.Type().Name() != "" {
			w.Write([]byte(v.Type().String()))
		}

		// special handling of Timestamp values
		if v.Type() == timestampType {
			fmt.Fprintf(w, "{%s}", v.Interface())
			return
		}

		w.Write([]byte{'{'})

		var sep bool
		for i := 0; i < v.NumField(); i++ {
			fv := v.Field(i)
			if fv.Kind() == reflect.Ptr && fv.IsNil() {
				continue
			}
			if fv.Kind() == reflect.Slice && fv.IsNil() {
				continue
			}

			if sep {
				w.Write([]byte(", "))
			} else {
				sep = true
			}

			w.Write([]byte(v.Type().Field(i).Name))
			w.Write([]byte{':'})
			stringifyValue(w, fv)
		}

		w.Write([]byte{'}'})
	default:
		if v.CanInterface() {
			fmt.Fprint(w, v.Interface())
		}
	}
}
