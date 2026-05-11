// Package mq holds RabbitMQ protocol constants and connection helpers
// shared between worker consumers and any future publisher.
package mq

// Exchange names — must match the Python eventhandlers protocol exactly.
const (
	ExchangeActions   = "actions"
	ExchangeEvents    = "events"
	ExchangeExports   = "exports"
	ExchangeJobs      = "jobs"
	ExchangeScheduler = "scheduler"
)

// Queue names consumed by the worker.
const (
	QueueActionsTriggers = "actions.triggers"
	QueueEventNew        = "event.new"
	QueueJobTasks        = "job.tasks"
	QueueSchedulerTasks  = "scheduler.tasks"
)

// Routing keys.
const (
	RoutingKeyEventsNew         = "events.new"
	RoutingKeyActionsNewTrigger = "actions.new_actions_trigger"
	RoutingKeyJobsAddTask       = "job.add_task"
	RoutingKeySchedulerAll      = "scheduler.all"
)
