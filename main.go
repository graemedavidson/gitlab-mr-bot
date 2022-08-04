package main

import (
	"errors"
	"net/http"
	"os"
	"runtime"

	health "github.com/nelkinda/health-go"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	"github.com/xanzy/go-gitlab"
	_ "go.uber.org/automaxprocs"
)

func main() {
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
	})
	logLevel, err := getLogLevel()
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Error("failed to set logging level, using default Warn level.")
	}

	log.SetLevel(logLevel)
	log.WithFields(log.Fields{"log_level": logLevel}).Info("set log level.")

	gitlab_token := os.Getenv("GITLAB_TOKEN")
	if gitlab_token == "" {
		log.WithFields(log.Fields{"var": "GITLAB_TOKEN"}).Fatal("environment variable required.")
	}

	gitlab_url := os.Getenv("GITLAB_URL")
	if gitlab_url == "" {
		log.WithFields(log.Fields{"var": "GITLAB_URL"}).Fatal("environment variable required.")
	}

	webhook_secret := os.Getenv("GITLAB_MR_WH_WEBHOOK_SECRET")
	if webhook_secret == "" {
		log.WithFields(log.Fields{"var": "GITLAB_MR_WH_WEBHOOK_SECRET"}).Fatal("environment variable required.")
	}

	server_port := os.Getenv("GITLAB_MR_WH_LISTEN_PORT")
	if server_port == "" {
		server_port = "8080"
	}

	slack_wh_url := os.Getenv("GITLAB_MR_WH_SLACK_WH_URL")
	if slack_wh_url == "" {
		log.WithFields(log.Fields{"var": "GITLAB_MR_WH_SLACK_WH_URL"}).Warn("to enable slack notifications please set environment variable.")
	}

	slack_token := os.Getenv("GITLAB_MR_WH_SLACK_TOKEN")
	if gitlab_token == "" {
		log.WithFields(log.Fields{"var": "GITLAB_MR_WH_SLACK_TOKEN"}).Fatal("environment variable required.")
	}

	os := osFS{}
	config := &Config{
		ConfigPath: "./config/config.yaml",
	}
	err = config.LoadConfig(&os)
	if err != nil {
		log.Fatal(err)
	}

	git, err := newGitlabClient(gitlab_url, gitlab_token)
	if err != nil {
		log.Fatal(err)
	}

	slack := newSlackClient(slack_token, slack_wh_url)

	scheduler, err := NewScheduler()
	if err != nil {
		log.Fatalf("could not create scheduler: %q.", err)
	}

	for i := 0; i < runtime.NumCPU(); i++ {
		worker := NewWorker()
		scheduler.AddWorker(worker)
		promWorkers.Inc()
	}

	// create user cache
	cache := newLocalCache()

	log.Info("starting scheduler.")
	go scheduler.Run(git, slack, *config, cache)

	gitlab_bot_user_identity, err := getBotUserIdentity(*git)
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Fatal("failed to get bot user identity.")
	}

	wh := webhook{
		Secret:          webhook_secret,
		EventsToAccept:  []gitlab.EventType{gitlab.EventTypeMergeRequest},
		GitlabBotUserID: gitlab_bot_user_identity.ID,
		Requests:        scheduler.requests,
	}

	health_endpoint := health.New(
		health.Health{Version: "v0.1.0"},
	)

	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	mux.Handle("/webhook", wh)
	mux.HandleFunc("/health", health_endpoint.Handler)

	// Handle Cache
	cacheHandler := cacheHandler{
		cache:  cache,
		config: *config,
	}
	mux.Handle("/cache", cacheHandler)

	// handle static files
	fileServer := http.FileServer(http.Dir("./static/css"))
	mux.Handle("/static/", http.StripPrefix("/static", fileServer))

	log.WithFields(log.Fields{"ip": "0.0.0.0", "server port": server_port}).Info("starting web server.")

	if err := http.ListenAndServe("0.0.0.0:"+server_port, mux); err != nil {
		log.WithFields(log.Fields{"error": err}).Fatal("http server failed to start.")
	}
}

// Get Bot User Identity
func getBotUserIdentity(gc Gitlab) (*gitlab.User, error) {
	result, _, err := gc.CurrentUser()
	promGitlabReqs.WithLabelValues("users", "get", "").Inc()
	if err != nil {
		return nil, errors.New("failed to get bot user identity")
	}
	return result, nil
}

// Get logging level through environment variable
func getLogLevel() (log.Level, error) {
	logLevel, exists := os.LookupEnv("GITLAB_MR_WH_LOG_LEVEL")
	if !exists {
		return log.WarnLevel, nil
	}

	level, err := log.ParseLevel(logLevel)
	if err != nil {
		promErrors.WithLabelValues("parse_log_lvl").Inc()
		return log.WarnLevel, err
	}
	return level, nil
}
