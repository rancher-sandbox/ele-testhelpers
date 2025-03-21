/*
Copyright © 2022 - 2025 SUSE LLC

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

package tools

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	scp "github.com/bramvdbogaerde/go-scp"
	"github.com/pkg/errors"
	ssh "golang.org/x/crypto/ssh"
)

/**
 * Get File through HTTP
 * @param url URL where to download the file
 * @param fileName of the file to create
 * @param skipVerify Skip TLS check if needed
 * @returns Nothing or an error
 */
func GetFileFromURL(url string, fileName string, skipVerify bool) error {
	if !skipVerify {
		http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}

	data, err := http.Get(url)
	if err != nil {
		return err
	}
	defer data.Body.Close()

	// Create file
	file, err := os.Create(fileName)
	if err != nil {
		return err
	}

	// Save data in file
	_, err = io.Copy(file, data.Body)
	return err
}

/**
 * Get a list of files
 * @param dir Directory where to search
 * @param pattern Search pattern
 * @returns List of files or an error
 */
func GetFilesList(dir string, pattern string) ([]string, error) {
	files, err := filepath.Glob(dir + "/" + pattern)
	if err != nil {
		return nil, err
	}

	if files != nil {
		return files, nil
	}

	return nil, err
}

/**
 * Simple sed command
 * @param oldValue Value or simple regex to modify
 * @param newValue New value to set
 * @param file File to modify
 * @returns Nothing or an error
 */
// Sed code partially from https://forum.golangbridge.org/t/using-sed-in-golang/23526/16
func Sed(oldValue, newValue, file string) error {
	fileData, err := os.ReadFile(file)
	if err != nil {
		return err
	}

	// Get file permissions
	info, err := os.Stat(file)
	if err != nil {
		return err
	}
	mode := info.Mode()

	// Regex is in the old value var
	regex := regexp.MustCompile(oldValue)
	fileString := string(fileData)
	fileString = regex.ReplaceAllString(fileString, newValue)
	fileData = []byte(fileString)

	err = os.WriteFile(file, fileData, mode)
	return err
}

type Client struct {
	Host     string
	Username string
	Password string
}

/**
 * Define SSH client
 * @remarks This function is only used internally, not exported
 * @returns SSH Client configuration
 */
// NOTE: clientConfig does not have unit test as it is
// used only in connectToHost
func (c *Client) clientConfig() *ssh.ClientConfig {
	sshConfig := &ssh.ClientConfig{
		User:            c.Username,
		Auth:            []ssh.AuthMethod{ssh.Password(c.Password)},
		Timeout:         30 * time.Second,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	return sshConfig
}

/**
 * Connect on the client through SSH
 * @remarks This function is only used internally, not exported
 * @param Client is the receiver where to execute the command
 * @param ssh.Client Client definition to connect to
 * @returns Pointer to the SSH Client or an error
 */
// NOTE: connectToHost does not have unit test as it is
// used in RunSSH which is already tested
func (c *Client) connectToHost() (*ssh.Client, error) {
	// Define ssh connection
	sshConfig := c.clientConfig()

	// Connect to client
	sshClient, err := ssh.Dial("tcp", c.Host, sshConfig)
	if err != nil {
		return nil, err
	}

	return sshClient, nil
}

/**
 * Get file through SSH (a SCP in fact!)
 * @param localFile Local file to create
 * @param remoteFile Remote file to copy from
 * @param perm Permissions to set on the local file
 * @returns Nothing or an error
 */
func (c *Client) GetFile(localFile, remoteFile string, perm fs.FileMode) error {
	// Define ssh connection
	sshConfig := c.clientConfig()

	// Create a local file to write to.
	f, err := os.OpenFile(localFile, os.O_RDWR|os.O_CREATE, perm)
	if err != nil {
		// Failed to open
		return err
	}
	defer f.Close()

	// Connect to client
	scpClient := scp.NewClient(c.Host, sshConfig)
	defer scpClient.Close()

	if err := scpClient.Connect(); err != nil {
		// Failed to connect
		return err
	}

	if err := scpClient.CopyFromRemote(context.Background(), f, remoteFile); err != nil {
		// Failed to copy
		return err
	}
	return nil
}

/**
 * Run a command on client through SSH
 * @param Client is the receiver where to execute the command
 * @param cmd Command to execute
 * @returns Result of the command or an error
 */
func (c *Client) RunSSH(cmd string) (string, error) {
	sshClient, err := c.connectToHost()
	if err != nil {
		// Failed to connect
		return "", err
	}
	defer sshClient.Close()

	// Open a client session
	session, err := sshClient.NewSession()
	if err != nil {
		// Failed to create session
		return "", err
	}
	defer session.Close()

	// Execute the command on the opened session
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	session.Stdout = &stdout
	session.Stderr = &stderr
	if err := session.Run(cmd); err != nil {
		// Failed to execute the command
		return stdout.String(), errors.Wrap(err, stderr.String())
	}
	return stdout.String(), nil
}

/**
 * Send file through SSH (a SCP in fact!)
 * @param src Source file
 * @param dst Destination file on the client
 * @param perm Permissions to set on the file
 * @returns Nothing or an error
 */
func (c *Client) SendFile(src, dst, perm string) error {
	// Define ssh connection
	sshConfig := c.clientConfig()

	// Connect to client
	scpClient := scp.NewClient(c.Host, sshConfig)
	defer scpClient.Close()

	if err := scpClient.Connect(); err != nil {
		// Failed to connect
		return err
	}

	f, err := os.Open(src)
	if err != nil {
		// Failed to open
		return err
	}
	defer f.Close()

	if err := scpClient.CopyFromFile(context.Background(), *f, dst, perm); err != nil {
		// Failed to copy
		return err
	}
	return nil
}

/**
 * Share files through HTTP (simple way, no security at all!)
 * @remarks A HTTP server is up and running
 * @param directory The directory where is files are
 * @param listenAddr Port where to listen to
 * @returns Nothing
 */
// TODO: Use Server function from http helpers instead
func HTTPShare(directory, listenAddr string) {
	fs := http.FileServer(http.Dir(directory))

	go func() {
		if err := http.ListenAndServe(listenAddr, fs); err != nil {
			fmt.Printf("Server failed: %s\n", err)
		}
	}()
}

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
// NOTE: CopyFile does not have unit test as it uses
// AddDataToFile which is already tested
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

/**
 * Check if the string is a valid IPv4
 * @param value Value to check
 * @returns True if value is an IPv4, otherwise False
 */
func IsIPv4(value string) bool {
	const regex = `^((25[0-5]|(2[0-4]|1\d|[1-9]|)\d)\.?\b){4}$`
	return regexp.MustCompile(regex).MatchString(value)
}
