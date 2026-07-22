// Copyright (c) 2026, The OTNS Authors.
// All rights reserved.
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted provided that the following conditions are met:
// 1. Redistributions of source code must retain the above copyright
//    notice, this list of conditions and the following disclaimer.
// 2. Redistributions in binary form must reproduce the above copyright
//    notice, this list of conditions and the following disclaimer in the
//    documentation and/or other materials provided with the distribution.
// 3. Neither the name of the copyright holder nor the
//    names of its contributors may be used to endorse or promote products
//    derived from this software without specific prior written permission.
//
// THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS"
// AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
// IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE
// ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE
// LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR
// CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF
// SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS
// INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN
// CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE)
// ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE
// POSSIBILITY OF SUCH DAMAGE.

package logger

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	tmpDir string
)

func init() {
	tmpDir = filepath.Join(os.TempDir(), "otns-logger-test-424387")
}

func cleanupLogger() {
	if logFileHandle != nil {
		_ = logFileHandle.Close()
		logFileHandle = nil // Reset package variable
	}
	zaplogger = nil
	currentLevel = DefaultLevel
	isLogToTerminal = true
	cbStdout = nil
	logFileHandle = nil
	logPath = ""
	cfg.OutputPaths = []string{"stderr"}
	rebuildLoggerFromCfg()
}

func TestClearExistingLogFile(t *testing.T) {
	t.Cleanup(cleanupLogger)

	// Create temporary directory for the test
	err := os.Mkdir(tmpDir, 0755)
	assert.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create the initial log file
	logFile := filepath.Join(tmpDir, "test.log")
	err = os.WriteFile(logFile, []byte("initial content\n"), 0644)
	assert.NoError(t, err)

	// Make the directory non-writable, which should prevent removing the file on Linux
	err = os.Chmod(tmpDir, 0555)
	assert.NoError(t, err)
	defer func() { _ = os.Chmod(tmpDir, 0755) }() // Ensure we can clean up

	// Call Init. It should be able to open it and clear file contents.
	Init(true, true, logFile, 0)

	// Verify that the file handle is set (it should be, since the file exists and is writable)
	assert.NotNil(t, logFileHandle)

	// Verify we can log to the file
	Infof("Test log message 1234")

	// Check that the file contains the logged line
	content, err := os.ReadFile(logFile)
	assert.NoError(t, err)
	assert.Contains(t, string(content), "Test log message 1234")

	// Check that the file was cleared.
	assert.NotContains(t, string(content), "initial content")
}

func TestInitReadOnlyFile(t *testing.T) {
	t.Cleanup(cleanupLogger)

	// Create temporary directory for the test
	err := os.Mkdir(tmpDir, 0755)
	assert.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create the initial log file and make it read-only
	logFile := filepath.Join(tmpDir, "test.log")
	err = os.WriteFile(logFile, []byte("initial content\n"), 0444)
	assert.NoError(t, err)

	// Call Init. It should attempt to remove the file.
	// Removing a read-only file in a writable directory works on Linux.
	// So let's make the directory non-writable too, to ensure remove fails,
	// AND the file is read-only, so opening it for WRONLY also fails.
	err = os.Chmod(tmpDir, 0555)
	assert.NoError(t, err)
	defer func() { _ = os.Chmod(tmpDir, 0755) }()

	Init(true, true, logFile, 0)

	// Verify that the file handle is nil because file-open should have failed.
	assert.Nil(t, logFileHandle)

	// Verify we can log
	Infof("Test log message 4567")
}

func TestPrintln(t *testing.T) {
	t.Cleanup(cleanupLogger)

	// Create temporary directory for the test
	err := os.Mkdir(tmpDir, 0755)
	assert.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Initialize logger (including file)
	logFile := filepath.Join(tmpDir, "test-levels.log")
	Init(false, true, logFile, 0)

	// Log messages
	Println("A: should be logged", false, true)
	Println("B: should be logged", true, true)
	Println("C: should NOT be logged", true, false)
	Println("D: should NOT be logged", false, false)

	// Verify content
	content, err := os.ReadFile(logFile)
	assert.NoError(t, err)
	s := string(content)
	assert.Contains(t, s, "A: should be logged")
	assert.Contains(t, s, "B: should be logged")
	assert.NotContains(t, s, "C: should NOT be logged")
	assert.NotContains(t, s, "D: should NOT be logged")

	// More log messages
	Println("E: should be logged now", true, true)

	content, err = os.ReadFile(logFile)
	assert.NoError(t, err)
	s = string(content)
	assert.Contains(t, s, "E: should be logged now")
}

func TestGetLogWriter(t *testing.T) {
	t.Cleanup(cleanupLogger)

	// Case 1: logFileHandle is nil
	writer := GetLogWriter()
	assert.NotNil(t, writer)
	_, err := writer.Write([]byte("test data when nil\n"))
	assert.NoError(t, err)

	// Case 2: logFileHandle is open
	err = os.Mkdir(tmpDir, 0755)
	assert.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	logFile := filepath.Join(tmpDir, "test-writer.log")
	Init(false, true, logFile, 0)

	writer = GetLogWriter()
	assert.NotNil(t, writer)

	// Case 2a: Write message that should be logged to file
	msg := "test data to writer"
	_, err = writer.Write([]byte(msg + "\n"))
	assert.NoError(t, err)

	content, err := os.ReadFile(logFile)
	assert.NoError(t, err)
	assert.Contains(t, string(content), msg)
}
