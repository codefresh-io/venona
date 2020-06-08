// Copyright 2020 The Codefresh Authors.
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

package codefresh

import "fmt"

type (
	// Error is an error that may be thrown from Codefresh API
	Error struct {
		Message       string
		APIStatusCode int
	}
)

func (c Error) Error() string {
	return fmt.Sprintf("HTTP request to Codefresh API rejected. Status-Code: %d. Message: %s", c.APIStatusCode, c.Message)
}
