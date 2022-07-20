// A cache store (in memory) for storing user slack status against unique user id
package main

import (
	"errors"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

type userMeta struct {
	username    string
	slackUserID string
	status      string
}

type cachedUser struct {
	user            userMeta
	expireTimestamp int64
}

type localCache struct {
	stop chan struct{}

	wg    sync.WaitGroup
	mu    sync.RWMutex
	users map[string]cachedUser
}

var (
	errUserNotInCache = errors.New("no_user_in_cache")
	errUserExpired    = errors.New("user_data_expired")
)

func newLocalCache() *localCache {
	lc := &localCache{
		users: make(map[string]cachedUser),
		stop:  make(chan struct{}),
	}

	lc.wg.Add(1)
	go func() {
		defer lc.wg.Done()
	}()

	return lc
}

// update: adds new cache entry otherwise udpates existing
func (lc *localCache) update(u userMeta, expireTimestamp int64) {
	lc.mu.Lock()
	defer lc.mu.Unlock()

	lc.users[u.username] = cachedUser{
		user:            u,
		expireTimestamp: expireTimestamp,
	}

	promCacheUpdate.Inc()
	log.WithFields(log.Fields{"func": "update", "username": u.username}).Debug("user cache updated.")
}

func (lc *localCache) read(username string) (userMeta, error) {
	lc.mu.RLock()
	defer lc.mu.RUnlock()

	cu, ok := lc.users[username]
	if !ok {
		promCacheRead.WithLabelValues("miss", "username_not_found").Inc()
		log.WithFields(log.Fields{"func": "read", "cache": "miss", "reason": "username_not_found", "username": username}).Debug("miss.")
		return userMeta{}, errUserNotInCache
	}

	t := time.Now()
	ct := time.Unix(cu.expireTimestamp, 0)

	if t.After(ct) {
		promCacheRead.WithLabelValues("miss", "expired").Inc()
		log.WithFields(log.Fields{"func": "read", "cache": "miss", "reason": "expired", "username": username}).Debug("expired.")
		return cu.user, errUserExpired
	}

	promCacheRead.WithLabelValues("hit", "").Inc()
	log.WithFields(log.Fields{"func": "read", "cache": "hit", "reason": "", "username": username}).Debug("hit.")
	return cu.user, nil
}

// delete: currently does not need this and expected to keep cache forever
func (lc *localCache) delete(username string) error {
	lc.mu.Lock()
	defer lc.mu.Unlock()

	_, ok := lc.users[username]
	if !ok {
		log.WithFields(log.Fields{"func": "delete", "reason": "username_not_found", "username": username}).Debug("delete failed.")
		return errUserNotInCache
	}

	promCacheDelete.Inc()
	log.WithFields(log.Fields{"func": "delete", "username": username}).Debug("delete successful.")

	delete(lc.users, username)

	return nil
}

// getMissingIDs: return a list of usernames if they do not have a slackUserID set in the cache
func (lc *localCache) getMissingIDs(userIDs ...string) []string {
	var missingIDs []string

	// No users in cache so return all
	if len(lc.users) == 0 {
		return userIDs
	}

	for _, userID := range userIDs {
		for k, v := range lc.users {
			if userID == k {
				if v.user.slackUserID == "" {
					missingIDs = append(missingIDs, userID)
				}
				break
			}
		}
	}

	return missingIDs
}

// clear: Clear a users status and expired time from the cache based on username
func (lc *localCache) clear(username string) error {
	lc.mu.RLock()
	defer lc.mu.RUnlock()

	cu, ok := lc.users[username]
	if !ok {
		log.WithFields(log.Fields{"func": "clear", "reason": "username_not_found", "username": username}).Debug("username not found in cache.")
		return errUserNotInCache
	}

	lc.users[username] = cachedUser{
		user: userMeta{username: cu.user.username, slackUserID: cu.user.slackUserID},
	}

	promCacheClear.Inc()
	log.WithFields(log.Fields{"func": "clear", "username": username}).Debug("user cache cleared.")
	return nil
}

type userList struct {
	Username    string
	SlackUserID string
	SlackStatus string
	CacheExpire time.Time
	Expired     bool
}

func (lc *localCache) getUserList() []userList {
	ul := []userList{}
	t := time.Now()

	for k, v := range lc.users {
		ct := time.Unix(v.expireTimestamp, 0)
		var expiredTS bool
		if t.After(ct) {
			expiredTS = true
		}
		ul = append(ul, userList{Username: k, SlackUserID: v.user.slackUserID, SlackStatus: v.user.status, CacheExpire: time.Unix(v.expireTimestamp, 0), Expired: expiredTS})
	}
	return ul
}
