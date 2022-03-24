package resource

import (
	"encoding"
	"encoding/base64"
	"encoding/json"

	"golang.org/x/exp/slices"
)

// Binary wraps a slice of bytes in order for it to be represented as a base64
// string in text-based encoding formats.
type Binary []byte

var _ encoding.TextMarshaler = Binary(nil)
var _ encoding.TextUnmarshaler = (*Binary)(nil)

func (b Binary) String() string {
	return base64.StdEncoding.EncodeToString(b)
}

func (b Binary) MarshalText() ([]byte, error) {
	enc := base64.StdEncoding
	buf := make([]byte, enc.EncodedLen(len(b)))
	enc.Encode(buf, b)
	return buf, nil
}

func (b *Binary) UnmarshalText(data []byte) error {
	enc := base64.StdEncoding
	buf := make([]byte, enc.DecodedLen(len(data)))
	n, err := enc.Decode(buf, data)
	if err != nil {
		return err
	}

	*b = make(Binary, n)
	copy(*b, buf[:n])
	return nil
}

type Labels map[string][]string

func (l Labels) MarshalJSON() ([]byte, error) {
	if len(l) == 0 {
		return []byte(`{}`), nil
	}
	return json.Marshal(map[string][]string(l))
}

// Set relplace the set of values at the given location. Any provided duplicates
// are automatically removed. If there is no values, the key is deleted.
func (l *Labels) Set(key string, values []string) {
	switch {
	case len(values) == 0 && (*l) == nil:
	case len(values) == 0:
		delete(*l, key)
	case (*l) == nil:
		(*l) = make(Labels)
		fallthrough
	default:
		ll := slices.Clone(values)
		slices.Sort(ll)
		(*l)[key] = slices.Compact(ll)
	}
}

// Add adds the specified value to the relevant key if it's not already present.
// The resulting values is a sorted array.
func (l *Labels) Add(key string, value string) {
	if *l == nil {
		*l = make(Labels)
	}
	ll := (*l)[key]
	if len(ll) == 0 {
		(*l)[key] = []string{value}
		return
	}

	slices.Sort(ll)
	if i, found := slices.BinarySearch(ll, value); !found {
		slices.Insert(ll, i, value)
	}

	(*l)[key] = ll
}

// Remove remvoes the specified value from the relevant key. If there are no
// values left for the key, the key is deleted.
func (l *Labels) Remove(key string, value string) {
	if *l == nil {
		return
	}
	ll := (*l)[key]
	if len(ll) == 0 {
		delete((*l), key)
		return
	}
	slices.Sort((*l)[key])
	if i, found := slices.BinarySearch(ll, value); !found {
		slices.Delete(ll, i, i+1)
	}
	if len(ll) == 0 {
		delete((*l), key)
		return
	}
	(*l)[key] = ll
}

// EnumValues maps integer Items values to strings.
type EnumValues map[int]string

func (e EnumValues) MarshalJSON() ([]byte, error) {
	if len(e) == 0 {
		return []byte(`{}`), nil
	}
	return json.Marshal(map[int]string(e))
}
