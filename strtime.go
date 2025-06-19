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

/* Strftime writes bytes representing t in the specified format to buf and returns the number of non-null bytes written (n) and an error. On success, this function also writes a null byte to buf[n]. On failure, the contents of buf are undefined. Because some valid format specifiers can result in output strings of 0 bytes, a return value of 0 does not necessarily indicate an error. C's strftime function provides no means for disambiguating such cases from failure cases. To simplify error handling, if nonzero is true, Strftime will return a non-nil error whenever C's strfime returns or would return a value of 0. If nonzero is false, Strftime will only return a non-nil error if t is invalid. This function is not concurrency safe: buf must not be altered by another goroutine for the duration of the call. */
func Strftime(t time.Time, buf []byte, format string, nonzero bool) (int, error) {
	/* this is the only instance in which the potential cause of a zero
	   result from fmt_time can be trivially determined */
	if len(buf) == 0 {
		if nonzero {
			return 0, ErrBufLen
		}
		return 0, nil
	}
	fbytes := []byte(format)
	fbuf := make([]byte, len(fbytes)+1)
	copy(fbuf, fbytes)

	unix := t.Unix()
	n := C.fmt_time(C.int64_t(unix), ptr(buf), C.int64_t(len(buf)), ptr(fbuf))
	if n < 0 {
		return 0, lut[-n]
	}
	if nonzero && n == 0 {
		return 0, ErrFormat
	}
	return int(n), nil
}
