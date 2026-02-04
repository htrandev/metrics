package info

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestPrintBuildInfo(t *testing.T) {
	naValue := getValue("")

	testCases := []struct {
		name    string
		version string
		date    string
		commit  string
	}{
		{
			name:    "all",
			version: "v0.0.1",
			date:    time.Now().String(),
			commit:  "test commit",
		},
		{
			name:    "n/a version",
			version: naValue,
			date:    time.Now().String(),
			commit:  "test commit",
		},
		{
			name:    "n/a date",
			version: "v0.0.1",
			date:    naValue,
			commit:  "test commit",
		},
		{
			name:    "n/a commit",
			version: "v0.0.1",
			date:    time.Now().String(),
			commit:  naValue,
		},
	}

	for _, tc := range testCases {
		buildVersion = tc.version
		buildDate = tc.date
		buildCommit = tc.commit

		r, w, err := os.Pipe()
		require.NoError(t, err)

		originalStdout := os.Stdout
		defer func() {
			os.Stdout = originalStdout
			r.Close()
		}()

		os.Stdout = w

		PrintBuildInfo()
		w.Close()

		var buf bytes.Buffer
		_, err = io.Copy(&buf, r)
		require.NoError(t, err)

		expected := buildExpectedOutput(tc.version, tc.date, tc.commit)
		require.Equal(t, expected, buf.String())
	}
}

func buildExpectedOutput(version, date, commit string) string {
	expected := strings.Builder{}
	expected.Grow(len(version) + len(date) + len(commit) + 15 + 12 + 14 + 3)
	expected.WriteString("Build version: ") // 15
	expected.WriteString(buildVersion)
	expected.WriteByte('\n') // 1

	expected.WriteString("Build date: ") // 12
	expected.WriteString(buildDate)
	expected.WriteByte('\n') // 1

	expected.WriteString("Build commit: ") // 14
	expected.WriteString(buildCommit)
	expected.WriteByte('\n') // 1

	return expected.String()
}
