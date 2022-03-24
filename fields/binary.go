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
	"encoding"
	"encoding/base64"
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
