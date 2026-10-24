package update

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestIsNewer(t *testing.T) {
	tests := []struct {
		latest, current string
		want            bool
	}{
		{"v0.2.0", "v0.1.0", true},
		{"0.2.0", "v0.1.0", true},
		{"v0.1.1", "v0.1.0", true},
		{"v1.0.0", "v0.99.99", true},
		{"v0.1.0", "v0.1.0", false},
		{"v0.1.0", "v0.2.0", false},
		{"v0.2.0-rc.1", "v0.2.0", false},
		{"v0.2.0", "v0.2.0-rc.1", true},
		{"v0.2.0+build.5", "v0.2.0", false},
		{"dev", "v0.1.0", false},
		{"", "v0.1.0", false},
		{"v0.2", "v0.1.9", true},
	}
	for _, tt := range tests {
		if got := IsNewer(tt.latest, tt.current); got != tt.want {
			t.Errorf("IsNewer(%q, %q) = %v, want %v", tt.latest, tt.current, got, tt.want)
		}
	}
}

func TestNoticePrintsFromFreshCache(t *testing.T) {
	t.Setenv("CI", "")
	t.Setenv("GOCRAFT_NO_UPDATE_CHECK", "")
	useTempCache(t)
	writeCache(cache{CheckedAt: time.Now(), Latest: "v0.2.0"})

	var buf bytes.Buffer
	Notice(&buf, "v0.1.0")

	out := buf.String()
	if !strings.Contains(out, "v0.2.0") || !strings.Contains(out, "gocraft upgrade") {
		t.Fatalf("notice = %q, want version and upgrade hint", out)
	}
}

func TestNoticeSilentWhenCurrentAndDisabled(t *testing.T) {
	useTempCache(t)
	writeCache(cache{CheckedAt: time.Now(), Latest: "v0.2.0"})

	var buf bytes.Buffer
	Notice(&buf, "v0.2.0")
	if buf.Len() != 0 {
		t.Fatalf("notice = %q, want empty when up to date", buf.String())
	}

	t.Setenv("GOCRAFT_NO_UPDATE_CHECK", "1")
	Notice(&buf, "v0.1.0")
	if buf.Len() != 0 {
		t.Fatalf("notice = %q, want empty when disabled", buf.String())
	}
}

func TestNoticeUsesStaleCacheWhenRefreshFails(t *testing.T) {
	t.Setenv("CI", "")
	t.Setenv("GOCRAFT_NO_UPDATE_CHECK", "")
	useTempCache(t)
	writeCache(cache{CheckedAt: time.Now().Add(-48 * time.Hour), Latest: "v0.9.0"})

	oldBase := apiBase
	oldDownload := downloadBase
	apiBase = "http://127.0.0.1:1"
	downloadBase = "http://127.0.0.1:1"
	t.Cleanup(func() { apiBase, downloadBase = oldBase, oldDownload })

	var buf bytes.Buffer
	Notice(&buf, "v0.1.0")
	if !strings.Contains(buf.String(), "v0.9.0") {
		t.Fatalf("notice = %q, want stale cached version", buf.String())
	}
}

func TestLatestFromAPI(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/repos/"+Repo+"/releases/latest" {
			http.NotFound(w, r)
			return
		}
		fmt.Fprint(w, `{"tag_name":"v0.3.0"}`)
	}))
	defer srv.Close()
	useBase(t, srv.URL, "http://127.0.0.1:1")

	tag, err := Latest(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if tag != "v0.3.0" {
		t.Fatalf("tag = %q, want v0.3.0", tag)
	}
}

func TestLatestFallsBackToRedirect(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/repos/" + Repo + "/releases/latest":
			http.Error(w, "rate limited", http.StatusForbidden)
		case "/" + Repo + "/releases/latest":
			http.Redirect(w, r, "/"+Repo+"/releases/tag/v0.4.0", http.StatusFound)
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()
	useBase(t, srv.URL, srv.URL)

	tag, err := Latest(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if tag != "v0.4.0" {
		t.Fatalf("tag = %q, want v0.4.0", tag)
	}
}

func TestArchiveName(t *testing.T) {
	tests := []struct {
		tag, goos, goarch, want string
	}{
		{"v0.2.0", "linux", "amd64", "gocraft_0.2.0_linux_amd64.tar.gz"},
		{"v0.2.0", "darwin", "arm64", "gocraft_0.2.0_darwin_arm64.tar.gz"},
		{"v0.2.0", "windows", "amd64", "gocraft_0.2.0_windows_amd64.zip"},
	}
	for _, tt := range tests {
		if got := archiveName(tt.tag, tt.goos, tt.goarch); got != tt.want {
			t.Errorf("archiveName(%q, %q, %q) = %q, want %q", tt.tag, tt.goos, tt.goarch, got, tt.want)
		}
	}
}

func TestVerifyChecksum(t *testing.T) {
	data := []byte("release binary")
	sum := sha256.Sum256(data)
	archive := "gocraft_0.2.0_linux_amd64.tar.gz"

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "%s  %s\n", hex.EncodeToString(sum[:]), archive)
	}))
	defer srv.Close()
	useBase(t, srv.URL, srv.URL)

	if err := verifyChecksum(context.Background(), "v0.2.0", archive, data); err != nil {
		t.Fatalf("verifyChecksum: %v", err)
	}
	if err := verifyChecksum(context.Background(), "v0.2.0", archive, []byte("tampered")); err == nil {
		t.Fatal("verifyChecksum accepted mismatched data")
	}
}

