//go:build integration
// +build integration

package main

import (
	"testing"
)

// TestMainFunction wraps the main driver of the program
// in a test function. This test is used to enable the
// capture of coverage metrics while the service is
// running.
func TestMainFunction(t *testing.T) {
	main()
}
