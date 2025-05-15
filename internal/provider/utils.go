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
	"os"
	"path"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
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
