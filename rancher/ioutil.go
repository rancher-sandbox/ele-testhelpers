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
	"io"
	"os"
	"strings"
)

/**
 * Add data to file
 * @remarks Added data is optional
 * @param srcFile Source file
 * @param dstFile Destination file
 * @param data Data to add at the end of destination file
 * @returns Nothing or an error
 */
func AddDataToFile(srcfile, dstfile string, data []byte) error {
	// Open source file
	f, err := os.Open(srcfile)
	if err != nil {
		return err
	}
	defer f.Close()

	// Open/create destination file
	d, err := os.OpenFile(dstfile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer d.Close()

	// Copy source to dest
	if _, err = io.Copy(d, f); err != nil {
		return err
	}

	// Add data to dest
	if _, err = d.Write([]byte(data)); err != nil {
		return err
	}

	return nil
}

/**
 * Write data to file
 * @remarks This function simply writes data to a file
 * @param dstFile Destination file
 * @param data Data to write
 * @returns Nothing or an error
 */
func WriteFile(dstfile string, data []byte) error {
	// Open/create destination file
	d, err := os.OpenFile(dstfile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer d.Close()

	// Add data to dest
	if _, err = d.Write([]byte(data)); err != nil {
		return err
	}

	return nil
}

/**
 * Copy file
 * @remarks This function simply copies a file to another
 * @param srcFile Source file
 * @param dstFile Destination file
 * @returns Nothing or an error
 */
func CopyFile(srcFile, dstFile string) error {
	// Add data to file without adding data(!) is in fact a copy
	return (AddDataToFile(srcFile, dstFile, []byte("")))
}

/**
 * Trim a string
 * @remarks Remove all from s string after c char
 * @param s String to trim
 * @param c Character used as a separator
 * @returns Trimmed string
 */
func TrimStringFromChar(s, c string) string {
	if idx := strings.Index(s, c); idx != -1 {
		return s[:idx]
	}

	return s
}

/**
 * Create a temporary file
 * @remarks Create temporary file with a name based on baseName
 * @param baseName String to trim
 * @returns Created filename or an error
 */
func CreateTemp(baseName string) (string, error) {
	t, err := os.CreateTemp("", baseName)
	if err != nil {
		return "", err
	}

	return t.Name(), nil
}
