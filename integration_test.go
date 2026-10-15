package main

import (
	"os"
	"os/exec"
	"strings"
	"testing"
)

func buildBinary(t *testing.T) string {
	t.Helper()
	binary := t.TempDir() + "/sow"
	cmd := exec.Command("go", "build", "-o", binary, ".")
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("build failed: %v", err)
	}
	return binary
}

func TestHelpOutput(t *testing.T) {
	bin := buildBinary(t)
	out, err := exec.Command(bin, "--help").CombinedOutput()
	if err != nil {
		// --help returns exit 0 in cobra
	}
	output := string(out)
	if !strings.Contains(output, "interactive, colorful charts") {
		t.Error("help should contain description")
	}
	for _, sub := range []string{"bar", "line", "pie", "scatter", "sparkline", "streamline", "heatmap", "table", "ping", "vmstat"} {
		if !strings.Contains(output, sub) {
			t.Errorf("help should list subcommand %q", sub)
		}
	}
}

func TestVersionOutput(t *testing.T) {
	bin := buildBinary(t)
	out, _ := exec.Command(bin, "--version").CombinedOutput()
	if !strings.Contains(string(out), "sow version") {
		t.Errorf("expected version output, got: %s", out)
	}
}

func TestBarSubcommandHelp(t *testing.T) {
	bin := buildBinary(t)
	out, _ := exec.Command(bin, "bar", "--help").CombinedOutput()
	if !strings.Contains(string(out), "Vertical bar chart") {
		t.Errorf("expected bar help, got: %s", out)
	}
}

func TestNoInputError(t *testing.T) {
	bin := buildBinary(t)
	cmd := exec.Command(bin, "bar")
	cmd.Stdin = strings.NewReader("") // empty stdin, no file
	out, err := cmd.CombinedOutput()
	if err == nil {
		t.Error("expected error with no input")
	}
	output := string(out)
	if !strings.Contains(output, "no input") && !strings.Contains(output, "no data") {
		t.Errorf("expected input error message, got: %s", output)
	}
}

func TestPingSubcommandNoHost(t *testing.T) {
	bin := buildBinary(t)
	out, err := exec.Command(bin, "ping").CombinedOutput()
	if err == nil {
		t.Error("expected error when ping has no host")
	}
	if !strings.Contains(string(out), "requires a host") {
		t.Errorf("expected host error, got: %s", out)
	}
}

func TestFlagsExist(t *testing.T) {
	bin := buildBinary(t)
	out, _ := exec.Command(bin, "--help").CombinedOutput()
	output := string(out)
	for _, flag := range []string{"--title", "--color", "--live", "--ymin", "--ymax", "--pie-style", "--overlay", "--metric"} {
		if !strings.Contains(output, flag) {
			t.Errorf("missing flag in help: %s", flag)
		}
	}
}
