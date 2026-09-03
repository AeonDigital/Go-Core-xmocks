package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestRun_SuccessGeneratesMock(t *testing.T) {
	tmpDir := t.TempDir()
	inputFile := filepath.Join(tmpDir, "input.go")
	content := `package sample

type IExample interface {
	ReadFile(name string) ([]byte, error)
}
`
	if err := os.WriteFile(inputFile, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	outputDir := filepath.Join(tmpDir, "output")

	exit := run([]string{"--file", inputFile, "--interface", "IExample", "--alias", "Example", "--output", outputDir}, stdout, stderr)
	if exit != 0 {
		t.Fatalf("expected exit 0, got %d; stderr=%s", exit, stderr.String())
	}

	if !strings.Contains(stdout.String(), "Mock generated:") {
		t.Fatalf("expected success message, got %s", stdout.String())
	}

	generatedPath := filepath.Join(outputDir, "example.go")
	if _, err := os.Stat(generatedPath); err != nil {
		t.Fatalf("expected generated file %s to exist: %v", generatedPath, err)
	}

	generated, err := os.ReadFile(generatedPath)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(string(generated), "type MockExample struct") {
		t.Fatal("generated file missing MockExample struct")
	}
}

func TestRun_MissingRequiredFlags(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	exit := run([]string{}, stdout, stderr)
	if exit != 1 {
		t.Fatalf("expected exit 1 for missing flags, got %d", exit)
	}

	if !strings.Contains(stderr.String(), "error: both --file and --interface are required") {
		t.Fatalf("expected missing flags error, got %s", stderr.String())
	}
}

func TestRun_ParseErrorReturnsTwo(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	exit := run([]string{"--unknown-flag"}, stdout, stderr)
	if exit != 2 {
		t.Fatalf("expected exit 2 for parse error, got %d", exit)
	}

	if !strings.Contains(stderr.String(), "flag provided but not defined") {
		t.Fatalf("expected parse error output, got %s", stderr.String())
	}
}

func TestMainFunctionExecutes(t *testing.T) {
	tmpDir := t.TempDir()
	inputFile := filepath.Join(tmpDir, "input.go")
	content := `package sample

type IExample interface {
	ReadFile(name string) ([]byte, error)
}
`
	if err := os.WriteFile(inputFile, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	outputDir := filepath.Join(tmpDir, "output")
	cmd := exec.Command("go", "run", "./xmocks", "--file", inputFile, "--interface", "IExample", "--alias", "Example", "--output", outputDir)
	cmd.Dir = ".."
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		t.Fatalf("expected main command to succeed, stderr=%s, err=%v", stderr.String(), err)
	}

	if !strings.Contains(stdout.String(), "Mock generated:") {
		t.Fatalf("expected success message from main, got %s", stdout.String())
	}

	generatedPath := filepath.Join(outputDir, "example.go")
	if _, err := os.Stat(generatedPath); err != nil {
		t.Fatalf("expected generated file %s to exist: %v", generatedPath, err)
	}
}

func TestMainFunctionViaSubprocess(t *testing.T) {
	if os.Getenv("TEST_MAIN") == "1" {
		main()
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestMainFunctionViaSubprocess")
	cmd.Env = append(os.Environ(), "TEST_MAIN=1")
	cmd.Stdout = &bytes.Buffer{}
	cmd.Stderr = &bytes.Buffer{}

	err := cmd.Run()
	if err == nil {
		t.Fatal("expected subprocess to exit with an error status")
	}

	exitErr, ok := err.(*exec.ExitError)
	if !ok {
		t.Fatalf("expected exec.ExitError, got %T: %v", err, err)
	}

	if exitErr.ExitCode() != 2 {
		t.Fatalf("expected exit code 2 from main subprocess, got %d", exitErr.ExitCode())
	}
}

func TestRun_InvalidInterfaceName(t *testing.T) {
	tmpDir := t.TempDir()
	inputFile := filepath.Join(tmpDir, "input.go")
	content := `package sample

type IExample interface {
	NoOp()
}
`
	if err := os.WriteFile(inputFile, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	exit := run([]string{"--file", inputFile, "--interface", "IMissing"}, stdout, stderr)
	if exit != 1 {
		t.Fatalf("expected exit 1 for invalid interface name, got %d", exit)
	}

	if !strings.Contains(stderr.String(), "interface IMissing not found") {
		t.Fatalf("expected interface not found error, got %s", stderr.String())
	}
}
