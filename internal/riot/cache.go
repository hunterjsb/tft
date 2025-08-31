package riot

import (
	"sync"
	"time"
)

// Cache provides a lightweight in-memory cache for expensive Riot API lookups.
// It is safe for concurrent use and supports per-type TTLs with periodic cleanup.
type Cache struct {
	mu sync.RWMutex

	// TTLs
	profileTTL  time.Duration
	matchTTL    time.Duration
	matchIDsTTL time.Duration

	// Data stores
	profiles map[string]cachedItem[*PlayerProfile] // key: PUUID
	matches  map[string]cachedItem[*MatchDto]      // key: matchID
	matchIDs map[string]cachedItem[[]string]       // key: PUUID

	// janitor
	janitorStop chan struct{}
}

// cachedItem wraps a cached value with an expiration time.
type cachedItem[T any] struct {
	value     T
	expiresAt time.Time
}

// NewCache creates a new Cache instance with the provided TTLs.
// If any TTL is <= 0, a sensible default will be used:
// - profileTTL: 1 hour
// - matchTTL: 24 hours
// - matchIDsTTL: 15 minutes
func NewCache(profileTTL, matchTTL, matchIDsTTL time.Duration) *Cache {
	if profileTTL <= 0 {
		profileTTL = time.Hour
	}
	if matchTTL <= 0 {
		matchTTL = 24 * time.Hour
	}
	if matchIDsTTL <= 0 {
		matchIDsTTL = 15 * time.Minute
	}

	return &Cache{
		profileTTL:  profileTTL,
		matchTTL:    matchTTL,
		matchIDsTTL: matchIDsTTL,
		profiles:    make(map[string]cachedItem[*PlayerProfile]),
		matches:     make(map[string]cachedItem[*MatchDto]),
		matchIDs:    make(map[string]cachedItem[[]string]),
	}
}

// NewDefaultCache creates a Cache with default TTLs.
func NewDefaultCache() *Cache {
	return NewCache(time.Hour, 24*time.Hour, 15*time.Minute)
}

// SetProfile caches a PlayerProfile for a PUUID.
func (c *Cache) SetProfile(puuid string, profile *PlayerProfile) {
	if c == nil || profile == nil || puuid == "" {
		return
	}
	now := time.Now()
	exp := now.Add(c.profileTTL)

	c.mu.Lock()
	c.profiles[puuid] = cachedItem[*PlayerProfile]{value: profile, expiresAt: exp}
	c.mu.Unlock()
}

// GetProfile returns a cached PlayerProfile for a PUUID, if present and not expired.
func (c *Cache) GetProfile(puuid string) (*PlayerProfile, bool) {
	if c == nil || puuid == "" {
		return nil, false
	}

	c.mu.RLock()
	item, ok := c.profiles[puuid]
	c.mu.RUnlock()
	if !ok {
		return nil, false
	}

	if time.Now().After(item.expiresAt) {
		// Expired - evict eagerly
		c.mu.Lock()
		delete(c.profiles, puuid)
		c.mu.Unlock()
		return nil, false
	}

	return item.value, true
}

// SetMatch caches a MatchDto for a matchID.
func (c *Cache) SetMatch(matchID string, match *MatchDto) {
	if c == nil || match == nil || matchID == "" {
		return
	}
	now := time.Now()
	exp := now.Add(c.matchTTL)

	c.mu.Lock()
	c.matches[matchID] = cachedItem[*MatchDto]{value: match, expiresAt: exp}
	c.mu.Unlock()
}

// GetMatch returns a cached MatchDto for a matchID, if present and not expired.
func (c *Cache) GetMatch(matchID string) (*MatchDto, bool) {
	if c == nil || matchID == "" {
		return nil, false
	}

	c.mu.RLock()
	item, ok := c.matches[matchID]
	c.mu.RUnlock()
	if !ok {
		return nil, false
	}

	if time.Now().After(item.expiresAt) {
		// Expired - evict eagerly
		c.mu.Lock()
		delete(c.matches, matchID)
		c.mu.Unlock()
		return nil, false
	}

	return item.value, true
}

