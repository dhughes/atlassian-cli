package cmd

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSplitHeader(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantKey   string
		wantValue string
		wantOK    bool
	}{
		{"simple", "Accept: application/json", "Accept", "application/json", true},
		{"no space after colon", "Accept:application/json", "Accept", "application/json", true},
		{"extra whitespace", "  Accept  :  application/json  ", "Accept", "application/json", true},
		{"value contains colons", "X-Time: 12:34:56", "X-Time", "12:34:56", true},
		{"empty value", "X-Custom:", "X-Custom", "", true},
		{"no colon", "BareWord", "", "", false},
		{"empty string", "", "", "", false},
		{"empty key", ":value", "", "", false},
		{"whitespace-only key", "   :value", "", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key, value, ok := splitHeader(tt.input)
			if ok != tt.wantOK {
				t.Errorf("ok = %v, want %v", ok, tt.wantOK)
			}
			if key != tt.wantKey {
				t.Errorf("key = %q, want %q", key, tt.wantKey)
			}
			if value != tt.wantValue {
				t.Errorf("value = %q, want %q", value, tt.wantValue)
			}
		})
	}
}

func TestReadBody_Empty(t *testing.T) {
	body, hasBody, err := readBody("", strings.NewReader(""))
	if err != nil {
		t.Fatalf("readBody: %v", err)
	}
	if hasBody {
		t.Error("hasBody = true, want false for empty data")
	}
	if body != nil {
		t.Error("body should be nil for empty data")
	}
}

func TestReadBody_Literal(t *testing.T) {
	body, hasBody, err := readBody(`{"key":"value"}`, strings.NewReader(""))
	if err != nil {
		t.Fatalf("readBody: %v", err)
	}
	if !hasBody {
		t.Error("hasBody = false, want true for literal data")
	}
	got, _ := io.ReadAll(body)
	if string(got) != `{"key":"value"}` {
		t.Errorf("body = %q, want %q", string(got), `{"key":"value"}`)
	}
}

func TestReadBody_Stdin(t *testing.T) {
	stdin := strings.NewReader("from stdin")
	body, hasBody, err := readBody("@-", stdin)
	if err != nil {
		t.Fatalf("readBody: %v", err)
	}
	if !hasBody {
		t.Error("hasBody = false, want true for @-")
	}
	got, _ := io.ReadAll(body)
	if string(got) != "from stdin" {
		t.Errorf("body = %q, want %q", string(got), "from stdin")
	}
}

func TestReadBody_File(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "body.json")
	want := `{"from":"file"}`
	if err := os.WriteFile(path, []byte(want), 0600); err != nil {
		t.Fatalf("write tempfile: %v", err)
	}

	body, hasBody, err := readBody("@"+path, strings.NewReader(""))
	if err != nil {
		t.Fatalf("readBody: %v", err)
	}
	if !hasBody {
		t.Error("hasBody = false, want true for @file")
	}
	got, _ := io.ReadAll(body)
	if string(got) != want {
		t.Errorf("body = %q, want %q", string(got), want)
	}
}

func TestReadBody_FileMissing(t *testing.T) {
	_, _, err := readBody("@/does/not/exist/atl-test.json", strings.NewReader(""))
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
	if !strings.Contains(err.Error(), "could not read body file") {
		t.Errorf("error = %q, want it to contain 'could not read body file'", err.Error())
	}
}

func TestReadBody_LiteralStartingWithAtSymbol(t *testing.T) {
	// A literal value that happens to start with "@" is interpreted as a
	// file reference. This matches curl's behavior; document it as a known
	// edge case rather than try to disambiguate.
	_, _, err := readBody("@not-a-real-file", strings.NewReader(""))
	if err == nil {
		t.Fatal("expected error treating @-prefixed literal as missing file")
	}
}
