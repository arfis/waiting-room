package translation

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"log"
	"sync"
	"time"
)

// CacheEntry represents a cached translation - todo: for now it is being saved in memory, but in the future it should be saved in a database
type CacheEntry struct {
	SourceText     string
	TargetText     string
	SourceLang     string
	TargetLang     string
	CachedAt       time.Time
	AccessCount    int
	LastAccessedAt time.Time
}

// TranslationCache provides in-memory caching for translations
type TranslationCache struct {
	cache map[string]*CacheEntry
	mutex sync.RWMutex
	// Configuration
	maxCacheSize   int
	expirationTime time.Duration
	enableStats    bool
	// Statistics
	hits         int64
	misses       int64
	totalSavings int64 // Estimated cost savings (number of API calls avoided)
}

// NewTranslationCache creates a new translation cache
func NewTranslationCache(maxSize int, expirationDuration time.Duration) *TranslationCache {
	cache := &TranslationCache{
		cache:          make(map[string]*CacheEntry),
		maxCacheSize:   maxSize,
		expirationTime: expirationDuration,
		enableStats:    true,
	}

	// Start cleanup goroutine
	go cache.cleanupExpiredEntries()

	log.Printf("Translation cache initialized: maxSize=%d, expiration=%s", maxSize, expirationDuration)
	return cache
}

// generateCacheKey creates a unique key for the cache entry
func (c *TranslationCache) generateCacheKey(text, sourceLang, targetLang string) string {
	// Create a hash of the text + language pair for efficient lookup
	data := fmt.Sprintf("%s|%s|%s", text, sourceLang, targetLang)
	hash := md5.Sum([]byte(data))
	return hex.EncodeToString(hash[:])
}

// Get retrieves a translation from cache
func (c *TranslationCache) Get(text, sourceLang, targetLang string) (string, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	key := c.generateCacheKey(text, sourceLang, targetLang)
	entry, exists := c.cache[key]

	if !exists {
		c.misses++
		log.Printf("Cache MISS for text: '%s' (%s->%s)", truncateText(text, 50), sourceLang, targetLang)
		return "", false
	}

	// Check if entry has expired
	if time.Since(entry.CachedAt) > c.expirationTime {
		log.Printf("Cache entry EXPIRED for text: '%s' (age: %s)", truncateText(text, 50), time.Since(entry.CachedAt))
		c.misses++
		// Note: Don't delete here to avoid deadlock, cleanup goroutine will handle it
		return "", false
	}

	// Update access statistics
	entry.AccessCount++
	entry.LastAccessedAt = time.Now()
	c.hits++
	c.totalSavings++

	log.Printf("Cache HIT for text: '%s' (%s->%s) [hits: %d, savings: %d]",
		truncateText(text, 50), sourceLang, targetLang, c.hits, c.totalSavings)

	return entry.TargetText, true
}

// Set stores a translation in cache
func (c *TranslationCache) Set(text, translatedText, sourceLang, targetLang string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// Check cache size limit
	if len(c.cache) >= c.maxCacheSize {
		c.evictLeastRecentlyUsed()
	}

	key := c.generateCacheKey(text, sourceLang, targetLang)
	entry := &CacheEntry{
		SourceText:     text,
		TargetText:     translatedText,
		SourceLang:     sourceLang,
		TargetLang:     targetLang,
		CachedAt:       time.Now(),
		AccessCount:    0,
		LastAccessedAt: time.Now(),
	}

	c.cache[key] = entry
	log.Printf("Cache SET for text: '%s' -> '%s' (%s->%s) [cache size: %d/%d]",
		truncateText(text, 30), truncateText(translatedText, 30),
		sourceLang, targetLang, len(c.cache), c.maxCacheSize)
}

// evictLeastRecentlyUsed removes the least recently used entry
func (c *TranslationCache) evictLeastRecentlyUsed() {
	var oldestKey string
	var oldestTime time.Time

	for key, entry := range c.cache {
		if oldestTime.IsZero() || entry.LastAccessedAt.Before(oldestTime) {
			oldestTime = entry.LastAccessedAt
			oldestKey = key
		}
	}

	if oldestKey != "" {
		delete(c.cache, oldestKey)
		log.Printf("Evicted LRU cache entry: last accessed %s ago", time.Since(oldestTime))
	}
}

// cleanupExpiredEntries periodically removes expired entries
func (c *TranslationCache) cleanupExpiredEntries() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		c.mutex.Lock()

		expiredCount := 0
		now := time.Now()

		for key, entry := range c.cache {
			if now.Sub(entry.CachedAt) > c.expirationTime {
				delete(c.cache, key)
				expiredCount++
			}
		}

		if expiredCount > 0 {
			log.Printf("Cleaned up %d expired cache entries. Current size: %d", expiredCount, len(c.cache))
		}

		c.mutex.Unlock()
	}
}

// GetStats returns cache statistics
func (c *TranslationCache) GetStats() map[string]interface{} {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	totalRequests := c.hits + c.misses
	hitRate := float64(0)
	if totalRequests > 0 {
		hitRate = float64(c.hits) / float64(totalRequests) * 100
	}

	return map[string]interface{}{
		"cache_size":      len(c.cache),
		"max_cache_size":  c.maxCacheSize,
		"hits":            c.hits,
		"misses":          c.misses,
		"total_requests":  totalRequests,
		"hit_rate":        fmt.Sprintf("%.2f%%", hitRate),
		"api_calls_saved": c.totalSavings,
		"expiration_time": c.expirationTime.String(),
	}
}

// LogStats logs current cache statistics
func (c *TranslationCache) LogStats() {
	stats := c.GetStats()
	log.Printf("Translation Cache Stats: %+v", stats)
}

// Clear removes all entries from cache
func (c *TranslationCache) Clear() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.cache = make(map[string]*CacheEntry)
	log.Printf("Translation cache cleared")
}

// truncateText helper function to limit text length in logs
func truncateText(text string, maxLen int) string {
	if len(text) <= maxLen {
		return text
	}
	return text[:maxLen] + "..."
}