// SetMatchIDs caches a slice of match IDs for a given PUUID.
func (c *Cache) SetMatchIDs(puuid string, ids []string) {
	if c == nil || puuid == "" || ids == nil {
		return
	}
	now := time.Now()
	exp := now.Add(c.matchIDsTTL)

	// Make a shallow copy to avoid accidental external mutation
	copied := make([]string, len(ids))
	copy(copied, ids)

	c.mu.Lock()
	c.matchIDs[puuid] = cachedItem[[]string]{value: copied, expiresAt: exp}
	c.mu.Unlock()
}

// GetMatchIDs returns cached match IDs for a given PUUID, if present and not expired.
func (c *Cache) GetMatchIDs(puuid string) ([]string, bool) {
	if c == nil || puuid == "" {
		return nil, false
	}

	c.mu.RLock()
	item, ok := c.matchIDs[puuid]
	c.mu.RUnlock()
	if !ok {
		return nil, false
	}

	if time.Now().After(item.expiresAt) {
		// Expired - evict eagerly
		c.mu.Lock()
		delete(c.matchIDs, puuid)
		c.mu.Unlock()
		return nil, false
	}

	// Return a shallow copy to avoid external mutation of cached slice
	out := make([]string, len(item.value))
	copy(out, item.value)
	return out, true
}

// PurgeExpired removes expired entries from all caches.
// This can be called manually or via the janitor.
func (c *Cache) PurgeExpired() {
	if c == nil {
		return
	}
	now := time.Now()

	c.mu.Lock()
	// Profiles
	for k, v := range c.profiles {
		if now.After(v.expiresAt) {
			delete(c.profiles, k)
		}
	}
	// Matches
	for k, v := range c.matches {
		if now.After(v.expiresAt) {
			delete(c.matches, k)
		}
	}
	// Match IDs
	for k, v := range c.matchIDs {
		if now.After(v.expiresAt) {
			delete(c.matchIDs, k)
		}
	}
	c.mu.Unlock()
}

// StartJanitor starts a background goroutine that periodically purges expired entries.
// It returns a function that can be called to stop the janitor.
// If interval <= 0, a default of 5 minutes is used.
func (c *Cache) StartJanitor(interval time.Duration) func() {
	if c == nil {
		return func() {}
	}
	if interval <= 0 {
		interval = 5 * time.Minute
	}

	c.mu.Lock()
	// If already running, stop the previous one
	if c.janitorStop != nil {
		close(c.janitorStop)
	}
	stop := make(chan struct{})
	c.janitorStop = stop
	c.mu.Unlock()

	ticker := time.NewTicker(interval)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				c.PurgeExpired()
			case <-stop:
				return
			}
		}
	}()

	return func() {
		c.mu.Lock()
		if c.janitorStop != nil {
			close(c.janitorStop)
			c.janitorStop = nil
		}
		c.mu.Unlock()
	}
}

// Stats returns the current number of non-expired entries in the cache.
// This performs an eager purge before counting to ensure accuracy.
func (c *Cache) Stats() (profiles int, matches int, matchIDs int) {
	if c == nil {
		return 0, 0, 0
	}
	c.PurgeExpired()

	c.mu.RLock()
	profiles = len(c.profiles)
	matches = len(c.matches)
	matchIDs = len(c.matchIDs)
	c.mu.RUnlock()
	return
}

// SetTTLs updates the TTLs for profiles, matches, and match IDs.
// Any value <= 0 will keep the previous TTL.
func (c *Cache) SetTTLs(profileTTL, matchTTL, matchIDsTTL time.Duration) {
	if c == nil {
		return
	}
	c.mu.Lock()
	if profileTTL > 0 {
		c.profileTTL = profileTTL
	}
	if matchTTL > 0 {
		c.matchTTL = matchTTL
	}
	if matchIDsTTL > 0 {
		c.matchIDsTTL = matchIDsTTL
	}
	c.mu.Unlock()
}
