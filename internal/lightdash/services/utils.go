// Copyright 2023 Ubie, inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package services

import (
	"fmt"
	"regexp"
)

// ExtractStringsByPattern takes an input string and a regex pattern with capture groups,
// and returns the captured strings
func ExtractStringsByPattern(input, pattern string) ([]string, error) {
	regex, err := regexp.Compile(pattern)
	if err != nil {
		return nil, fmt.Errorf("invalid regex pattern: %w", err)
	}

	matches := regex.FindStringSubmatch(input)
	if len(matches) < 2 {
		return nil, fmt.Errorf("no matches found for pattern: %s", pattern)
	}

	// Return all capture groups (skip the first element which is the full match)
	return matches[1:], nil
}

// compareParentSpaceUUID compares two parent space UUIDs
func CompareParentSpaceUUID(parentSpaceUUID1, parentSpaceUUID2 *string) bool {
	if parentSpaceUUID1 == nil && parentSpaceUUID2 == nil {
		return true
	}
	if parentSpaceUUID1 == nil || parentSpaceUUID2 == nil {
		return false
	}
	return *parentSpaceUUID1 == *parentSpaceUUID2
}
