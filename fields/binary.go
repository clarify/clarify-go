// Copyright 2022 Searis AS
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package fields

import (
	"bytes"
	"encoding"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
)

// HexadecimalNullZero is a variant of Hexadecimal which zero value JSON-encodes
// to null.
type HexadecimalNullZero Hexadecimal

var _ json.Marshaler = HexadecimalNullZero{}
var _ json.Unmarshaler = (*HexadecimalNullZero)(nil)

func (zn HexadecimalNullZero) MarshalJSON() ([]byte, error) {
	if zn == nil {
		return []byte(`null`), nil
	}
	return json.Marshal(Hexadecimal(zn))
}

func (zn *HexadecimalNullZero) UnmarshalJSON(data []byte) error {
	data = bytes.TrimSpace(data)
	if bytes.Equal(data, []byte(`null`)) {
		*zn = nil
	}
	return json.Unmarshal(data, (*Hexadecimal)(zn))
}

// Hexadecimal wraps a slice of bytes in order for it to be represented as a
// hexadecimal string when encoded. It is less compact that Base64 encoding.
type Hexadecimal []byte

var _ encoding.TextMarshaler = Hexadecimal(nil)
var _ encoding.TextUnmarshaler = (*Hexadecimal)(nil)

func (b Hexadecimal) String() string {
	return hex.EncodeToString(b)
}

func (b Hexadecimal) MarshalText() ([]byte, error) {
	buf := make([]byte, hex.EncodedLen(len(b)))
	hex.Encode(buf, b)
	return buf, nil
}

func (b *Hexadecimal) UnmarshalText(data []byte) error {
	buf := make([]byte, hex.DecodedLen(len(data)))
	n, err := hex.Decode(buf, data)
	if err != nil {
		return err
	}
	*b = Hexadecimal(buf[:n])
	return nil
}

// Base64NullZero is a variant of Base64 which zero value JSON-encodes to null.
type Base64NullZero Base64

var _ json.Marshaler = Base64NullZero{}
var _ json.Unmarshaler = (*Base64NullZero)(nil)

func (zn Base64NullZero) MarshalJSON() ([]byte, error) {
	if zn == nil {
		return []byte(`null`), nil
	}
	return json.Marshal(Base64(zn))
}

func (zn *Base64NullZero) UnmarshalJSON(data []byte) error {
	data = bytes.TrimSpace(data)
	if bytes.Equal(data, []byte(`null`)) {
		*zn = nil
	}
	return json.Unmarshal(data, (*Base64)(zn))
}

// Base64 wraps a slice of bytes in order for it to be represented as an RF 4648
// raw URL encoded string (i.e without padding). It's more compact than
// Hexadecimal.
type Base64 []byte

var _ encoding.TextMarshaler = Hexadecimal(nil)
var _ encoding.TextUnmarshaler = (*Hexadecimal)(nil)
var b64enc = base64.RawURLEncoding

func (b Base64) String() string {
	return b64enc.EncodeToString(b)
}

func (b Base64) MarshalText() ([]byte, error) {
	buf := make([]byte, b64enc.EncodedLen(len(b)))
	b64enc.Encode(buf, b)
	return buf, nil
}

func (b *Base64) UnmarshalText(data []byte) error {
	buf := make([]byte, b64enc.DecodedLen(len(data)))
	n, err := b64enc.Decode(buf, data)
	if err != nil {
		return err
	}
	*b = Base64(buf[:n])
	return nil
}