func TestExtractBinaryTarGz(t *testing.T) {
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)
	payload := []byte("#!/bin/sh\necho gocraft\n")
	if err := tw.WriteHeader(&tar.Header{Name: "gocraft", Mode: 0o755, Size: int64(len(payload))}); err != nil {
		t.Fatal(err)
	}
	if _, err := tw.Write(payload); err != nil {
		t.Fatal(err)
	}
	tw.Close()
	gz.Close()

	got, err := binaryFromTarGz("gocraft", buf.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, payload) {
		t.Fatalf("payload = %q, want %q", got, payload)
	}
}

func TestExtractBinaryZip(t *testing.T) {
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	w, err := zw.Create("gocraft.exe")
	if err != nil {
		t.Fatal(err)
	}
	payload := []byte("MZ fake binary")
	if _, err := w.Write(payload); err != nil {
		t.Fatal(err)
	}
	zw.Close()

	got, err := binaryFromZip("gocraft.exe", buf.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, payload) {
		t.Fatalf("payload = %q, want %q", got, payload)
	}
}

func TestRunUpgradesBinary(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("install path differs on Windows")
	}

	newBinary := []byte("#!/bin/sh\necho gocraft v0.2.0\n")
	archive := "gocraft_0.2.0_" + runtime.GOOS + "_" + runtime.GOARCH + ".tar.gz"

	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)
	if err := tw.WriteHeader(&tar.Header{Name: "gocraft", Mode: 0o755, Size: int64(len(newBinary))}); err != nil {
		t.Fatal(err)
	}
	if _, err := tw.Write(newBinary); err != nil {
		t.Fatal(err)
	}
	tw.Close()
	gz.Close()

	sum := sha256.Sum256(buf.Bytes())
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/" + Repo + "/releases/download/v0.2.0/" + archive:
			w.Write(buf.Bytes())
		case "/" + Repo + "/releases/download/v0.2.0/checksums.txt":
			fmt.Fprintf(w, "%s  %s\n", hex.EncodeToString(sum[:]), archive)
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()
	useBase(t, srv.URL, srv.URL)

	exe := filepath.Join(t.TempDir(), "gocraft")
	if err := os.WriteFile(exe, []byte("old"), 0o755); err != nil {
		t.Fatal(err)
	}
	oldExe := executable
	executable = func() (string, error) { return exe, nil }
	t.Cleanup(func() { executable = oldExe })

	var stdout, stderr bytes.Buffer
	if code := Run([]string{"-version", "v0.2.0"}, &stdout, &stderr, "v0.1.0"); code != 0 {
		t.Fatalf("Run = %d, stderr = %s", code, stderr.String())
	}

	got, err := os.ReadFile(exe)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, newBinary) {
		t.Fatalf("installed binary = %q, want %q", got, newBinary)
	}
	if info, err := os.Stat(exe); err != nil || info.Mode().Perm()&0o100 == 0 {
		t.Fatalf("installed binary is not executable: %v", err)
	}
	if !strings.Contains(stdout.String(), "Updated gocraft: v0.1.0 -> v0.2.0") {
		t.Fatalf("stdout = %q", stdout.String())
	}
}

func useTempCache(t *testing.T) {
	t.Helper()
	dir := t.TempDir()
	old := userCacheDir
	userCacheDir = func() (string, error) { return dir, nil }
	t.Cleanup(func() { userCacheDir = old })
}

func useBase(t *testing.T, api, download string) {
	t.Helper()
	oldAPI, oldDownload := apiBase, downloadBase
	apiBase, downloadBase = api, download
	t.Cleanup(func() { apiBase, downloadBase = oldAPI, oldDownload })
}
