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

package models

type ProjectType string

// List of ProjectType
const (
	DEFAULT__ProjectType ProjectType = "DEFAULT"
	PREVIEW_ProjectType  ProjectType = "PREVIEW"
)

// Check if a given ProjectType is valid
func (e ProjectType) IsValid() bool {
	switch e {
	case DEFAULT__ProjectType,
		PREVIEW_ProjectType:
		return true
	}
	return false
}
