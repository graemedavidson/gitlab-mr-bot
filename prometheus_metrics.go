package main

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	promEvents = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "gitlab_mr_wh_events",
		Help: "The total number of gitlab events.",
	},
		[]string{
			"type",
			"group",
		},
	)

	promProcessedMRs = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "gitlab_mr_wh_processed_mrs",
		Help: "The total number of gitlab merge requests that require assigning reviewers.",
	},
		[]string{
			"group",
		},
	)

	promRemoveReviewer = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "gitlab_mr_wh_review_removed",
		Help: "The total number of times a merge request is set to WIP and then a reviewer removed.",
	},
		[]string{
			"group",
		},
	)

	promRecursiveCalls = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "gitlab_mr_wh_recursive_calls",
		Help: "When this service calls itself after updating a merge request activing the webhooks.",
	},
		[]string{
			"group",
		},
	)

	promErrors = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "gitlab_mr_wh_errors",
		Help: "The total number of errors encountered handling events.",
	},
		[]string{
			"error",
		},
	)

	promIgnoreActions = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "gitlab_mr_wh_ignore_actions",
		Help: "The total number of ignored actions (merge, approve) handling events.",
	},
		[]string{
			"action",
			"group",
		},
	)

	promSlackAPIReqs = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "gitlab_mr_wh_slack_api_reqs",
		Help: "The total number of slack api requests",
	},
		[]string{
			"request",
		},
	)

	promSlackAPIErrs = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "gitlab_mr_wh_slack_api_errors",
		Help: "The total number of slack api request errors",
	},
		[]string{
			"request",
			"error",
		},
	)

	promSlackMsgs = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "gitlab_mr_wh_slack_msgs",
		Help: "The total number of slack messages sent.",
	},
		[]string{
			"group",
			"channel",
		},
	)

	promSlackMsgsErrors = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "gitlab_mr_wh_slack_msgs_errors",
		Help: "Errors encountered when attempting to send slack messages",
	},
		[]string{
			"error",
			"group",
			"channel",
		},
	)

	promGitlabReqs = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "gitlab_mr_wh_gitlab_reqs",
		Help: "The total number of gitlab requests made.",
	},
		[]string{
			"request",
			"method",
			"group",
		},
	)

	promCacheRead = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "gitlab_mr_wh_cache_read",
		Help: "Cache reads with hit/miss labels with a reason for miss.",
	},
		[]string{
			"response",
			"reason",
		},
	)

	promCacheUpdate = promauto.NewCounter(prometheus.CounterOpts{
		Name: "gitlab_mr_wh_cache_updates",
		Help: "Cache updates.",
	})

	promCacheDelete = promauto.NewCounter(prometheus.CounterOpts{
		Name: "gitlab_mr_wh_cache_delete",
		Help: "Cache delete.",
	})

	promCacheClear = promauto.NewCounter(prometheus.CounterOpts{
		Name: "gitlab_mr_wh_cache_clear",
		Help: "Cache entry cleared.",
	})

	promCacheAdmin = promauto.NewCounter(prometheus.CounterOpts{
		Name: "gitlab_mr_wh_cache_admin",
		Help: "Cache Admin page accessed.",
	})

	promSlackUsersMissing = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "gitlab_mr_wh_no_matching_slack_user",
		Help: "A gitlab user in the codeowners does not have a matching entry in slack.",
	},
		[]string{
			"group",
		},
	)

	promSlackStatusUnavailable = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "gitlab_mr_wh_slack_status_unavailable",
		Help: "slack status of user means they are unavailable as an approver.",
	},
		[]string{
			"reason",
			"group",
		},
	)

	promWorkers = promauto.NewCounter(prometheus.CounterOpts{
		Name: "gitlab_mr_wh_workers",
		Help: "Number of workers created.",
	})

	promWorkersWorking = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "gitlab_mr_wh_workers_working",
		Help: "Number of workers working.",
	})
)
