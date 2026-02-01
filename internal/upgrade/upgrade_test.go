package upgrade

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/inconshreveable/go-update"
)

func TestGetLatestVersion(t *testing.T) {
	tests := []struct {
		name         string
		server       *httptest.Server
		want         string
		wantErr      bool
		expectErrStr string
	}{
		{
			name: "successful response",
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				if _, err := w.Write([]byte(`{"tag_name": "v1.2.3"}`)); err != nil {
					t.Fatal(err)
				}
			})),
			want:    "v1.2.3",
			wantErr: false,
		},
		{
			name: "server error",
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
			})),
			wantErr:      true,
			expectErrStr: "unexpected status code: 500",
		},
		{
			name: "invalid json",
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				if _, err := w.Write([]byte(`{"tag_name": "v1.2.3"`)); err != nil {
					t.Fatal(err)
				}
			})),
			wantErr:      true,
			expectErrStr: "unexpected EOF",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer tt.server.Close()
			originalURL := githubAPIURL
			githubAPIURL = tt.server.URL
			defer func() { githubAPIURL = originalURL }()

			got, err := getLatestVersion()

			if (err != nil) != tt.wantErr {
				t.Errorf("getLatestVersion() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && err.Error() != tt.expectErrStr {
				t.Errorf("getLatestVersion() error = %v, wantErrStr %v", err.Error(), tt.expectErrStr)
			}

			if got != tt.want {
				t.Errorf("getLatestVersion() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDoUpgrade(t *testing.T) {
	// Create a mock tar.gz archive with a binary
	var archiveBuffer bytes.Buffer
	gzWriter := gzip.NewWriter(&archiveBuffer)
	tarWriter := tar.NewWriter(gzWriter)

	// Add a fake binary to the archive
	binaryContent := []byte("new binary content")
	header := &tar.Header{
		Name: "jellywolproxy-linux-amd64",
		Mode: 0755,
		Size: int64(len(binaryContent)),
	}
	if err := tarWriter.WriteHeader(header); err != nil {
		t.Fatal(err)
	}
	if _, err := tarWriter.Write(binaryContent); err != nil {
		t.Fatal(err)
	}
	if err := tarWriter.Close(); err != nil {
		t.Fatal(err)
	}
	if err := gzWriter.Close(); err != nil {
		t.Fatal(err)
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write(archiveBuffer.Bytes()); err != nil {
			t.Fatal(err)
		}
	}))
	defer server.Close()

	originalGetDownloadURL := getDownloadURL
	getDownloadURL = func(version string) (string, error) {
		return server.URL, nil
	}
	defer func() { getDownloadURL = originalGetDownloadURL }()

	originalGetconfigDownloadURL := getconfigDownloadURL
	getconfigDownloadURL = func(version string) (string, error) {
		return server.URL, nil
	}
	defer func() { getconfigDownloadURL = originalGetconfigDownloadURL }()

	originalApplyUpdate := applyUpdate
	applyUpdate = func(body io.Reader, opts update.Options) error {
		return nil
	}
	defer func() { applyUpdate = originalApplyUpdate }()

	originalCompareConfigs := compareConfigs
	compareConfigs = func(newConfig io.Reader) error {
		return nil
	}
	defer func() { compareConfigs = originalCompareConfigs }()

	err := doUpgrade("v1.2.3")
	if err != nil {
		t.Errorf("doUpgrade() error = %v, wantErr %v", err, false)
	}
}

func TestCompareConfigs(t *testing.T) {
	tests := []struct {
		name       string
		newConfig  string
		oldConfig  string
		wantOutput string
	}{
		{
			name:      "same config",
			newConfig: "{\"foo\": \"bar\"}",
			oldConfig: "{\"foo\": \"bar\"}",
		},
		{
			name:       "different config",
			newConfig:  "{\"foo\": \"baz\"}",
			oldConfig:  "{\"foo\": \"bar\"}",
			wantOutput: "Configuration has been updated. Please check your config.json file.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpfile, err := os.CreateTemp("", "config.json.example")
			if err != nil {
				t.Fatal(err)
			}
			defer func() {
				if err := os.Remove(tmpfile.Name()); err != nil {
					t.Fatal(err)
				}
			}()

			if _, err := tmpfile.Write([]byte(tt.oldConfig)); err != nil {
				t.Fatal(err)
			}
			if err := tmpfile.Close(); err != nil {
				t.Fatal(err)
			}

			originalReadFile := readFile
			readFile = func(filename string) ([]byte, error) {
				return os.ReadFile(tmpfile.Name())
			}
			defer func() { readFile = originalReadFile }()

			r := strings.NewReader(tt.newConfig)

			// Capture stdout
			oldStdout := os.Stdout
			defer func() { os.Stdout = oldStdout }()
			read, write, _ := os.Pipe()
			os.Stdout = write

			err = compareConfigs(r)
			if err != nil {
				t.Errorf("compareConfigs() error = %v, wantErr %v", err, false)
			}

			if err := write.Close(); err != nil {
				t.Fatal(err)
			}
			out, _ := io.ReadAll(read)

			if !strings.Contains(string(out), tt.wantOutput) {
				t.Errorf("compareConfigs() output = %q, want %q", string(out), tt.wantOutput)
			}
		})
	}
}
