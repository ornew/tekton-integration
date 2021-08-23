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

import "fmt"

type ErrorCode string

const (
	ErrorCodeInvalidProviderSpec = ErrorCode("InvalidProviderSpec")
	ErrorCodeNotFoundPrivateKey  = ErrorCode("NotFoundPrivateKey")
	ErrorCodeFailedValidation    = ErrorCode("FailedValidation")
	ErrorCodeRuntimeError        = ErrorCode("RuntimeError")
)

type ProviderError struct {
	Code    ErrorCode
	Message string
}

var _ error = (*ProviderError)(nil)

func (e *ProviderError) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func NewInvalidProviderSpecError(msg string) *ProviderError {
	return &ProviderError{
		Code:    ErrorCodeInvalidProviderSpec,
		Message: msg,
	}
}

func NewNotFoundPrivateKeyError(msg string) *ProviderError {
	return &ProviderError{
		Code:    ErrorCodeNotFoundPrivateKey,
		Message: msg,
	}
}

func NewFailedValidationError(msg string) *ProviderError {
	return &ProviderError{
		Code:    ErrorCodeFailedValidation,
		Message: msg,
	}
}

func NewRuntimeError(msg string) *ProviderError {
	return &ProviderError{
		Code:    ErrorCodeRuntimeError,
		Message: msg,
	}
}
