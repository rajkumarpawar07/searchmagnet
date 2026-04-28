package format

import "testing"

func TestHumanBytes(t *testing.T) {
	tests := []struct {
		input int64
		want  string
	}{
		{0, "0 B"},
		{512, "512 B"},
		{1024, "1.0 KB"},
		{1536, "1.5 KB"},
		{1048576, "1.0 MB"},
		{1572864, "1.5 MB"},
		{16542531, "15.8 MB"},
		{1073741824, "1.0 GB"},
	}
	for _, tc := range tests {
		got := HumanBytes(tc.input)
		if got != tc.want {
			t.Errorf("HumanBytes(%d) = %q, want %q", tc.input, got, tc.want)
		}
	}
}

func TestTruncate(t *testing.T) {
	tests := []struct {
		input string
		n     int
		want  string
	}{
		{"hello", 10, "hello"},
		{"hello world", 5, "hell…"},
		{"ab", 2, "ab"},
		{"abc", 2, "a…"},
	}
	for _, tc := range tests {
		got := Truncate(tc.input, tc.n)
		if got != tc.want {
			t.Errorf("Truncate(%q, %d) = %q, want %q", tc.input, tc.n, got, tc.want)
		}
	}
}

func TestTypeIcon(t *testing.T) {
	if TypeIcon("image") == "" {
		t.Error("image should have an icon")
	}
	if TypeIcon("video") == "" {
		t.Error("video should have an icon")
	}
	if TypeIcon("unknown") == "" {
		t.Error("unknown type should still return a fallback icon")
	}
}
