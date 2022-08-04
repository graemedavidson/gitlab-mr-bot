package main

import (
	"fmt"
	"net/http"
	"strconv"
	"text/template"
	"time"

	log "github.com/sirupsen/logrus"
)

type cacheHandler struct {
	cache  *localCache
	config Config
}

type cacheResponseData struct {
	UserList     []userList
	Response     cacheFormResponse
	UserStatuses map[string]int
	ServerTime   time.Time
}

type cacheFormResponse struct {
	Result   string
	Username string
	Error    string
}

func (c cacheHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	var cfr cacheFormResponse
	t := time.Now()

	if request.Method == http.MethodPost {
		username := request.FormValue("username")
		if request.FormValue("clear") == "clear" {
			log.Debug("cache admin: clearing cache for user: ", username)
			err := c.cache.clear(username)
			if err != nil {
				cfr = cacheFormResponse{"cleared", username, err.Error()}
			} else {
				cfr = cacheFormResponse{Result: "cleared", Username: username}
			}
		} else if request.FormValue("update") == "update" {
			log.Debug("cache admin: updating cache for user: ", username)

			slackStatus := request.FormValue("slackStatus")
			slackUserID := request.FormValue("slackUserID")

			// Custom Expire
			customExpireWeeks, _ := strconv.Atoi(request.FormValue("customExpireWeeks"))
			customExpireDays, _ := strconv.Atoi(request.FormValue("customExpireDays"))
			customExpireHours, _ := strconv.Atoi(request.FormValue("customExpireHours"))

			var expire time.Time
			if customExpireHours > 0 || customExpireDays > 0 || customExpireWeeks > 0 {
				expire = t.Add(time.Hour * time.Duration(customExpireHours+(customExpireDays*24)+(customExpireWeeks*7*24)))
			} else {
				ttl := getStatusTTL(c.config.UserStatuses, slackStatus)
				expire = t.Add(time.Hour * time.Duration(ttl))
			}

			u := userMeta{username: username, slackUserID: slackUserID, status: slackStatus}
			c.cache.update(u, expire.Unix())
			cfr = cacheFormResponse{"updated", username, ""}
		} else if request.FormValue("delete") == "delete" {
			log.Debug("cache admin: deleting cache entry for user: ", username)
			err := c.cache.delete(username)
			if err != nil {
				cfr = cacheFormResponse{"deleted", username, err.Error()}
			} else {
				cfr = cacheFormResponse{Result: "deleted", Username: username}
			}
		} else {
			log.Error("cache admin: unexpected form entry when dealing with: ", username)
		}
	}

	promCacheAdmin.Inc()

	testTemplate, err := template.New("index.html").ParseFiles("./templates/index.html")
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Error("handling MergeEvent request.")
		writer.WriteHeader(500)
		_, err := writer.Write([]byte(fmt.Sprintf("error handling the event: %v", err)))
		if err != nil {
			log.WithFields(log.Fields{"error": err}).Error("failed to write fail header to external connection.")
		}
		return
	}

	err = testTemplate.Execute(writer, cacheResponseData{c.cache.getUserList(), cfr, c.config.UserStatuses, t.Local()})
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Error("handling MergeEvent request.")
		writer.WriteHeader(500)
		_, err := writer.Write([]byte(fmt.Sprintf("error handling the event: %v", err)))
		if err != nil {
			log.WithFields(log.Fields{"error": err}).Error("failed to write fail header to external connection.")
		}
		return
	}
}
