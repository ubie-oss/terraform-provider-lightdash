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
	"context"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strings"

	"github.com/hashicorp/terraform-plugin-log/tflog"
)

const (
	integrationTestModeEnvVar  = "TF_ACC"
	lightdashApiKeyEnvVar      = "LIGHTDASH_API_KEY" // #nosec G101
	lightdashProjectUuidEnvVar = "LIGHTDASH_PROJECT" // #nosec G101
	lightdashUrlEnvVar         = "LIGHTDASH_URL"
)

func isIntegrationTestMode() bool {
	// If the environment variable is set to 1, then we are in test mode
	test_mode := os.Getenv(integrationTestModeEnvVar)
	return test_mode == "1"
}

func getLightdashApiKey() (*string, error) {
	// If the environment variable is set to 1, then we are in test mode
	api_key := strings.TrimSpace(os.Getenv(lightdashApiKeyEnvVar))
	fmt.Println("api_key", api_key)
	if api_key == "" {
		return nil, fmt.Errorf("LIGHTDASH_API_KEY environment variable is not set")
	}
	return &api_key, nil
}

func getLightdashProjectUuid() (*string, error) {
	// If the environment variable is set to 1, then we are in test mode
	projectUuid := strings.TrimSpace(os.Getenv(lightdashProjectUuidEnvVar))
	fmt.Println("projectUuid", projectUuid)
	if projectUuid == "" {
		return nil, fmt.Errorf("LIGHTDASH_PROJECT environment variable is not set")
	}
	return &projectUuid, nil
}

func getLightdashUrl() (*string, error) {
	// If the environment variable is set to 1, then we are in test mode
	url := strings.TrimSpace(os.Getenv(lightdashUrlEnvVar))
	if url == "" {
		return nil, fmt.Errorf("LIGHTDASH_URL environment variable is not set")
	}
	return &url, nil
}

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

// Get the path to the acc_tests directory relative to the current file.
func getPathToAccTests() (string, error) {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		return "", fmt.Errorf("failed to get current file path")
	}

	// Get the directory of the current file (utils.go)
	currentDir := path.Dir(filename)

	// Navigate up two directories (internal/provider -> internal -> .) and then down into acc_tests
	accTestsPath := path.Join(currentDir, "acc_tests")

	// Check if the directory exists
	if _, err := os.Stat(accTestsPath); os.IsNotExist(err) {
		return "", fmt.Errorf("acc_tests directory does not exist at %s", accTestsPath)
	}
	return accTestsPath, nil
}

// Get the path to an acc_test resource
func getPathToAccTestResource(elements []string) (string, error) {
	pathToAccTests, err := getPathToAccTests()
	if err != nil {
		return "", err
	}

	// Combine the base path with the elements from the slice
	// Use path.Join to construct the path safely
	allElements := append([]string{pathToAccTests}, elements...)
	accTestResourcePath := path.Join(allElements...)

	// Add a security check: ensure the constructed path is within the acc_tests directory
	// Use filepath.Clean to normalize paths before comparison
	cleanedAccTestsPath := path.Clean(pathToAccTests)
	cleanedResourcePath := path.Clean(accTestResourcePath)

	// Check if the cleaned resource path is a sub-path of the cleaned acc_tests path
	// This prevents path traversal attacks using '..'
	if !strings.HasPrefix(cleanedResourcePath, cleanedAccTestsPath) {
		return "", fmt.Errorf("attempted to access file outside acc_tests directory: %s", accTestResourcePath)
	}

	// Also check that the constructed path actually exists
	if _, err := os.Stat(accTestResourcePath); os.IsNotExist(err) {
		return "", fmt.Errorf("acc_tests resource does not exist at %s", accTestResourcePath)
	}
	return accTestResourcePath, nil
}

func ReadAccTestResource(elements []string) (string, error) {
	path, err := getPathToAccTestResource(elements)
	if err != nil {
		return "", err
	}
	resource, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		return "", err
	}
	return string(resource), nil
}

