package update

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// maxDownloadSize caps archive and checksum downloads at 100 MiB.
const maxDownloadSize = 100 << 20

// Run executes `gocraft upgrade`. It returns a process exit code.
func Run(args []string, stdout, stderr io.Writer, current string) int {
	fs := flag.NewFlagSet("gocraft upgrade", flag.ContinueOnError)
	fs.SetOutput(stderr)

	var (
		check   = fs.Bool("check", false, "report whether a newer release exists, without installing it")
		target  = fs.String("version", "", "release tag to install (default: latest)")
		timeout = fs.Duration("timeout", 2*time.Minute, "overall time limit for the download and install")
	)
	fs.Usage = func() {
		fmt.Fprintf(stderr, "Usage: gocraft upgrade [flags]\n\nFlags:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}
		return 2
	}

	if current == "" || current == "dev" {
		fmt.Fprintln(stderr, "gocraft was built from source, so it cannot update itself.")
		fmt.Fprintln(stderr, "Update it with:")
		fmt.Fprintln(stderr, "  go install github.com/VJ-2303/gocraft/cmd/gocraft@latest")
		if *check {
			return 0
		}
		return 1
	}

	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()

	tag := strings.TrimSpace(*target)
	if tag == "" {
		latest, err := Latest(ctx)
		if err != nil {
			fmt.Fprintf(stderr, "error: checking for updates: %v\n", err)
			return 1
		}
		tag = latest
	}
	if !strings.HasPrefix(tag, "v") {
		tag = "v" + tag
	}

	switch {
	case !IsNewer(tag, current):
		fmt.Fprintf(stdout, "gocraft %s is up to date.\n", current)
		return 0
	case *check:
		fmt.Fprintf(stdout, "gocraft %s is available (you have %s).\nRun 'gocraft upgrade' to install it.\n", tag, current)
		return 0
	}

	fmt.Fprintf(stdout, "Downloading gocraft %s for %s/%s...\n", tag, runtime.GOOS, runtime.GOARCH)

	archive := archiveName(tag, runtime.GOOS, runtime.GOARCH)
	url := downloadBase + "/" + Repo + "/releases/download/" + tag + "/" + archive

	data, err := fetch(ctx, url)
	if err != nil {
		fmt.Fprintf(stderr, "error: downloading %s: %v\n", url, err)
		return 1
	}
	if err := verifyChecksum(ctx, tag, archive, data); err != nil {
		fmt.Fprintf(stderr, "error: %v\n", err)
		return 1
	}

	bin, err := extractBinary(archive, data)
	if err != nil {
		fmt.Fprintf(stderr, "error: extracting %s: %v\n", archive, err)
		return 1
	}

	path, err := install(bin)
	if err != nil {
		fmt.Fprintf(stderr, "error: installing update: %v\n", err)
		return 1
	}

	fmt.Fprintf(stdout, "Updated gocraft: %s -> %s (%s)\n", current, tag, path)
	return 0
}

// archiveName reproduces the goreleaser archive naming scheme.
func archiveName(tag, goos, goarch string) string {
	name := fmt.Sprintf("gocraft_%s_%s_%s", strings.TrimPrefix(tag, "v"), goos, goarch)
	if goos == "windows" {
		return name + ".zip"
	}
	return name + ".tar.gz"
}

// fetch downloads a URL into memory, refusing responses above maxDownloadSize.
func fetch(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "gocraft")

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected HTTP status %s", resp.Status)
	}

	data, err := io.ReadAll(io.LimitReader(resp.Body, maxDownloadSize+1))
	if err != nil {
		return nil, err
	}
	if len(data) > maxDownloadSize {
		return nil, fmt.Errorf("response exceeds %d bytes", maxDownloadSize)
	}
	return data, nil
}

// verifyChecksum compares data against the release's checksums.txt entry.
func verifyChecksum(ctx context.Context, tag, archive string, data []byte) error {
	url := downloadBase + "/" + Repo + "/releases/download/" + tag + "/checksums.txt"
	checksums, err := fetch(ctx, url)
	if err != nil {
		return fmt.Errorf("downloading checksums: %w", err)
	}

	expected := ""
	for _, line := range strings.Split(string(checksums), "\n") {
		fields := strings.Fields(line)
		if len(fields) == 2 && strings.TrimPrefix(fields[1], "*") == archive {
			expected = strings.ToLower(fields[0])
			break
		}
	}
	if expected == "" {
		return fmt.Errorf("checksums.txt has no entry for %s", archive)
	}

	sum := sha256.Sum256(data)
	if actual := hex.EncodeToString(sum[:]); actual != expected {
		return fmt.Errorf("checksum mismatch for %s: expected %s, got %s", archive, expected, actual)
	}
	return nil
}

// extractBinary pulls the gocraft executable out of a release archive.
func extractBinary(archive string, data []byte) ([]byte, error) {
	name := "gocraft"
	if runtime.GOOS == "windows" {
		name = "gocraft.exe"
	}
	if strings.HasSuffix(archive, ".zip") {
		return binaryFromZip(name, data)
	}
	return binaryFromTarGz(name, data)
}

func binaryFromTarGz(name string, data []byte) ([]byte, error) {
	gz, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer gz.Close()

	tr := tar.NewReader(gz)
	for {
		hdr, err := tr.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, err
		}
		if hdr.Typeflag == tar.TypeReg && filepath.Base(hdr.Name) == name {
			return io.ReadAll(io.LimitReader(tr, maxDownloadSize))
		}
	}
	return nil, fmt.Errorf("%s not found in archive", name)
}

func binaryFromZip(name string, data []byte) ([]byte, error) {
	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, err
	}
	for _, f := range zr.File {
		if filepath.Base(f.Name) != name {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			return nil, err
		}
		defer rc.Close()
		return io.ReadAll(io.LimitReader(rc, maxDownloadSize))
	}
	return nil, fmt.Errorf("%s not found in archive", name)
}

// install atomically replaces the running executable with data.
func install(data []byte) (string, error) {
	exe, err := executable()
	if err != nil {
		return "", err
	}
	if resolved, err := filepath.EvalSymlinks(exe); err == nil {
		exe = resolved
	}
	if err := installAt(data, exe); err != nil {
		return "", err
	}
	return exe, nil
}

// installAt atomically replaces the executable at exe with data.
func installAt(data []byte, exe string) error {
	dir := filepath.Dir(exe)

	if err := checkWritable(dir); err != nil {
		return fmt.Errorf("%s is not writable (%v); reinstall with the install script or move gocraft to a user-writable directory", dir, err)
	}

	tmp, err := os.CreateTemp(dir, ".gocraft-new-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)

	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Chmod(0o755); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}

	// On Windows a running executable cannot be overwritten, but it can be
	// renamed. Move it aside first, then restore it if the swap fails.
	if runtime.GOOS == "windows" {
		old := exe + ".old"
		_ = os.Remove(old)
		if err := os.Rename(exe, old); err != nil {
			return err
		}
		if err := os.Rename(tmpName, exe); err != nil {
			_ = os.Rename(old, exe)
			return err
		}
		_ = os.Remove(old)
		return nil
	}

	return os.Rename(tmpName, exe)
}

// checkWritable confirms a directory accepts new files.
func checkWritable(dir string) error {
	f, err := os.CreateTemp(dir, ".gocraft-write-*")
	if err != nil {
		return err
	}
	name := f.Name()
	f.Close()
	return os.Remove(name)
}
