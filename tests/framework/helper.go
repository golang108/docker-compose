/*
	Copyright (c) 2020 Docker Inc.

	Permission is hereby granted, free of charge, to any person
	obtaining a copy of this software and associated documentation
	files (the "Software"), to deal in the Software without
	restriction, including without limitation the rights to use, copy,
	modify, merge, publish, distribute, sublicense, and/or sell copies
	of the Software, and to permit persons to whom the Software is
	furnished to do so, subject to the following conditions:

	The above copyright notice and this permission notice shall be
	included in all copies or substantial portions of the Software.

	THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND,
	EXPRESS OR IMPLIED,
	INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
	FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.
	IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT
	HOLDERS BE LIABLE FOR ANY CLAIM,
	DAMAGES OR OTHER LIABILITY,
	WHETHER IN AN ACTION OF CONTRACT,
	TORT OR OTHERWISE,
	ARISING FROM, OUT OF OR IN CONNECTION WITH
	THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
*/

package framework

import (
	"runtime"
	"strings"

	"github.com/robpike/filter"
	log "github.com/sirupsen/logrus"
)

func nonEmptyString(s string) bool {
	return strings.TrimSpace(s) != ""
}

// Lines get lines from a raw string
func Lines(output string) []string {
	return filter.Choose(strings.Split(output, "\n"), nonEmptyString).([]string)
}

// Columns get columns from a line
func Columns(line string) []string {
	return filter.Choose(strings.Split(line, " "), nonEmptyString).([]string)
}

// GoldenFile golden file specific to platform
func GoldenFile(name string) string {
	if IsWindows() {
		return name + "-windows.golden"
	}
	return name + ".golden"
}

// IsWindows windows or other GOOS
func IsWindows() bool {
	return runtime.GOOS == "windows"
}

// It runs func
func It(description string, test func()) {
	test()
	log.Print("Passed: ", description)
}
