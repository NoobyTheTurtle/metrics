package util

import (
	"bytes"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPrintBuildInfo(t *testing.T) {
	t.Run("with all values", func(t *testing.T) {
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		PrintBuildInfo("1.0.0", "2024-01-01", "abc123")

		w.Close()
		os.Stdout = oldStdout

		var buf bytes.Buffer
		buf.ReadFrom(r)
		output := buf.String()

		assert.Contains(t, output, "Build version: 1.0.0")
		assert.Contains(t, output, "Build date: 2024-01-01")
		assert.Contains(t, output, "Build commit: abc123")
	})

	t.Run("with empty values", func(t *testing.T) {
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		PrintBuildInfo("", "", "")

		w.Close()
		os.Stdout = oldStdout

		var buf bytes.Buffer
		buf.ReadFrom(r)
		output := buf.String()

		assert.Contains(t, output, "Build version: N/A")
		assert.Contains(t, output, "Build date: N/A")
		assert.Contains(t, output, "Build commit: N/A")
	})

	t.Run("with partial values", func(t *testing.T) {
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		PrintBuildInfo("1.0.0", "", "abc123")

		w.Close()
		os.Stdout = oldStdout

		var buf bytes.Buffer
		buf.ReadFrom(r)
		output := buf.String()

		assert.Contains(t, output, "Build version: 1.0.0")
		assert.Contains(t, output, "Build date: N/A")
		assert.Contains(t, output, "Build commit: abc123")
	})
}

func TestBuildInfo_Structure(t *testing.T) {
	info := BuildInfo{
		Version: "1.0.0",
		Date:    "2024-01-01",
		Commit:  "abc123",
	}

	assert.Equal(t, "1.0.0", info.Version)
	assert.Equal(t, "2024-01-01", info.Date)
	assert.Equal(t, "abc123", info.Commit)
}
