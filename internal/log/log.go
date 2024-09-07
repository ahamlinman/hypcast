// Package log provides special log formatting features for Hypcast.
package log

import (
	"log"
	"reflect"
)

// Tprintf prints its arguments in the manner of [log.Printf], with a prefix of
// the form "Type(0xabcd...)" indicating the element type and address of the src
// pointer. This can be a convenient way to differentiate instances of the same
// type, though addresses are somewhat opaque as identifiers, and revealing them
// might even be a security concern (albeit a far-fetched one).
func Tprintf[T any](src *T, fmt string, v ...any) {
	tfmt := "%s(%p): " + fmt
	tval := make([]any, len(v)+2)
	tval[0], tval[1] = reflect.TypeFor[T]().Name(), src
	copy(tval[2:], v)
	log.Printf(tfmt, tval...)
}
