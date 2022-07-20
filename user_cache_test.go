package main

import (
	// "fmt"
	// "errors"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// Tests

func TestUpdate(t *testing.T) {
	cache1 := userMeta{username: "test1", slackUserID: "1"}
	cache2 := userMeta{username: "test2", slackUserID: "2"}
	timeNow := time.Now()
	timeExpire := timeNow.Add(time.Hour * 8)
	timeExpired := timeNow.AddDate(-1, 0, 0)

	type test struct {
		user userMeta
		got  map[string]cachedUser
	}

	tests := []test{
		{
			cache1,
			map[string]cachedUser{"test1": {cache1, timeExpire.Unix()}, "test2": {cache2, timeExpired.Unix()}},
		},
		{
			cache2,
			map[string]cachedUser{"test1": {cache1, timeExpired.Unix()}, "test2": {cache2, timeExpire.Unix()}},
		},
	}

	for _, tc := range tests {
		cache := newLocalCache()

		cache.update(cache1, timeExpired.Unix())
		cache.update(cache2, timeExpired.Unix())

		cache.update(tc.user, timeExpire.Unix())
		assert.Equal(t, tc.got, cache.users)
	}
}

func TestRead(t *testing.T) {
	cache1 := userMeta{username: "test1", slackUserID: "1"}
	cache2 := userMeta{username: "test2", slackUserID: "2"}
	timeNow := time.Now()
	timeExpire := timeNow.Add(time.Hour * 8)
	timeExpired := timeNow.AddDate(-1, 0, 0)

	type test struct {
		username string
		got      userMeta
		err      error
	}

	tests := []test{
		{
			"test1",
			userMeta{username: "test1", slackUserID: "1"},
			nil,
		},
		{
			"test2",
			userMeta{username: "test2", slackUserID: "2"},
			errors.New("user_data_expired"),
		},
		{
			"test3",
			userMeta{},
			errors.New("no_user_in_cache"),
		},
	}

	for _, tc := range tests {
		cache := newLocalCache()

		cache.update(cache1, timeExpire.Unix())
		cache.update(cache2, timeExpired.Unix())

		got, err := cache.read(tc.username)
		if err != nil {
			assert.Equal(t, err.Error(), tc.err.Error())
			continue
		}
		assert.NoError(t, err)
		assert.Equal(t, tc.got, got)
	}
}

func TestGetMissingIDs(t *testing.T) {
	cache1 := userMeta{username: "test1", slackUserID: "1"}
	cache2 := userMeta{username: "test2", slackUserID: "2"}
	cache3 := userMeta{username: "test3"}
	timeNow := time.Now()
	timeExpire := timeNow.Add(time.Hour * 8)

	type test struct {
		userIDs []string
		got     []string
	}

	tests := []test{
		{
			[]string{"test1", "test2", "test3"},
			[]string{"test3"},
		},
		{
			[]string{"test1", "test2", "test3", "test4"},
			[]string{"test3"},
		},
		{
			[]string{"test1", "test2"},
			nil,
		},
	}

	for _, tc := range tests {
		cache := newLocalCache()

		cache.update(cache1, timeExpire.Unix())
		cache.update(cache2, timeExpire.Unix())
		cache.update(cache3, timeExpire.Unix())

		got := cache.getMissingIDs(tc.userIDs...)
		assert.Equal(t, tc.got, got)
	}
}

func TestDelete(t *testing.T) {
	cache1 := userMeta{username: "test1", slackUserID: "1"}
	cache2 := userMeta{username: "test2", slackUserID: "2"}
	timeNow := time.Now()
	timeExpire := timeNow.Add(time.Hour * 8)
	timeExpired := timeNow.AddDate(-1, 0, 0)

	type test struct {
		username string
		err      error
	}

	tests := []test{
		{
			"test1",
			nil,
		},
		{
			"test3",
			errors.New("no_user_in_cache"),
		},
	}

	for _, tc := range tests {
		cache := newLocalCache()

		cache.update(cache1, timeExpire.Unix())
		cache.update(cache2, timeExpired.Unix())

		err := cache.delete(tc.username)
		if err != nil {
			assert.Equal(t, err.Error(), tc.err.Error())
			continue
		}
		assert.NoError(t, err)
	}
}

func TestClear(t *testing.T) {
	cache1 := userMeta{username: "test1", slackUserID: "1"}
	cache2 := userMeta{username: "test2", slackUserID: "2"}
	timeNow := time.Now()
	timeExpire := timeNow.Add(time.Hour * 8)
	timeExpired := timeNow.AddDate(-1, 0, 0)

	type test struct {
		username string
		err      error
	}

	tests := []test{
		{
			"test1",
			nil,
		},
		{
			"test3",
			errors.New("no_user_in_cache"),
		},
	}

	for _, tc := range tests {
		cache := newLocalCache()

		cache.update(cache1, timeExpire.Unix())
		cache.update(cache2, timeExpired.Unix())

		err := cache.clear(tc.username)
		if err != nil {
			assert.Equal(t, err.Error(), tc.err.Error())
			continue
		}
		assert.NoError(t, err)
	}
}

func TestGetUserList(t *testing.T) {
	cache1 := userMeta{username: "test1", slackUserID: "1"}
	cache2 := userMeta{username: "test2", slackUserID: "2"}
	timeNow := time.Now()
	timeExpire := timeNow.Add(time.Hour * 8)
	timeExpired := timeNow.AddDate(-1, 0, 0)

	cache := newLocalCache()
	cache.update(cache1, timeExpire.Unix())
	cache.update(cache2, timeExpired.Unix())

	type test struct {
		got []userList
	}

	tests := []test{
		{
			[]userList{
				{Username: "test1", SlackUserID: "1", CacheExpire: time.Unix(timeExpire.Unix(), 0), Expired: false},
				{Username: "test2", SlackUserID: "2", CacheExpire: time.Unix(timeExpired.Unix(), 0), Expired: true},
			},
		},
	}

	for _, tc := range tests {
		ul := cache.getUserList()
		assert.Equal(t, tc.got, ul)
	}
}
