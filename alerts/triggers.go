package alerts

import (
	//"github.com/evergreen-ci/evergreen"
	"github.com/evergreen-ci/evergreen/model"
	"github.com/evergreen-ci/evergreen/model/alertrecord"
	"github.com/evergreen-ci/evergreen/model/host"
	"github.com/evergreen-ci/evergreen/model/task"
	"github.com/evergreen-ci/evergreen/model/version"
	"gopkg.in/mgo.v2/bson"
)

// Trigger is a rule that determines if an alert should be queued for some event.
type Trigger interface {
	// Id is a string that identifies this trigger logic. It's used to store in the database for
	// alerts that are queued to identify the reason the alert was generated.
	Id() string

	// Display is a human-readable description of the trigger logic. It may be used in log messages
	// for debugging or for presentation in a UI control panel.
	Display() string

	// ShouldExecute returns a boolean indicating if this trigger's logic has matched and should
	// result in an alert being queued. This logic may involve querying for bookkeeping
	// generated by CreateBookkeepingRecord in previous executions.
	ShouldExecute(ctx triggerContext) (bool, error)

	// CreateAlertRecord returns an instance of AlertRecord which, after insertion, should allow
	// this trigger to detect previous executions of itself in its ShouldExecute function.
	CreateAlertRecord(ctx triggerContext) *alertrecord.AlertRecord
}

// triggerContext is a set of data about the situation in which a trigger is tested. The trigger
// can use this logic to 1) determine if it should execute, and 2) store a bookkeeping record
// to prevent it from double-execution. Some fields may be ignored or irrelevant depending on the
// trigger in use.
type triggerContext struct {
	projectRef        *model.ProjectRef
	version           *version.Version
	task              *task.Task
	previousCompleted *task.Task
	host              *host.Host
}

var (
	// AvailableTaskTriggers is a list of all the supported task triggers, which is used by the UI
	// package to generate a control panel for configuring how to react to these triggers.
	AvailableTaskFailTriggers = []Trigger{TaskFailed{},
		FirstFailureInVersion{},
		FirstFailureInVariant{},
		FirstFailureInTaskType{},
		TaskFailTransition{},
	}

	AvailableProjectTriggers = []Trigger{
		LastRevisionNotFound{},
	}

	SpawnWarningTriggers = []Trigger{SpawnTwoHourWarning{}, SpawnTwelveHourWarning{}}
)

// newAlertRecord creates an instance of an alert record for the given alert type, populating it
// with as much data from the triggerContext as possible
func newAlertRecord(ctx triggerContext, alertType string) *alertrecord.AlertRecord {
	record := &alertrecord.AlertRecord{
		Id:   bson.NewObjectId(),
		Type: alertType,
	}
	if ctx.task != nil {
		record.ProjectId = ctx.task.Project
		record.VersionId = ctx.task.Version
		record.RevisionOrderNumber = ctx.task.RevisionOrderNumber
		record.TaskName = ctx.task.DisplayName
		record.Variant = ctx.task.BuildVariant
		record.TaskId = ctx.task.Id
		record.HostId = ctx.task.HostId
	}

	if ctx.host != nil {
		record.HostId = ctx.host.Id
	}

	return record
}
