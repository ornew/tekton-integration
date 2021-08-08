/*
Copyright 2021 Arata Furukawa.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package providers

const redacted = "[REDACTED]"

var redactedJSONBytes = []byte("\"" + redacted + "\"")

type SecretBytes struct {
	data []byte
}

func NewSecretBytes(data []byte) SecretBytes {
	return SecretBytes{
		data: data,
	}
}

func (s *SecretBytes) String() string {
	return redacted
}

func (s *SecretBytes) GoString() string {
	return redacted
}

func (s *SecretBytes) MarshalJSON() ([]byte, error) {
	return redactedJSONBytes, nil
}

func (s *SecretBytes) GetNoRedacted() []byte {
	return s.data
}

func (s *SecretBytes) GetNoRedactedString() string {
	return string(s.data)
}

type SecretString struct {
	data string
}

func NewSecretString(data string) SecretString {
	return SecretString{
		data: data,
	}
}

func (s *SecretString) String() string {
	return redacted
}

func (s *SecretString) GoString() string {
	return redacted
}

func (s *SecretString) MarshalJSON() ([]byte, error) {
	return redactedJSONBytes, nil
}

func (s *SecretString) GetNoRedacted() []byte {
	return []byte(s.data)
}

func (s *SecretString) GetNoRedactedString() string {
	return s.data
}
