package loader

import (
	"strconv"
)

// StringKey implements the Key interface for a string
type StringKey string

// String is an identity method. Used to implement String interface
func (k StringKey) String() string { return string(k) }

// Raw is an identity method. Used to implement Key Raw
func (k StringKey) Raw() interface{} { return k }

// NewKeysFromStrings converts a `[]strings` to a `Keys` ([]Key)
func NewKeysFromStrings(strings []string) Keys {
	list := make(Keys, len(strings))
	for i := range strings {
		list[i] = StringKey(strings[i])
	}
	return list
}

// Int32Key implements the Key interface for a string
type Int32Key int

// String is an identity method. Used to implement String interface
func (k Int32Key) String() string { return strconv.Itoa(int(k)) }

// Raw is an identity method. Used to implement Key Raw
func (k Int32Key) Raw() interface{} { return k }
