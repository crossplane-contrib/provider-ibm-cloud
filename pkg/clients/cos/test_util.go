/*
Copyright 2021 The Crossplane Authors.

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

package cos

import (
	"time"

	"github.com/go-openapi/strfmt"
)

// ADateTimeInAYear returns a  (random, but fixed) date time in the given year
func ADateTimeInAYear(year int) *strfmt.DateTime {
	result := strfmt.DateTime(time.Date(year, 10, 12, 8, 5, 5, 0, time.UTC))

	return &result
}

// AStrArray returns an array of strings
func AStrArray() []string {
	result := []string{"a", "b", "cd"}

	return result
}
