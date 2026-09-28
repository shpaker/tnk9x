//go:build !js

package raw

import "testing"

func TestStorageRepository_RoundTrip(t *testing.T) {
	storage := NewStorageRepository(t.TempDir())

	data, err := storage.Load("progress")
	if err != nil || data != nil {
		t.Fatalf("missing key: %q %v, want nil nil", data, err)
	}

	if err := storage.Save("progress", []byte("saved")); err != nil {
		t.Fatalf("save: %v", err)
	}
	data, err = storage.Load("progress")
	if err != nil || string(data) != "saved" {
		t.Errorf("load: %q %v", data, err)
	}
}
