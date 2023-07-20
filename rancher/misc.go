/*
Copyright © 2022 - 2023 SUSE LLC

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

package rancher

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

/**
 * Configure a timeout based on TIMEOUT_SCALE env variable
 * @remarks Multiply a duration with TIMEOUT_SCALE value
 * @param timeout Initial timeout value
 * @returns Modified timeout value
 */
func SetTimeout(timeout time.Duration) time.Duration {
	s, set := os.LookupEnv("TIMEOUT_SCALE")

	// Only if TIMEOUT_SCALE is set
	if set {
		scale, err := strconv.Atoi(s)
		if err != nil {
			return timeout
		}

		// Return the scaled timeout
		return timeout * time.Duration(scale)
	}

	// Don't return error, in the worst case return the initial value
	// Otherwise an additional step will be needed for some commands (like Eventually)
	return timeout
}

/**
 * Set hostname of the node
 * @remarks Define the hostname base on baseName and node index
 * @param baseName Basename to use, "empty" if nothing provided
 * @param index index of the node
 * @returns Full hostname of the node
 */
func SetHostname(baseName string, index int) string {
	if baseName == "" {
		baseName = "emtpy"
	}

	if index < 0 {
		index = 0
	}

	return baseName + "-" + fmt.Sprintf("%03d", index)
}
