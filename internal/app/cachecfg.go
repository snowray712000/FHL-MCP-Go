package app

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/snowray712000/FHL-MCP-Go/internal/cache"
)

// Environment variables that mirror the cache flags (flag > environment > default).
const (
	EnvDataDir     = "FHL_DATA_DIR"
	EnvCacheMaxMB  = "FHL_CACHE_MAX_MB"
	defaultCacheMB = int(cache.DefaultMaxBytes >> 20)
)

// ResolveCacheDir returns the cache directory: <dataDir>/cache when a data directory is given
// (flag, else FHL_DATA_DIR), otherwise <os.UserCacheDir>/fhl-mcp/cache. It never depends on the
// working directory unless the user passes a relative --data-dir.
func ResolveCacheDir(dataDir string) (string, error) {
	abs, err := ResolveDataDir(dataDir)
	if err != nil {
		return "", err
	}
	return filepath.Join(abs, "cache"), nil
}

// ResolveCacheMaxBytes turns the --cache-max-mb flag (-1 = not given) or FHL_CACHE_MAX_MB into a byte cap.
// 0 disables the cache.
func ResolveCacheMaxBytes(flagMB int) (int64, error) {
	mb := flagMB
	if mb < 0 {
		mb = defaultCacheMB
		if v := os.Getenv(EnvCacheMaxMB); v != "" {
			n, err := strconv.Atoi(v)
			if err != nil || n < 0 {
				return 0, fmt.Errorf("%s=%q is not a non-negative number of megabytes", EnvCacheMaxMB, v)
			}
			mb = n
		}
	}
	return int64(mb) << 20, nil
}
