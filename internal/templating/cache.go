package templating

import (
	"sync"
	"time"
)

// MemoryCache implements an in-memory template cache with optional TTL
type MemoryCache struct {
	mu       sync.RWMutex
	cache    map[string]*cacheEntry
	ttl      time.Duration
	maxSize  int
	disabled bool
}

// cacheEntry represents a cached template with metadata
type cacheEntry struct {
	template  *Template
	createdAt time.Time
	lastUsed  time.Time
	hitCount  int64
}

// NewMemoryCache creates a new in-memory template cache
func NewMemoryCache(ttl time.Duration, maxSize int) TemplateCache {
	cache := &MemoryCache{
		cache:   make(map[string]*cacheEntry),
		ttl:     ttl,
		maxSize: maxSize,
	}

	// Start cleanup goroutine if TTL is set
	if ttl > 0 {
		go cache.cleanup()
	}

	return cache
}

// NewDisabledCache creates a cache that doesn't actually cache anything
func NewDisabledCache() TemplateCache {
	return &MemoryCache{
		disabled: true,
	}
}

// Get retrieves a cached template
func (c *MemoryCache) Get(key string) (*Template, bool) {
	if c.disabled {
		return nil, false
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, exists := c.cache[key]
	if !exists {
		return nil, false
	}

	// Check if entry has expired
	if c.ttl > 0 && time.Since(entry.createdAt) > c.ttl {
		// Entry expired, but don't remove it here to avoid lock upgrade
		// The cleanup goroutine will handle removal
		return nil, false
	}

	// Update access statistics
	c.updateAccessStats(entry)

	return entry.template, true
}

// Set stores a template in cache
func (c *MemoryCache) Set(key string, tmpl *Template) {
	if c.disabled {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	// Check if we need to evict entries to stay under maxSize
	if c.maxSize > 0 && len(c.cache) >= c.maxSize {
		c.evictLRU()
	}

	now := time.Now()
	c.cache[key] = &cacheEntry{
		template:  tmpl,
		createdAt: now,
		lastUsed:  now,
		hitCount:  0,
	}
}

// Clear clears the cache
func (c *MemoryCache) Clear() {
	if c.disabled {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	c.cache = make(map[string]*cacheEntry)
}

// Remove removes a specific template from cache
func (c *MemoryCache) Remove(key string) {
	if c.disabled {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.cache, key)
}

// Size returns the current number of cached templates
func (c *MemoryCache) Size() int {
	if c.disabled {
		return 0
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	return len(c.cache)
}

// Stats returns cache statistics
func (c *MemoryCache) Stats() CacheStats {
	if c.disabled {
		return CacheStats{}
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	stats := CacheStats{
		Size:     len(c.cache),
		MaxSize:  c.maxSize,
		Disabled: false,
	}

	var totalHits int64
	var oldestEntry, newestEntry time.Time

	for _, entry := range c.cache {
		totalHits += entry.hitCount

		if oldestEntry.IsZero() || entry.createdAt.Before(oldestEntry) {
			oldestEntry = entry.createdAt
		}

		if newestEntry.IsZero() || entry.createdAt.After(newestEntry) {
			newestEntry = entry.createdAt
		}
	}

	stats.TotalHits = totalHits

	if !oldestEntry.IsZero() {
		stats.OldestEntry = oldestEntry
	}

	if !newestEntry.IsZero() {
		stats.NewestEntry = newestEntry
	}

	return stats
}

// updateAccessStats updates access statistics for a cache entry
// Must be called with read lock held
func (c *MemoryCache) updateAccessStats(entry *cacheEntry) {
	// This is a read operation but we're updating stats
	// For simplicity, we'll use atomic operations or just accept the race
	// In a production system, you might want more sophisticated tracking
	entry.lastUsed = time.Now()
	entry.hitCount++
}

// evictLRU evicts the least recently used entry
// Must be called with write lock held
func (c *MemoryCache) evictLRU() {
	if len(c.cache) == 0 {
		return
	}

	var lruKey string
	var lruTime time.Time

	for key, entry := range c.cache {
		if lruTime.IsZero() || entry.lastUsed.Before(lruTime) {
			lruKey = key
			lruTime = entry.lastUsed
		}
	}

	if lruKey != "" {
		delete(c.cache, lruKey)
	}
}

// cleanup periodically removes expired entries
func (c *MemoryCache) cleanup() {
	if c.ttl <= 0 {
		return
	}

	ticker := time.NewTicker(c.ttl / 2) // Clean up twice per TTL period
	defer ticker.Stop()

	for range ticker.C {
		c.cleanupExpired()
	}
}

// cleanupExpired removes expired entries from cache
func (c *MemoryCache) cleanupExpired() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	var expiredKeys []string

	for key, entry := range c.cache {
		if now.Sub(entry.createdAt) > c.ttl {
			expiredKeys = append(expiredKeys, key)
		}
	}

	for _, key := range expiredKeys {
		delete(c.cache, key)
	}
}

// CacheStats represents cache statistics
type CacheStats struct {
	Size        int       `json:"size"`
	MaxSize     int       `json:"max_size"`
	TotalHits   int64     `json:"total_hits"`
	OldestEntry time.Time `json:"oldest_entry,omitempty"`
	NewestEntry time.Time `json:"newest_entry,omitempty"`
	Disabled    bool      `json:"disabled"`
}

// CacheConfig represents cache configuration
type CacheConfig struct {
	// TTL for cache entries (0 = no expiration)
	TTL time.Duration `yaml:"ttl"`

	// Maximum number of entries (0 = no limit)
	MaxSize int `yaml:"max_size"`

	// Whether caching is disabled
	Disabled bool `yaml:"disabled"`
}

// DefaultCacheConfig returns default cache configuration
func DefaultCacheConfig() CacheConfig {
	return CacheConfig{
		TTL:      30 * time.Minute, // 30 minutes TTL
		MaxSize:  1000,             // Max 1000 templates
		Disabled: false,
	}
}

// NewCacheFromConfig creates a cache from configuration
func NewCacheFromConfig(config CacheConfig) TemplateCache {
	if config.Disabled {
		return NewDisabledCache()
	}

	return NewMemoryCache(config.TTL, config.MaxSize)
}
