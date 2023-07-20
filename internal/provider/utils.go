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

package provider

import (
	"fmt"
	"regexp"
)

func extractStrings(input, pattern string) ([]string, error) {
	// Compile the regular expression
	regex, err := regexp.Compile(pattern)
	if err != nil {
		return nil, err
	}

	// Find the matches in the input string
	matches := regex.FindStringSubmatch(input)
	if len(matches) == 0 {
		return nil, fmt.Errorf("no matches found for the given pattern")
	}

	// Extract the captured groups
	groups := matches[1:]

	return groups, nil
}
