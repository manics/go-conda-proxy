package repodata

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func mockServer(t *testing.T, failOnError bool) *httptest.Server {
	count := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count++
		if r.URL.Path == "/channel-test/win-arm64/repodata.json" {
			w.WriteHeader(http.StatusOK)
			if _, err := w.Write([]byte(`{"info":{"subdir":"win-arm64"}}`)); err != nil {
				t.Fatalf("Unexpected error: %s", err)
			}
		} else if r.URL.Path == "/channel-test/noarch/repodata.json" {
			w.WriteHeader(http.StatusOK)
			if _, err := w.Write([]byte(`{"info":{"subdir":"noarch"}}`)); err != nil {
				t.Fatalf("Unexpected error: %s", err)
			}
		} else if r.URL.Path == "/count" {
			w.WriteHeader(http.StatusOK)
			if _, err := w.Write([]byte(fmt.Sprint(count))); err != nil {
				t.Fatalf("Unexpected error: %s", err)
			}
		} else {
			if failOnError {
				t.Errorf("Unexpected URL path %s", r.URL.Path)
			}
			w.WriteHeader(http.StatusNotFound)
			if _, err := w.Write([]byte("Not found " + r.URL.Path)); err != nil {
				t.Fatalf("Unexpected error: %s", err)
			}
		}
	}))
	return server
}

func TestWriteTempAndRename(t *testing.T) {
	tmpdir := t.TempDir()

	reader := bytes.NewReader([]byte("Hello, world!"))
	destination := filepath.Join(tmpdir, "test.txt")

	err := WriteTempAndRename(reader, destination)
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}

	contents, err := os.ReadFile(destination)
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
	assert.Equal(t, "Hello, world!", string(contents))
}

func TestUpdateDownload(t *testing.T) {
	server := mockServer(t, true)
	defer server.Close()

	url := server.URL + "/count"
	tmpdir := t.TempDir()
	destination := filepath.Join(tmpdir, "count")

	// Make two requests with a 0 maxAgeMinutes (should result in updated file),
	// and one with a 1440 maxAgeMinutes (should result in cached file)

	if err := UpdateDownload(url, destination, 0); err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}

	if contents, err := os.ReadFile(destination); err != nil {
		t.Fatalf("Unexpected error: %s", err)
	} else {
		assert.Equal(t, "1", string(contents))
	}

	if err := UpdateDownload(url, destination, 0); err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}

	if contents, err := os.ReadFile(destination); err != nil {
		t.Fatalf("Unexpected error: %s", err)
	} else {
		assert.Equal(t, "2", string(contents))
	}

	if err := UpdateDownload(url, destination, 1440); err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}

	if contents, err := os.ReadFile(destination); err != nil {
		t.Fatalf("Unexpected error: %s", err)
	} else {
		assert.Equal(t, "2", string(contents))
	}
}

func TestGetDestinationFilename(t *testing.T) {
	assert.Equal(t, filepath.Join("tmp", "foo", "linux-64", "repodata.json"), GetDestinationFilename("tmp", "foo", "linux-64"))
}

func TestUpdateChannelRepodata(t *testing.T) {
	server := mockServer(t, true)
	defer server.Close()

	tmpdir := t.TempDir()
	if err := UpdateChannelRepodata(server.URL, tmpdir, "channel-test", []string{"win-arm64", "noarch"}, 0); err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}

	for _, subdir := range []string{"win-arm64", "noarch"} {
		if content, err := os.ReadFile(filepath.Join(tmpdir, "channel-test", subdir, "repodata.json")); err != nil {
			t.Fatalf("Unexpected error: %s", err)
		} else {
			assert.Equal(t, `{"info":{"subdir":"`+subdir+`"}}`, string(content))
		}
	}
}

func TestUpdateFromConfig(t *testing.T) {
	server := mockServer(t, false)
	defer server.Close()

	tmpdir := t.TempDir()
	cfg := &CondaRepoConfig{
		CondaHost:           server.URL,
		OriginalRepodataDir: filepath.Join(tmpdir, "original"),
		Channels: map[string]condaChannelConfig{
			"channel-test": {
				Subdirs: []string{"win-arm64", "noarch"},
			},
			"nonexistent": {
				Subdirs: []string{"osx-64", "linux-64"},
			},
		},
		MaxAgeMinutes: 0,
	}
	forceUpdate := false

	err := UpdateFromConfig(cfg, forceUpdate)
	assert.Equal(t, fmt.Sprintf("404 Not Found %s/nonexistent/osx-64/repodata.json\n404 Not Found %s/nonexistent/linux-64/repodata.json", server.URL, server.URL), err.Error())

	for _, subdir := range []string{"win-arm64", "noarch"} {
		if content, err := os.ReadFile(filepath.Join(tmpdir, "original", "channel-test", subdir, "repodata.json")); err != nil {
			t.Fatalf("Unexpected error: %s", err)
		} else {
			assert.Equal(t, `{"info":{"subdir":"`+subdir+`"}}`, string(content))
		}
	}

	for _, subdir := range []string{"osx-64", "linux-64"} {
		assert.NoFileExists(t, filepath.Join(tmpdir, "original", "nonexistent", subdir, "repodata.json"))
	}
}
