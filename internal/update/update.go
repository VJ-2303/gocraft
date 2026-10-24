// Package update checks for newer gocraft releases and installs them.
//
// Checks are cached for a day so normal runs stay offline and fast. The
// upgrade path never runs on its own: it is only triggered by an explicit
// `gocraft upgrade` command.
package update

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const (
	// Repo is the GitHub repository that publishes gocraft releases.
	Repo = "VJ-2303/gocraft"

	// cacheTTL is how long a version check result stays valid.
	cacheTTL = 24 * time.Hour

	// checkTimeout bounds the network call so startup never hangs.
	checkTimeout = 2 * time.Second
)

var (
	// apiBase and downloadBase are variables so tests can point them at a
	// local server.
	apiBase      = "https://api.github.com"
	downloadBase = "https://github.com"

	// userCacheDir and executable are variables for the same reason.
	userCacheDir = os.UserCacheDir
	executable   = os.Executable
)

// cache is the on-disk result of the last version check.
type cache struct {
	CheckedAt time.Time `json:"checked_at"`
	Latest    string    `json:"latest"`
}

// Notice prints a one-line upgrade hint when a newer release is known.
// It refreshes the cache at most once per day, and never fails loudly:
// a missing network or a GitHub outage is silently ignored.
func Notice(w io.Writer, current string) {
	if current == "" || current == "dev" || disabled() {
		return
	}

	c, have := readCache()
	if !have || time.Since(c.CheckedAt) > cacheTTL {
		ctx, cancel := context.WithTimeout(context.Background(), checkTimeout)
		defer cancel()

		if latest, err := Latest(ctx); err == nil {
			c = cache{CheckedAt: time.Now(), Latest: latest}
			have = true
			writeCache(c)
		}
	}

	if have && IsNewer(c.Latest, current) {
		fmt.Fprintf(w, "\nA new version of gocraft is available: %s (you have %s).\nUpdate with: gocraft upgrade\n\n", c.Latest, current)
	}
}

// disabled reports whether update checks are switched off for this run.
func disabled() bool {
	if os.Getenv("GOCRAFT_NO_UPDATE_CHECK") != "" {
		return true
	}
	return os.Getenv("CI") != ""
}

func cachePath() (string, error) {
	dir, err := userCacheDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "gocraft", "update-check.json"), nil
}

func readCache() (cache, bool) {
	path, err := cachePath()
	if err != nil {
		return cache{}, false
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return cache{}, false
	}
	var c cache
	if err := json.Unmarshal(data, &c); err != nil || c.Latest == "" {
		return cache{}, false
	}
	return c, true
}

func writeCache(c cache) {
	path, err := cachePath()
	if err != nil {
		return
	}
	data, err := json.Marshal(c)
	if err != nil {
		return
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return
	}
	_ = os.WriteFile(path, data, 0o644)
}

// IsNewer reports whether release tag a is newer than version b.
// Both may carry a leading "v"; anything unparseable is treated as older.
func IsNewer(a, b string) bool {
	av, ok := parseVersion(a)
	if !ok {
		return false
	}
	bv, ok := parseVersion(b)
	if !ok {
		return false
	}
	return av.compare(bv) > 0
}

type version struct {
	nums [3]int
	pre  string
}

func parseVersion(s string) (version, bool) {
	s = strings.TrimPrefix(strings.TrimSpace(s), "v")
	if i := strings.IndexByte(s, '+'); i >= 0 {
		s = s[:i]
	}
	var pre string
	if i := strings.IndexByte(s, '-'); i >= 0 {
		s, pre = s[:i], s[i+1:]
	}
	return parseVersionParts(s, pre)
}

func parseVersionParts(core, pre string) (version, bool) {
	parts := strings.Split(core, ".")
	if len(parts) < 2 || len(parts) > 3 {
		return version{}, false
	}
	var v version
	for i, p := range parts {
		n, err := strconv.Atoi(p)
		if err != nil || n < 0 {
			return version{}, false
		}
		v.nums[i] = n
	}
	v.pre = pre
	return v, true
}

func (v version) compare(o version) int {
	for i := range v.nums {
		switch {
		case v.nums[i] < o.nums[i]:
			return -1
		case v.nums[i] > o.nums[i]:
			return 1
		}
	}
	// A release outranks its prereleases: 1.0.0 > 1.0.0-rc.1.
	switch {
	case v.pre == o.pre:
		return 0
	case v.pre == "":
		return 1
	case o.pre == "":
		return -1
	}
	return strings.Compare(v.pre, o.pre)
}