// Create a provider config string
func getProviderConfig() (string, error) {
	// Get the Lightdash URL
	lightdashUrl, err := getLightdashUrl()
	if err != nil {
		return "", err
	}
	// Get the Lightdash API key
	lightdashApiKey, err := getLightdashApiKey()
	if err != nil {
		return "", err
	}
	lightdashProjectUuid, err := getLightdashProjectUuid()
	if err != nil {
		return "", err
	}

	// Create the provider config string
	providerConfig := fmt.Sprintf(`
provider "lightdash" {
	host  = "%s"
	token = "%s"
}

data "lightdash_project" "test" {
	project_uuid = "%s"
}
`, *lightdashUrl, *lightdashApiKey, *lightdashProjectUuid)
	return providerConfig, nil
}

// Subtract list2 from list1
func subtractStringList(list1, list2 []string) []string {
	// Create a frequency map of the second list
	list2Map := make(map[string]int)
	for _, item := range list2 {
		list2Map[item]++
	}

	result := []string{}
	// Iterate through the first list and add items if their count in list2Map is zero or less
	for _, item := range list1 {
		if list2Map[item] > 0 {
			list2Map[item]--
		} else {
			result = append(result, item)
		}
	}

	// Sort the result
	sort.Strings(result)

	return result
}

// readMarkdownDescription reads the content of a markdown file located relative to the calling Go file.
func readMarkdownDescription(ctx context.Context, filename string) (string, error) {
	// Get the file path of the caller. We use Caller(1) because Caller(0) is this function itself.
	_, callerFile, _, ok := runtime.Caller(1)
	if !ok {
		return "", fmt.Errorf("failed to get caller file path")
	}

	// Get the directory of the caller file
	callerDir := filepath.Dir(callerFile)

	// Construct the full path to the markdown file relative to the caller's directory.
	// Assuming the markdown files are in internal/provider/docs relative to the project root,
	// and the caller is in internal/provider. The path from callerDir to the docs directory is ../docs.
	// The 'filename' parameter is expected to be relative from the 'provider' directory, e.g., "docs/resources/resource_name.md".
	// Let's adjust the logic to handle the full path passed currently, like "internal/provider/docs/...".
	// A simpler approach is to find the 'provider' directory and then join the rest of the path.

	// Navigate up directories from the caller's file to find the 'provider' directory
	dir := callerDir
	for {
		if filepath.Base(dir) == "provider" {
			break
		}
		parentDir := filepath.Dir(dir)
		if parentDir == dir {
			// If we can't find 'provider', fall back to the original behavior or return an error
			// For now, let's assume we can find 'provider' or the relative path is from project root.
			// The previous error indicates the relative path from project root doesn't work in gen-docs.
			// So finding the 'provider' directory is the more reliable approach.
			tflog.Error(ctx, fmt.Sprintf("Could not find 'provider' directory for caller %s", callerFile))
			// Fallback to the original problematic behavior for now, and log the error.
			// This should ideally be a fatal error if the 'provider' directory structure is expected.
			// Let's return an informative error.
			return "", fmt.Errorf("could not find 'provider' directory in path for caller %s", callerFile)
		}
		dir = parentDir
	}
	providerDir := dir

	// Extract the path relative to the 'provider' directory from the input 'filename'
	// We expect filename to start with "internal/provider/", so we strip that.
	relativePathInProvider := strings.TrimPrefix(filename, "internal/provider/")

	// Construct the full path by joining the provider directory and the relative path within it
	fullPath := filepath.Join(providerDir, relativePathInProvider)

	tflog.Debug(ctx, fmt.Sprintf("Attempting to read markdown file from calculated path: %s (original filename: %s)", fullPath, filename))

	content, err := os.ReadFile(filepath.Clean(fullPath))
	if err != nil {
		// Log the error with the full path that failed
		tflog.Error(ctx, fmt.Sprintf("Error reading markdown file %s: %s", fullPath, err.Error()))
		return "", fmt.Errorf("failed to read markdown file %s (tried path: %s): %w", filename, fullPath, err)
	}

	return string(content), nil
}
