package embedder

import "testing"

func TestDetectType(t *testing.T) {
	tests := []struct {
		path        string
		wantMIME    string
		wantContent string
		wantErr     bool
	}{
		{"photo.png", "image/png", "image", false},
		{"photo.PNG", "image/png", "image", false}, // case insensitive
		{"photo.jpg", "image/jpeg", "image", false},
		{"photo.jpeg", "image/jpeg", "image", false},
		{"clip.mp4", "video/mp4", "video", false},
		{"clip.mov", "video/quicktime", "video", false},
		{"song.mp3", "audio/mpeg", "audio", false},
		{"song.wav", "audio/wav", "audio", false},
		{"song.flac", "audio/flac", "audio", false},
		{"doc.pdf", "application/pdf", "pdf", false},
		{"file.xyz", "", "", true},         // unsupported
		{"noext", "", "", true},            // no extension
		{"path/to/file.webp", "image/webp", "image", false},
	}

	for _, tc := range tests {
		mime, ct, err := DetectType(tc.path)
		if tc.wantErr {
			if err == nil {
				t.Errorf("DetectType(%q): expected error, got nil", tc.path)
			}
			continue
		}
		if err != nil {
			t.Errorf("DetectType(%q): unexpected error: %v", tc.path, err)
			continue
		}
		if mime != tc.wantMIME {
			t.Errorf("DetectType(%q): MIME = %q, want %q", tc.path, mime, tc.wantMIME)
		}
		if ct != tc.wantContent {
			t.Errorf("DetectType(%q): ContentType = %q, want %q", tc.path, ct, tc.wantContent)
		}
	}
}

func TestIsSupportedExt(t *testing.T) {
	if !IsSupportedExt("photo.png") {
		t.Error("expected .png to be supported")
	}
	if !IsSupportedExt("PHOTO.PNG") {
		t.Error("expected .PNG (uppercase) to be supported")
	}
	if IsSupportedExt("file.xyz") {
		t.Error("expected .xyz to be unsupported")
	}
}
