package log

import (
	"fmt"
	"os"
)

var (
	// Verbose determines whether to print verbose output
	Verbose bool
)

// Info prints a message to stdout
func Info(format string, a ...interface{}) {
	fmt.Printf(format, a...)
}

// Infof prints a formatted message to stdout
func Infof(format string, a ...interface{}) {
	fmt.Fprintf(os.Stdout, format, a...)
}

// Warnf prints a formatted warning message to stderr
func Warnf(format string, a ...interface{}) {
	fmt.Fprintf(os.Stderr, "Warning: %s\n", fmt.Sprintf(format, a...))
}

// Errorf prints a formatted error message to stderr
func Errorf(format string, a ...interface{}) {
	fmt.Fprintf(os.Stderr, "Error: %s\n", fmt.Sprintf(format, a...))
}

// Debugf prints a formatted debug message if verbose is enabled
func Debugf(format string, a ...interface{}) {
	if Verbose {
		fmt.Fprintf(os.Stderr, "Debug: %s\n", fmt.Sprintf(format, a...))
	}
}

// Fatalf prints a formatted error message and exits
func Fatalf(format string, a ...interface{}) {
	Errorf(format, a...)
	os.Exit(1)
}
