/*
Package strtime provides Go wrappers for C's strftime and strptime functions. This package allows users to parse and format Go time.Time values using C's time formatting directives.
*/
package strtime

// #include "gotime.h"
// #cgo nocallback parse_time
// #cgo nocallback fmt_time
import "C"

import (
	"time"
	"unsafe"
)

/* returns a c char * pointing to b's backing byte array */
func ptr(b []byte) *C.char { return (*C.char)(unsafe.Pointer(&b[0])) }

type CError string

const (
	/* The C strftime function can write 0 bytes to buf in several cases, some of which do not indicate actual errors. It does not provide any means for resolving this ambiguity. */
	ErrFormat CError = "unable to format time; consider amending the layout string or increasing the buf size"
	ErrParse  CError = "unable to parse time"
	ErrValid  CError = "invalid time"
)

var lut = [...]CError{
	ErrFormat,
	ErrParse,
	ErrValid,
}

func (c CError) Error() string { return string(c) }

type ArgError string

const (
	ErrLayout ArgError = "invalid layout arg"
	ErrValue  ArgError = "invalid value arg"
	ErrBufLen ArgError = "buf len is too short to accommodate a result"
)

func (a ArgError) Error() string { return string(a) }

/* Strptime parses the time represented by value using the format specified by the layout parameter. It returns a time.Time object and an error. The minimum resolution of the value is 1 second. */
func Strptime(value, format string) (time.Time, error) {
	if format == "" {
		return time.Time{}, ErrLayout
	}
	if value == "" {
		return time.Time{}, ErrValue
	}
	fbytes := []byte(format)
	vbytes := []byte(value)

	/* ensure these values are null-terminated */
	fbuf := make([]byte, len(fbytes)+1)
	copy(fbuf, fbytes)
	vbuf := make([]byte, len(vbytes)+1)
	copy(vbuf, vbytes)

	res := C.parse_time(ptr(vbuf), ptr(fbuf))

	if res.status != 0 {
		return time.Time{}, lut[res.status]
	}

	return time.Unix(int64(res.value), 0), nil
}

/* Strftime writes bytes representing t in the specified format to buf and returns the number of non-null bytes written and an error. If nonzero is true, it returns a non-nil error when a write of 0 non-null bytes occurred or would occur. Note: on success, this function also writes a null byte to buf[n]. (The null byte is not included in the returned count.) If the len of the provided buf is less than the size needed to write the requested bytes (including the null byte), no bytes will be written. A result of 0 can be ambiguous: it can either indicate failure (no bytes were written) or success (the layout resulted in a string representation of len 0, but the terminating null byte was written). To disambiguate this, the nonzero parameter can be used to specify whether or not a 0 result is acceptable. If nonzero is true, Strftime will return a non-nil error whenever the result is 0. */
func Strftime(t time.Time, buf []byte, format string, nonzero bool) (int, error) {
	/* this is the only instance in which the potential cause of a zero
	   result from fmt_time can be trivially determined */
	if nonzero && len(buf) == 0 {
		return 0, ErrBufLen
	}
	fbytes := []byte(format)
	fbuf := make([]byte, len(fbytes)+1)
	copy(fbuf, fbytes)

	unix := t.Unix()
	n := C.fmt_time(C.int64_t(unix), ptr(fbuf), ptr(buf), C.int64_t(len(buf)))
	if n < 0 {
		return 0, lut[-n]
	}
	if nonzero && n == 0 {
		return 0, ErrFormat
	}
	return int(n), nil
}
