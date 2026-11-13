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
	"strings"
	"testing"
)

// resourceIDCodec pairs a resource ID formatter with its parser.
type resourceIDCodec struct {
	name    string
	format  func(a, b string) string
	extract func(string) ([]string, error)
}

var resourceIDCodecs = []resourceIDCodec{
	{"project role member", getProjectRoleMemberResourceId, extractProjectRoleMemberResourceId},
	{"project role group", getProjectRoleGroupResourceId, extractProjectRoleGroupResourceId},
	{"organization group", getGroupResourceId, extractGroupResourceId},
	{"project upstream", getProjectUpstreamResourceId, extractProjectUpstreamResourceId},
}

// TestResourceIDCodecs_roundTripValidSegments checks that every ID built from
// non-empty segments without slashes parses back to the same segments.
func TestResourceIDCodecs_roundTripValidSegments(t *testing.T) {
	segments := []string{"a", "x-y", "3f4a9aa8-2b58-4767-b646-d92a43fbfd9d", "UPPER", "a_b.c"}
	for _, codec := range resourceIDCodecs {
		for _, a := range segments {
			for _, b := range segments {
				id := codec.format(a, b)
				got, err := codec.extract(id)
				if err != nil {
					t.Errorf("%s: extract(%q): unexpected error: %v", codec.name, id, err)
					continue
				}
				if len(got) != 2 || got[0] != a || got[1] != b {
					t.Errorf("%s: extract(%q) = %v, want [%s %s]", codec.name, id, got, a, b)
				}
			}
		}
	}
}

// checkCodecSafety asserts the parser's safety rules for one (a, b) pair:
// it must succeed exactly when both segments are non-empty and slash-free, and
// whenever it succeeds it must return exactly (a, b).
func checkCodecSafety(t *testing.T, codec resourceIDCodec, a, b string) {
	t.Helper()
	id := codec.format(a, b)
	got, err := codec.extract(id)

	valid := a != "" && b != "" && !strings.Contains(a, "/") && !strings.Contains(b, "/")
	if valid && err != nil {
		t.Errorf("%s: extract(%q) failed for valid segments (%q, %q): %v", codec.name, id, a, b, err)
		return
	}
	if !valid && err == nil {
		t.Errorf("%s: extract(%q) succeeded for invalid segments (%q, %q)", codec.name, id, a, b)
	}
	if err == nil && (len(got) != 2 || got[0] != a || got[1] != b) {
		t.Errorf("%s: extract(%q) = %v without error, want [%s %s]", codec.name, id, got, a, b)
	}
}

// TestResourceIDCodecs_exhaustiveSmallAlphabet applies the safety check to every
// pair of strings up to length 3 over an alphabet containing separators and tokens.
func TestResourceIDCodecs_exhaustiveSmallAlphabet(t *testing.T) {
	alphabet := []string{"a", "/", "-"}
	var segments []string
	frontier := []string{""}
	segments = append(segments, frontier...)
	for n := 0; n < 3; n++ {
		var next []string
		for _, s := range frontier {
			for _, c := range alphabet {
				next = append(next, s+c)
			}
		}
		segments = append(segments, next...)
		frontier = next
	}

	for _, codec := range resourceIDCodecs {
		for _, a := range segments {
			for _, b := range segments {
				checkCodecSafety(t, codec, a, b)
			}
		}
	}
}

// FuzzResourceIDCodecs applies the same safety property to arbitrary inputs.
func FuzzResourceIDCodecs(f *testing.F) {
	f.Add("a", "b")
	f.Add("a/b", "c")
	f.Add("", "")
	f.Add("access", "upstream")
	f.Fuzz(func(t *testing.T, a string, b string) {
		for _, codec := range resourceIDCodecs {
			checkCodecSafety(t, codec, a, b)
		}
	})
}
