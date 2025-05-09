package specialized

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"bitspark.dev/go-tree/pkg/core/typesys"
	"bitspark.dev/go-tree/pkg/run/execute"
)

// CacheEntry represents a cached function execution result
type CacheEntry struct {
	Result     interface{}
	Error      error
	Timestamp  time.Time
	Expiration time.Time
}

// CacheOptions defines how caching should be performed
type CacheOptions struct {
	TTL                 time.Duration // Time to live for cache entries
	MaxSize             int           // Maximum number of entries in the cache (0 = unlimited)
	DisableCacheOnError bool          // Whether to cache error results
}

// DefaultCacheOptions returns reasonable default cache options
func DefaultCacheOptions() *CacheOptions {
	return &CacheOptions{
		TTL:                 30 * time.Minute,
		MaxSize:             1000,
		DisableCacheOnError: false,
	}
}

// CachedFunctionRunner caches function execution results
type CachedFunctionRunner struct {
	*execute.FunctionRunner // Embed the base FunctionRunner
	Options                 *CacheOptions
	cache                   map[string]*CacheEntry
	hitCount                int
	missCount               int
	mutex                   sync.RWMutex
}

// NewCachedFunctionRunner creates a new cached function runner
func NewCachedFunctionRunner(base *execute.FunctionRunner) *CachedFunctionRunner {
	return &CachedFunctionRunner{
		FunctionRunner: base,
		Options:        DefaultCacheOptions(),
		cache:          make(map[string]*CacheEntry),
	}
}

// WithOptions sets the cache options
func (r *CachedFunctionRunner) WithOptions(options *CacheOptions) *CachedFunctionRunner {
	r.Options = options
	return r
}

// WithTTL sets the time to live for cache entries
func (r *CachedFunctionRunner) WithTTL(ttl time.Duration) *CachedFunctionRunner {
	r.Options.TTL = ttl
	return r
}

// WithMaxSize sets the maximum number of entries in the cache
func (r *CachedFunctionRunner) WithMaxSize(maxSize int) *CachedFunctionRunner {
	r.Options.MaxSize = maxSize
	return r
}

// CleanCache removes expired entries from the cache
func (r *CachedFunctionRunner) CleanCache() {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	now := time.Now()
	for key, entry := range r.cache {
		if entry.Expiration.Before(now) {
			delete(r.cache, key)
		}
	}
}

// ClearCache removes all entries from the cache
func (r *CachedFunctionRunner) ClearCache() {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	r.cache = make(map[string]*CacheEntry)
	r.hitCount = 0
	r.missCount = 0
}

// CacheSize returns the number of entries in the cache
func (r *CachedFunctionRunner) CacheSize() int {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	return len(r.cache)
}

// CacheStats returns cache hit and miss statistics
func (r *CachedFunctionRunner) CacheStats() (hits, misses int) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	return r.hitCount, r.missCount
}

// ExecuteFunc executes a function with caching
func (r *CachedFunctionRunner) ExecuteFunc(
	module *typesys.Module,
	funcSymbol *typesys.Symbol,
	args ...interface{}) (interface{}, error) {

	// Generate a cache key for this function execution
	cacheKey, err := r.generateCacheKey(module, funcSymbol, args...)
	if err != nil {
		// If we can't generate a cache key, just execute the function without caching
		return r.FunctionRunner.ExecuteFunc(module, funcSymbol, args...)
	}

	// Check if result is in cache
	r.mutex.RLock()
	entry, found := r.cache[cacheKey]
	r.mutex.RUnlock()

	// If found and not expired, return the cached result
	if found && entry.Expiration.After(time.Now()) {
		r.mutex.Lock()
		r.hitCount++
		r.mutex.Unlock()
		return entry.Result, entry.Error
	}

	// If not found or expired, execute the function
	r.mutex.Lock()
	r.missCount++
	r.mutex.Unlock()

	result, err := r.FunctionRunner.ExecuteFunc(module, funcSymbol, args...)

	// Cache the result if caching is enabled for this result
	if err == nil || !r.Options.DisableCacheOnError {
		r.mutex.Lock()
		defer r.mutex.Unlock()

		// Check if we need to evict entries due to size limit
		if r.Options.MaxSize > 0 && len(r.cache) >= r.Options.MaxSize {
			// Simple eviction strategy: remove the oldest entry
			var oldestKey string
			var oldestTime time.Time
			first := true

			for k, e := range r.cache {
				if first || e.Timestamp.Before(oldestTime) {
					oldestKey = k
					oldestTime = e.Timestamp
					first = false
				}
			}

			if oldestKey != "" {
				delete(r.cache, oldestKey)
			}
		}

		// Add the new entry to the cache
		now := time.Now()
		r.cache[cacheKey] = &CacheEntry{
			Result:     result,
			Error:      err,
			Timestamp:  now,
			Expiration: now.Add(r.Options.TTL),
		}
	}

	return result, err
}

// generateCacheKey creates a cache key for a function execution
func (r *CachedFunctionRunner) generateCacheKey(
	module *typesys.Module,
	funcSymbol *typesys.Symbol,
	args ...interface{}) (string, error) {

	// Create a data structure to hash
	data := struct {
		ModulePath  string
		PackagePath string
		FuncName    string
		Args        []interface{}
	}{
		ModulePath:  module.Path,
		PackagePath: funcSymbol.Package.ImportPath,
		FuncName:    funcSymbol.Name,
		Args:        args,
	}

	// Serialize to JSON
	jsonData, err := json.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("failed to serialize cache key: %w", err)
	}

	// Hash the JSON data
	hash := sha256.Sum256(jsonData)
	return hex.EncodeToString(hash[:]), nil
}
