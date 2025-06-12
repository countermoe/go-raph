package main

import (
	"flag"
	"os"
	"testing"
)

func TestFlagDefaults(t *testing.T) {
	// Reset flags for testing
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	// Reset global variables
	targetPath = ""

	// Simulate no arguments
	os.Args = []string{"go-raph"}

	var port *string
	flag.StringVar(&targetPath, "path", ".", "Path to analyze")
	port = flag.String("port", "8080", "Server port")
	flag.Parse()

	if targetPath != "." {
		t.Errorf("Expected default targetPath to be '.', got '%s'", targetPath)
	}

	if *port != "8080" {
		t.Errorf("Expected default port to be '8080', got '%s'", *port)
	}
}

func TestPathFlag(t *testing.T) {
	// Reset flags for testing
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	// Reset global variables
	targetPath = ""

	// Simulate -path flag
	os.Args = []string{"go-raph", "-path", "/some/test/path"}

	var port *string
	flag.StringVar(&targetPath, "path", ".", "Path to analyze")
	port = flag.String("port", "8080", "Server port")
	flag.Parse()

	if targetPath != "/some/test/path" {
		t.Errorf("Expected targetPath to be '/some/test/path', got '%s'", targetPath)
	}

	if *port != "8080" {
		t.Errorf("Expected default port to be '8080', got '%s'", *port)
	}
}

func TestPortFlag(t *testing.T) {
	// Reset flags for testing
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	// Reset global variables
	targetPath = ""

	// Simulate -port flag
	os.Args = []string{"go-raph", "-port", "9000"}

	var port *string
	flag.StringVar(&targetPath, "path", ".", "Path to analyze")
	port = flag.String("port", "8080", "Server port")
	flag.Parse()

	if targetPath != "." {
		t.Errorf("Expected default targetPath to be '.', got '%s'", targetPath)
	}

	if *port != "9000" {
		t.Errorf("Expected port to be '9000', got '%s'", *port)
	}
}

func TestBothFlags(t *testing.T) {
	// Reset flags for testing
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	// Reset global variables
	targetPath = ""

	// Simulate both flags
	os.Args = []string{"go-raph", "-path", "/custom/path", "-port", "3000"}

	var port *string
	flag.StringVar(&targetPath, "path", ".", "Path to analyze")
	port = flag.String("port", "8080", "Server port")
	flag.Parse()

	if targetPath != "/custom/path" {
		t.Errorf("Expected targetPath to be '/custom/path', got '%s'", targetPath)
	}

	if *port != "3000" {
		t.Errorf("Expected port to be '3000', got '%s'", *port)
	}
}

func TestPositionalArgument(t *testing.T) {
	// Reset flags for testing
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	// Reset global variables
	targetPath = ""

	// Simulate positional argument
	os.Args = []string{"go-raph", "/positional/path"}

	var port *string
	flag.StringVar(&targetPath, "path", ".", "Path to analyze")
	port = flag.String("port", "8080", "Server port")
	flag.Parse()

	// Simulate the logic from main()
	if len(flag.Args()) > 0 {
		targetPath = flag.Args()[0]
	}

	if targetPath != "/positional/path" {
		t.Errorf("Expected targetPath to be '/positional/path', got '%s'", targetPath)
	}

	if *port != "8080" {
		t.Errorf("Expected default port to be '8080', got '%s'", *port)
	}
}

func TestPositionalOverridesFlag(t *testing.T) {
	// Reset flags for testing
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	// Reset global variables
	targetPath = ""

	// Simulate both flag and positional argument
	os.Args = []string{"go-raph", "-path", "/flag/path", "/positional/path"}

	var port *string
	flag.StringVar(&targetPath, "path", ".", "Path to analyze")
	port = flag.String("port", "8080", "Server port")
	flag.Parse()

	// Simulate the logic from main()
	if len(flag.Args()) > 0 {
		targetPath = flag.Args()[0]
	}

	if targetPath != "/positional/path" {
		t.Errorf("Expected positional arg to override flag, got '%s'", targetPath)
	}

	// Use port to avoid linter error
	_ = *port
}

func TestComplexArguments(t *testing.T) {
	// Reset flags for testing
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	// Reset global variables
	targetPath = ""

	// Simulate complex arguments with spaces and special chars
	os.Args = []string{"go-raph", "-port", "8888", "/path/with spaces/and-dashes"}

	var port *string
	flag.StringVar(&targetPath, "path", ".", "Path to analyze")
	port = flag.String("port", "8080", "Server port")
	flag.Parse()

	// Simulate the logic from main()
	if len(flag.Args()) > 0 {
		targetPath = flag.Args()[0]
	}

	if targetPath != "/path/with spaces/and-dashes" {
		t.Errorf("Expected complex path to be parsed correctly, got '%s'", targetPath)
	}

	if *port != "8888" {
		t.Errorf("Expected port to be '8888', got '%s'", *port)
	}
}

func TestEmptyPath(t *testing.T) {
	// Reset flags for testing
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	// Reset global variables
	targetPath = ""

	// Simulate empty path flag
	os.Args = []string{"go-raph", "-path", ""}

	var port *string
	flag.StringVar(&targetPath, "path", ".", "Path to analyze")
	port = flag.String("port", "8080", "Server port")
	flag.Parse()

	if targetPath != "" {
		t.Errorf("Expected empty targetPath to be '', got '%s'", targetPath)
	}

	// Use port to avoid linter error and add note about empty path handling
	_ = *port
	// Note: In main(), we might want to handle empty path by defaulting to "."
}

// Benchmark flag parsing performance
func BenchmarkFlagParsing(b *testing.B) {
	for i := 0; i < b.N; i++ {
		// Reset flags for benchmarking
		flag.CommandLine = flag.NewFlagSet("go-raph", flag.ExitOnError)

		// Reset global variables
		targetPath = ""

		os.Args = []string{"go-raph", "-path", "/test/path", "-port", "8080"}

		var port *string
		flag.StringVar(&targetPath, "path", ".", "Path to analyze")
		port = flag.String("port", "8080", "Server port")
		flag.Parse()

		// Simulate main() logic
		if len(flag.Args()) > 0 {
			targetPath = flag.Args()[0]
		}

		_ = *port // Use the port to avoid optimization
	}
}
