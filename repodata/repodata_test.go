package repodata

import (
	"encoding/json"
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/assert"
)

func TestEncodeJSON(t *testing.T) {
	s := map[string]interface{}{
		"key<&>": "Hello 🐧",
	}
	rawBytes, err := EncodeJSON(s, "")
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}

	assert.Equal(t, []byte(`{"key<&>":"Hello 🐧"}`+"\n"), rawBytes)
}

func TestRepodataRecordJSON(t *testing.T) {
	r := RepodataRecord{
		Subdir:      "linux-arm64",
		Name:        "penguin",
		Version:     "1.2.3-beta",
		BuildNumber: 999,
		Build:       "build-gentoo",
		// Fn
		Sha256: "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
		Size:   12345,
		Md5:    "0123456789abcdef0123456789abcdef",
		Extra: map[string]interface{}{
			"extra1": "value1",
			"extra2": float64(2),
		},
	}

	rawBytes, err := r.MarshalJSON()
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}

	assert.True(t, utf8.Valid(rawBytes))

	s := RepodataRecord{}
	err = s.UnmarshalJSON(rawBytes)
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}

	assert.Equal(t, r, s)
}

func TestRepodataJSON(t *testing.T) {
	repodataBytes := []byte(`{
		"repodata_version": 1,
		"info": {"subdir": "win-arm64"},
		"packages": {
			"penguin-1.tar.bz2": {"name": "penguin", "version": "1"}
		},
		"packages.conda": {
			"penguin-2.conda": {"name": "penguin", "version": "2"}
		}
	}`)

	r := &Repodata{}
	err := json.Unmarshal(repodataBytes, r)
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}

	assert.Equal(t, 1, r.RepodataVersion)
	assert.Equal(t, "win-arm64", r.Info.Subdir)

	assert.Equal(t, 1, len(r.Packages))
	assert.Equal(t, "penguin", r.Packages["penguin-1.tar.bz2"].Name)
	assert.Equal(t, "1", r.Packages["penguin-1.tar.bz2"].Version)

	assert.Equal(t, 1, len(r.PackagesConda))
	assert.Equal(t, "penguin", r.PackagesConda["penguin-2.conda"].Name)
	assert.Equal(t, "2", r.PackagesConda["penguin-2.conda"].Version)
}
