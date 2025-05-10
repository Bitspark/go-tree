package resolve

import (
	"sync"
	"time"

	"bitspark.dev/go-tree/pkg/core/typesys"
)

// CacheEntry represents a cached module
type CacheEntry struct {
	Module      *typesys.Module
	LastAccess  time.Time
	AccessCount int
}

// ResolutionCache provides caching for module resolution
type ResolutionCache struct {
	// Cache by import path
	importCache map[string]*CacheEntry

	// Cache by filesystem path
	pathCache map[string]*CacheEntry

	// Maximum number of entries to keep
	maxEntries int

	// Mutex for thread safety
	mu sync.RWMutex
}

// NewResolutionCache creates a new resolution cache
func NewResolutionCache(maxEntries int) *ResolutionCache {
	if maxEntries <= 0 {
		maxEntries = 100 // Default
	}

	return &ResolutionCache{
		importCache: make(map[string]*CacheEntry),
		pathCache:   make(map[string]*CacheEntry),
		maxEntries:  maxEntries,
	}
}

// Get retrieves a module from the cache by import path
func (c *ResolutionCache) Get(importPath string) (*typesys.Module, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, ok := c.importCache[importPath]
	if !ok {
		return nil, false
	}

	// Update access time and count
	entry.LastAccess = time.Now()
	entry.AccessCount++

	return entry.Module, true
}

// GetByPath retrieves a module from the cache by filesystem path
func (c *ResolutionCache) GetByPath(path string) (*typesys.Module, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, ok := c.pathCache[path]
	if !ok {
		return nil, false
	}

	// Update access time and count
	entry.LastAccess = time.Now()
	entry.AccessCount++

	return entry.Module, true
}

// Put adds a module to the cache
func (c *ResolutionCache) Put(importPath, path string, module *typesys.Module) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Check if we need to evict entries
	if len(c.importCache) >= c.maxEntries {
		c.evictOldest()
	}

	// Create new entry
	entry := &CacheEntry{
		Module:      module,
		LastAccess:  time.Now(),
		AccessCount: 1,
	}

	// Add to caches
	c.importCache[importPath] = entry
	c.pathCache[path] = entry
}

// evictOldest removes the least recently used entry
func (c *ResolutionCache) evictOldest() {
	var oldestKey string
	var oldestTime time.Time

	// Initialize with first entry
	for key, entry := range c.importCache {
		oldestKey = key
		oldestTime = entry.LastAccess
		break
	}

	// Find oldest entry
	for key, entry := range c.importCache {
		if entry.LastAccess.Before(oldestTime) {
			oldestKey = key
			oldestTime = entry.LastAccess
		}
	}

	// Get path from entry
	path := ""
	if entry, ok := c.importCache[oldestKey]; ok {
		// Find corresponding path
		for p, e := range c.pathCache {
			if e == entry {
				path = p
				break
			}
		}
	}

	// Remove from both caches
	delete(c.importCache, oldestKey)
	if path != "" {
		delete(c.pathCache, path)
	}
}

// Clear empties the cache
func (c *ResolutionCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.importCache = make(map[string]*CacheEntry)
	c.pathCache = make(map[string]*CacheEntry)
}

// GetEntryCount returns the number of entries in the cache
func (c *ResolutionCache) GetEntryCount() int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return len(c.importCache)
}

// GetModuleStats returns cache statistics for a module
func (c *ResolutionCache) GetModuleStats(importPath string) (accessCount int, lastAccess time.Time, ok bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, ok := c.importCache[importPath]
	if !ok {
		return 0, time.Time{}, false
	}

	return entry.AccessCount, entry.LastAccess, true
}
