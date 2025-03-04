package host

import (
	"context"
	"time"

	"github.com/evergreen-ci/evergreen"
	"github.com/evergreen-ci/evergreen/db"
	"github.com/evergreen-ci/evergreen/model/build"
	"github.com/evergreen-ci/evergreen/model/distro"
	"github.com/evergreen-ci/evergreen/model/event"
	"github.com/evergreen-ci/evergreen/model/task"
	"github.com/evergreen-ci/utility"
	"github.com/mongodb/anser/bsonutil"
	adb "github.com/mongodb/anser/db"
	"github.com/mongodb/grip"
	"github.com/mongodb/grip/message"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

const (
	// Collection is the name of the MongoDB collection that stores hosts.
	Collection        = "hosts"
	VolumesCollection = "volumes"
)

var (
	IdKey                              = bsonutil.MustHaveTag(Host{}, "Id")
	DNSKey                             = bsonutil.MustHaveTag(Host{}, "Host")
	SecretKey                          = bsonutil.MustHaveTag(Host{}, "Secret")
	UserKey                            = bsonutil.MustHaveTag(Host{}, "User")
	ServicePasswordKey                 = bsonutil.MustHaveTag(Host{}, "ServicePassword")
	TagKey                             = bsonutil.MustHaveTag(Host{}, "Tag")
	DistroKey                          = bsonutil.MustHaveTag(Host{}, "Distro")
	ProviderKey                        = bsonutil.MustHaveTag(Host{}, "Provider")
	IPKey                              = bsonutil.MustHaveTag(Host{}, "IP")
	IPv4Key                            = bsonutil.MustHaveTag(Host{}, "IPv4")
	ProvisionedKey                     = bsonutil.MustHaveTag(Host{}, "Provisioned")
	ProvisionTimeKey                   = bsonutil.MustHaveTag(Host{}, "ProvisionTime")
	ExtIdKey                           = bsonutil.MustHaveTag(Host{}, "ExternalIdentifier")
	DisplayNameKey                     = bsonutil.MustHaveTag(Host{}, "DisplayName")
	RunningTaskFullKey                 = bsonutil.MustHaveTag(Host{}, "RunningTaskFull")
	RunningTaskKey                     = bsonutil.MustHaveTag(Host{}, "RunningTask")
	RunningTaskExecutionKey            = bsonutil.MustHaveTag(Host{}, "RunningTaskExecution")
	RunningTaskGroupKey                = bsonutil.MustHaveTag(Host{}, "RunningTaskGroup")
	RunningTaskGroupOrderKey           = bsonutil.MustHaveTag(Host{}, "RunningTaskGroupOrder")
	RunningTaskBuildVariantKey         = bsonutil.MustHaveTag(Host{}, "RunningTaskBuildVariant")
	RunningTaskVersionKey              = bsonutil.MustHaveTag(Host{}, "RunningTaskVersion")
	RunningTaskProjectKey              = bsonutil.MustHaveTag(Host{}, "RunningTaskProject")
	CreateTimeKey                      = bsonutil.MustHaveTag(Host{}, "CreationTime")
	ExpirationTimeKey                  = bsonutil.MustHaveTag(Host{}, "ExpirationTime")
	NoExpirationKey                    = bsonutil.MustHaveTag(Host{}, "NoExpiration")
	TerminationTimeKey                 = bsonutil.MustHaveTag(Host{}, "TerminationTime")
	LTCTimeKey                         = bsonutil.MustHaveTag(Host{}, "LastTaskCompletedTime")
	LTCTaskKey                         = bsonutil.MustHaveTag(Host{}, "LastTask")
	LTCGroupKey                        = bsonutil.MustHaveTag(Host{}, "LastGroup")
	LTCBVKey                           = bsonutil.MustHaveTag(Host{}, "LastBuildVariant")
	LTCVersionKey                      = bsonutil.MustHaveTag(Host{}, "LastVersion")
	LTCProjectKey                      = bsonutil.MustHaveTag(Host{}, "LastProject")
	StatusKey                          = bsonutil.MustHaveTag(Host{}, "Status")
	AgentRevisionKey                   = bsonutil.MustHaveTag(Host{}, "AgentRevision")
	NeedsNewAgentKey                   = bsonutil.MustHaveTag(Host{}, "NeedsNewAgent")
	NeedsNewAgentMonitorKey            = bsonutil.MustHaveTag(Host{}, "NeedsNewAgentMonitor")
	JasperCredentialsIDKey             = bsonutil.MustHaveTag(Host{}, "JasperCredentialsID")
	NeedsReprovisionKey                = bsonutil.MustHaveTag(Host{}, "NeedsReprovision")
	StartedByKey                       = bsonutil.MustHaveTag(Host{}, "StartedBy")
	InstanceTypeKey                    = bsonutil.MustHaveTag(Host{}, "InstanceType")
	VolumesKey                         = bsonutil.MustHaveTag(Host{}, "Volumes")
	LastCommunicationTimeKey           = bsonutil.MustHaveTag(Host{}, "LastCommunicationTime")
	UserHostKey                        = bsonutil.MustHaveTag(Host{}, "UserHost")
	ZoneKey                            = bsonutil.MustHaveTag(Host{}, "Zone")
	ProjectKey                         = bsonutil.MustHaveTag(Host{}, "Project")
	ProvisionOptionsKey                = bsonutil.MustHaveTag(Host{}, "ProvisionOptions")
	TaskCountKey                       = bsonutil.MustHaveTag(Host{}, "TaskCount")
	StartTimeKey                       = bsonutil.MustHaveTag(Host{}, "StartTime")
	BillingStartTimeKey                = bsonutil.MustHaveTag(Host{}, "BillingStartTime")
	AgentStartTimeKey                  = bsonutil.MustHaveTag(Host{}, "AgentStartTime")
	TotalIdleTimeKey                   = bsonutil.MustHaveTag(Host{}, "TotalIdleTime")
	HasContainersKey                   = bsonutil.MustHaveTag(Host{}, "HasContainers")
	ParentIDKey                        = bsonutil.MustHaveTag(Host{}, "ParentID")
	ContainerImagesKey                 = bsonutil.MustHaveTag(Host{}, "ContainerImages")
	ContainerBuildAttempt              = bsonutil.MustHaveTag(Host{}, "ContainerBuildAttempt")
	LastContainerFinishTimeKey         = bsonutil.MustHaveTag(Host{}, "LastContainerFinishTime")
	SpawnOptionsKey                    = bsonutil.MustHaveTag(Host{}, "SpawnOptions")
	ContainerPoolSettingsKey           = bsonutil.MustHaveTag(Host{}, "ContainerPoolSettings")
	InstanceTagsKey                    = bsonutil.MustHaveTag(Host{}, "InstanceTags")
	SSHKeyNamesKey                     = bsonutil.MustHaveTag(Host{}, "SSHKeyNames")
	SSHPortKey                         = bsonutil.MustHaveTag(Host{}, "SSHPort")
	HomeVolumeIDKey                    = bsonutil.MustHaveTag(Host{}, "HomeVolumeID")
	PortBindingsKey                    = bsonutil.MustHaveTag(Host{}, "PortBindings")
	IsVirtualWorkstationKey            = bsonutil.MustHaveTag(Host{}, "IsVirtualWorkstation")
	SpawnOptionsTaskIDKey              = bsonutil.MustHaveTag(SpawnOptions{}, "TaskID")
	SpawnOptionsTaskExecutionNumberKey = bsonutil.MustHaveTag(SpawnOptions{}, "TaskExecutionNumber")
	SpawnOptionsBuildIDKey             = bsonutil.MustHaveTag(SpawnOptions{}, "BuildID")
	SpawnOptionsTimeoutKey             = bsonutil.MustHaveTag(SpawnOptions{}, "TimeoutTeardown")
	SpawnOptionsSpawnedByTaskKey       = bsonutil.MustHaveTag(SpawnOptions{}, "SpawnedByTask")
	VolumeIDKey                        = bsonutil.MustHaveTag(Volume{}, "ID")
	VolumeDisplayNameKey               = bsonutil.MustHaveTag(Volume{}, "DisplayName")
	VolumeCreatedByKey                 = bsonutil.MustHaveTag(Volume{}, "CreatedBy")
	VolumeTypeKey                      = bsonutil.MustHaveTag(Volume{}, "Type")
	VolumeSizeKey                      = bsonutil.MustHaveTag(Volume{}, "Size")
	VolumeExpirationKey                = bsonutil.MustHaveTag(Volume{}, "Expiration")
	VolumeNoExpirationKey              = bsonutil.MustHaveTag(Volume{}, "NoExpiration")
	VolumeHostKey                      = bsonutil.MustHaveTag(Volume{}, "Host")
	VolumeMigratingKey                 = bsonutil.MustHaveTag(Volume{}, "Migrating")
	VolumeAttachmentIDKey              = bsonutil.MustHaveTag(VolumeAttachment{}, "VolumeID")
	VolumeDeviceNameKey                = bsonutil.MustHaveTag(VolumeAttachment{}, "DeviceName")
)

var (
	HostsByDistroDistroIDKey          = bsonutil.MustHaveTag(IdleHostsByDistroID{}, "DistroID")
	HostsByDistroIdleHostsKey         = bsonutil.MustHaveTag(IdleHostsByDistroID{}, "IdleHosts")
	HostsByDistroRunningHostsCountKey = bsonutil.MustHaveTag(IdleHostsByDistroID{}, "RunningHostsCount")
)

// === Queries ===

// All is a query that returns all hosts
var All = db.Query(struct{}{})

// ByUserWithRunningStatus produces a query that returns all
// running hosts for the given user id.
func ByUserWithRunningStatus(user string) db.Q {
	return db.Query(
		bson.M{
			StartedByKey: user,
			StatusKey:    bson.M{"$ne": evergreen.HostTerminated},
		})
}

// ByUserRecentlyTerminated produces a query that returns all
// terminated hosts whose TerminationTimeKey is after the given
// timestamp.
func ByUserRecentlyTerminated(user string, timestamp time.Time) db.Q {
	return db.Query(
		bson.M{
			StartedByKey:       user,
			StatusKey:          evergreen.HostTerminated,
			TerminationTimeKey: bson.M{"$gt": timestamp},
		})
}

// IsLive is a query that returns all working hosts started by Evergreen
func IsLive() bson.M {
	return bson.M{
		StartedByKey: evergreen.User,
		StatusKey:    bson.M{"$in": evergreen.ActiveStatus},
	}
}

// ByUserWithUnterminatedStatus produces a query that returns all running hosts
// for the given user id.
func ByUserWithUnterminatedStatus(user string) db.Q {
	return db.Query(
		bson.M{
			StartedByKey: user,
			StatusKey:    bson.M{"$ne": evergreen.HostTerminated},
		},
	)
}

// IdleEphemeralGroupedByDistroId groups and collates the following by distro.Id:
// - []host.Host of ephemeral hosts without containers which having no running task, ordered by {host.CreationTime: 1}
// - the total number of ephemeral hosts that are capable of running tasks
func IdleEphemeralGroupedByDistroID() ([]IdleHostsByDistroID, error) {
	var idlehostsByDistroID []IdleHostsByDistroID
	bootstrapKey := bsonutil.GetDottedKeyName(DistroKey, distro.BootstrapSettingsKey, distro.BootstrapSettingsMethodKey)
	pipeline := []bson.M{
		{
			"$match": bson.M{
				StartedByKey:     evergreen.User,
				ProviderKey:      bson.M{"$in": evergreen.ProviderSpawnable},
				HasContainersKey: bson.M{"$ne": true},
				"$or": []bson.M{
					{
						StatusKey: evergreen.HostRunning,
					},
					{
						StatusKey:    evergreen.HostStarting,
						bootstrapKey: distro.BootstrapMethodUserData,
						// User data hosts have a grace period between creation
						// and provisioning during which they are not considered
						// for idle termination to give agents time to start.
						CreateTimeKey: bson.M{"$lte": time.Now().Add(-provisioningCutoff)},
					},
				},
			},
		},
		{
			"$sort": bson.M{CreateTimeKey: 1},
		},
		{
			"$group": bson.M{
				"_id":                             "$" + bsonutil.GetDottedKeyName(DistroKey, distro.IdKey),
				HostsByDistroRunningHostsCountKey: bson.M{"$sum": 1},
				HostsByDistroIdleHostsKey:         bson.M{"$push": bson.M{"$cond": []interface{}{bson.M{"$eq": []interface{}{"$running_task", primitive.Undefined{}}}, "$$ROOT", primitive.Undefined{}}}},
			},
		},
		{
			"$project": bson.M{"_id": 0, HostsByDistroDistroIDKey: "$_id", HostsByDistroIdleHostsKey: 1, HostsByDistroRunningHostsCountKey: 1},
		},
	}

	if err := db.Aggregate(Collection, pipeline, &idlehostsByDistroID); err != nil {
		return nil, errors.Wrap(err, "grouping idle hosts by distro ID")
	}

	return idlehostsByDistroID, nil
}

func runningHostsQuery(distroID string) bson.M {
	query := IsLive()
	if distroID != "" {
		query[bsonutil.GetDottedKeyName(DistroKey, distro.IdKey)] = distroID
	}

	return query
}

func idleStartedTaskHostsQuery(distroID string) bson.M {
	query := bson.M{
		StatusKey:      bson.M{"$in": evergreen.StartedHostStatus},
		StartedByKey:   evergreen.User,
		RunningTaskKey: bson.M{"$exists": false},
	}
	if distroID != "" {
		query[bsonutil.GetDottedKeyName(DistroKey, distro.IdKey)] = distroID
	}
	return query
}

func idleHostsQuery(distroID string) bson.M {
	query := bson.M{
		StartedByKey:     evergreen.User,
		ProviderKey:      bson.M{"$in": evergreen.ProviderSpawnable},
		RunningTaskKey:   bson.M{"$exists": false},
		HasContainersKey: bson.M{"$ne": true},
		StatusKey:        evergreen.HostRunning,
	}
	if distroID != "" {
		query[bsonutil.GetDottedKeyName(DistroKey, distro.IdKey)] = distroID
	}
	return query
}

func CountRunningHosts(distroID string) (int, error) {
	num, err := Count(db.Query(runningHostsQuery(distroID)))
	return num, errors.Wrap(err, "counting running hosts")
}

func CountAllRunningDynamicHosts() (int, error) {
	query := IsLive()
	query[ProviderKey] = bson.M{"$in": evergreen.ProviderSpawnable}
	num, err := Count(db.Query(query))
	return num, errors.Wrap(err, "counting running dynamic hosts")
}

// CountIdleStartedTaskHosts returns the count of task hosts that are starting
// and not currently running a task.
func CountIdleStartedTaskHosts() (int, error) {
	num, err := Count(db.Query(idleStartedTaskHostsQuery("")))
	return num, errors.Wrap(err, "counting starting hosts")
}

// IdleHostsWithDistroID, given a distroID, returns a slice of all idle hosts in that distro
func IdleHostsWithDistroID(distroID string) ([]Host, error) {
	q := idleHostsQuery(distroID)
	idleHosts, err := Find(db.Query(q))
	if err != nil {
		return nil, errors.Wrap(err, "finding idle hosts")
	}
	return idleHosts, nil
}

// AllActiveHosts produces a HostGroup for all hosts with UpHost
// status as well as quarantined hosts. These do not count spawn
// hosts.
func AllActiveHosts(distroID string) (HostGroup, error) {
	q := bson.M{
		StartedByKey: evergreen.User,
		StatusKey:    bson.M{"$in": append(evergreen.UpHostStatus, evergreen.HostQuarantined)},
	}

	if distroID != "" {
		q[bsonutil.GetDottedKeyName(DistroKey, distro.IdKey)] = distroID
	}

	activeHosts, err := Find(db.Query(q))
	if err != nil {
		return nil, errors.Wrap(err, "finding active hosts")
	}
	return activeHosts, nil
}

// AllHostsSpawnedByTasksToTerminate finds all hosts spawned by tasks that should be terminated.
func AllHostsSpawnedByTasksToTerminate() ([]Host, error) {
	catcher := grip.NewBasicCatcher()
	var hosts []Host
	timedOutHosts, err := allHostsSpawnedByTasksTimedOut()
	hosts = append(hosts, timedOutHosts...)
	catcher.Wrap(err, "finding hosts that have hit their timeout")

	taskHosts, err := allHostsSpawnedByFinishedTasks()
	hosts = append(hosts, taskHosts...)
	catcher.Wrap(err, "finding hosts whose tasks have finished")

	buildHosts, err := allHostsSpawnedByFinishedBuilds()
	hosts = append(hosts, buildHosts...)
	catcher.Wrap(err, "finding hosts whose builds have finished")

	if catcher.HasErrors() {
		return nil, catcher.Resolve()
	}
	return hosts, nil
}

// allHostsSpawnedByTasksTimedOut finds hosts spawned by tasks that should be terminated because they are past their timeout.
func allHostsSpawnedByTasksTimedOut() ([]Host, error) {
	query := db.Query(bson.M{
		StatusKey: evergreen.HostRunning,
		bsonutil.GetDottedKeyName(SpawnOptionsKey, SpawnOptionsSpawnedByTaskKey): true,
		bsonutil.GetDottedKeyName(SpawnOptionsKey, SpawnOptionsTimeoutKey):       bson.M{"$lte": time.Now()},
	})
	return Find(query)
}

// allHostsSpawnedByFinishedTasks finds hosts spawned by tasks that should be terminated because their tasks have finished.
func allHostsSpawnedByFinishedTasks() ([]Host, error) {
	const runningTasks = "running_tasks"
	pipeline := []bson.M{
		{"$match": bson.M{
			StatusKey: bson.M{"$in": evergreen.UpHostStatus},
			bsonutil.GetDottedKeyName(SpawnOptionsKey, SpawnOptionsSpawnedByTaskKey): true}},
		{"$lookup": bson.M{
			"from":         task.Collection,
			"localField":   bsonutil.GetDottedKeyName(SpawnOptionsKey, SpawnOptionsTaskIDKey),
			"foreignField": task.IdKey,
			"as":           runningTasks,
		}},
		{"$unwind": "$" + runningTasks},
		{"$match": bson.M{
			"$or": []bson.M{
				// If the task is finished, then its host should be terminated.
				{
					bsonutil.GetDottedKeyName(runningTasks, task.StatusKey): bson.M{"$in": evergreen.TaskCompletedStatuses},
				},

				// If the task execution number is greater than the host's task
				// execution number, than the host belongs to an old task and should
				// be terminated.
				{
					"$expr": bson.M{
						"$gt": []string{
							"$" + bsonutil.GetDottedKeyName(runningTasks, task.ExecutionKey),
							"$" + bsonutil.GetDottedKeyName(SpawnOptionsKey, SpawnOptionsTaskExecutionNumberKey),
						},
					},
				},
			},
		}},
		{"$project": bson.M{runningTasks: 0}},
	}
	var hosts []Host
	if err := db.Aggregate(Collection, pipeline, &hosts); err != nil {
		return nil, err
	}
	return hosts, nil
}

// allHostsSpawnedByFinishedBuilds finds hosts spawned by tasks that should be terminated because their builds have finished.
func allHostsSpawnedByFinishedBuilds() ([]Host, error) {
	const runningBuilds = "running_builds"
	pipeline := []bson.M{
		{"$match": bson.M{
			StatusKey: bson.M{"$in": evergreen.UpHostStatus},
			bsonutil.GetDottedKeyName(SpawnOptionsKey, SpawnOptionsSpawnedByTaskKey): true}},
		{"$lookup": bson.M{
			"from":         build.Collection,
			"localField":   bsonutil.GetDottedKeyName(SpawnOptionsKey, SpawnOptionsBuildIDKey),
			"foreignField": build.IdKey,
			"as":           runningBuilds,
		}},
		{"$unwind": "$" + runningBuilds},
		{"$match": bson.M{bsonutil.GetDottedKeyName(runningBuilds, build.StatusKey): bson.M{"$in": build.CompletedStatuses}}},
		{"$project": bson.M{runningBuilds: 0}},
	}
	var hosts []Host
	if err := db.Aggregate(Collection, pipeline, &hosts); err != nil {
		return nil, err
	}
	return hosts, nil
}

// ByTaskSpec returns a query that finds all running hosts that are running a
// task with the given group, buildvariant, project, and version.
func ByTaskSpec(group, buildVariant, project, version string) db.Q {
	return db.Query(
		bson.M{
			StatusKey: bson.M{"$in": []string{evergreen.HostStarting, evergreen.HostRunning}},
			"$or": []bson.M{
				{
					RunningTaskKey:             bson.M{"$exists": "true"},
					RunningTaskGroupKey:        group,
					RunningTaskBuildVariantKey: buildVariant,
					RunningTaskProjectKey:      project,
					RunningTaskVersionKey:      version,
				},
				{
					LTCTaskKey:    bson.M{"$exists": "true"},
					LTCGroupKey:   group,
					LTCBVKey:      buildVariant,
					LTCProjectKey: project,
					LTCVersionKey: version,
				},
			},
		},
	)
}

// NumHostsByTaskSpec returns the number of running hosts that are running a task with
// the given group, buildvariant, project, and version.
func NumHostsByTaskSpec(group, buildVariant, project, version string) (int, error) {
	if group == "" || buildVariant == "" || project == "" || version == "" {
		return 0, errors.Errorf("all arguments must be non-empty strings: (group is '%s', build variant is '%s', "+
			"project is '%s' and version is '%s')", group, buildVariant, project, version)
	}

	numHosts, err := Count(ByTaskSpec(group, buildVariant, project, version))
	if err != nil {
		return 0, errors.Wrap(err, "counting hosts by task spec")
	}

	return numHosts, nil
}

// MinTaskGroupOrderRunningByTaskSpec returns the smallest task group order number for tasks with the
// given group, buildvariant, project, and version that are running on hosts.
// Returns 0 in the case of missing task group order numbers or no hosts.
func MinTaskGroupOrderRunningByTaskSpec(group, buildVariant, project, version string) (int, error) {
	if group == "" || buildVariant == "" || project == "" || version == "" {
		return 0, errors.Errorf("all arguments must be non-empty strings: (group is '%s', build variant is '%s', "+
			"project is '%s' and version is '%s')", group, buildVariant, project, version)
	}

	hosts, err := Find(ByTaskSpec(group, buildVariant, project, version).WithFields(RunningTaskGroupOrderKey).Sort([]string{RunningTaskGroupOrderKey}))
	if err != nil {
		return 0, errors.Wrap(err, "finding hosts by task spec with running task group order")
	}
	minTaskGroupOrder := 0
	//  can look at only one host because we sorted
	if len(hosts) > 0 {
		minTaskGroupOrder = hosts[0].RunningTaskGroupOrder
	}
	return minTaskGroupOrder, nil
}

// IsUninitialized is a query that returns all unstarted + uninitialized Evergreen hosts.
var IsUninitialized = db.Query(
	bson.M{StatusKey: evergreen.HostUninitialized},
)

// FindByProvisioning finds all hosts that are not yet provisioned by the app
// server.
func FindByProvisioning() ([]Host, error) {
	return Find(db.Query(bson.M{
		StatusKey:           evergreen.HostProvisioning,
		NeedsReprovisionKey: bson.M{"$exists": false},
		ProvisionedKey:      false,
	}))
}

// FindByShouldConvertProvisioning finds all hosts that are ready and waiting to
// convert their provisioning type.
func FindByShouldConvertProvisioning() ([]Host, error) {
	return Find(db.Query(bson.M{
		StatusKey:           bson.M{"$in": []string{evergreen.HostProvisioning, evergreen.HostRunning}},
		StartedByKey:        evergreen.User,
		RunningTaskKey:      bson.M{"$exists": false},
		HasContainersKey:    bson.M{"$ne": true},
		ParentIDKey:         bson.M{"$exists": false},
		NeedsReprovisionKey: bson.M{"$in": []ReprovisionType{ReprovisionToNew, ReprovisionToLegacy}},
		"$or": []bson.M{
			{NeedsNewAgentKey: true},
			{NeedsNewAgentMonitorKey: true},
		},
	}))
}

// FindByNeedsToRestartJasper finds all hosts that are ready and waiting to
// restart their Jasper service.
func FindByNeedsToRestartJasper() ([]Host, error) {
	bootstrapKey := bsonutil.GetDottedKeyName(DistroKey, distro.BootstrapSettingsKey, distro.BootstrapSettingsMethodKey)
	return Find(db.Query(bson.M{
		StatusKey:           bson.M{"$in": []string{evergreen.HostProvisioning, evergreen.HostRunning}},
		bootstrapKey:        bson.M{"$in": []string{distro.BootstrapMethodSSH, distro.BootstrapMethodUserData}},
		RunningTaskKey:      bson.M{"$exists": false},
		HasContainersKey:    bson.M{"$ne": true},
		ParentIDKey:         bson.M{"$exists": false},
		NeedsReprovisionKey: ReprovisionRestartJasper,
		"$or": []bson.M{
			{"$and": []bson.M{
				{StartedByKey: bson.M{"$ne": evergreen.User}},
				{UserHostKey: true},
			}},
			{NeedsNewAgentMonitorKey: true},
		},
	}))
}

// IsRunningTask is a query that returns all running hosts with a running task
var IsRunningTask = db.Query(
	bson.M{
		RunningTaskKey: bson.M{"$exists": true},
		StatusKey: bson.M{
			"$ne": evergreen.HostTerminated,
		},
	},
)

// IsTerminated is a query that returns all hosts that are terminated
// (and not running a task).
var IsTerminated = db.Query(
	bson.M{
		RunningTaskKey: bson.M{"$exists": false},
		StatusKey:      evergreen.HostTerminated},
)

// ByDistroIDs produces a query that returns all up hosts of the given distros.
func ByDistroIDs(distroIDs ...string) bson.M {
	distroIDKey := bsonutil.GetDottedKeyName(DistroKey, distro.IdKey)
	return bson.M{
		distroIDKey:  bson.M{"$in": distroIDs},
		StartedByKey: evergreen.User,
		StatusKey:    bson.M{"$in": evergreen.UpHostStatus},
	}
}

// ById produces a query that returns a host with the given id.
func ById(id string) db.Q {
	return db.Query(bson.D{{Key: IdKey, Value: id}})
}

// ByIPAndRunning produces a query that returns a running host with the given ip address.
func ByIPAndRunning(ip string) db.Q {
	return db.Query(bson.M{
		"$or": []bson.M{
			{IPKey: ip},
			{IPv4Key: ip},
		},
		StatusKey: evergreen.HostRunning,
	})
}

// ByDistroIDOrAliasesRunning returns a query that returns all hosts with
// matching distro IDs or aliases.
func ByDistroIDsOrAliasesRunning(distroNames ...string) bson.M {
	distroIDKey := bsonutil.GetDottedKeyName(DistroKey, distro.IdKey)
	distroAliasesKey := bsonutil.GetDottedKeyName(DistroKey, distro.AliasesKey)
	return bson.M{
		StatusKey: evergreen.HostRunning,
		"$or": []bson.M{
			{distroIDKey: bson.M{"$in": distroNames}},
			{distroAliasesKey: bson.M{"$in": distroNames}},
		},
	}
}

// ByIds produces a query that returns all hosts in the given list of ids.
func ByIds(ids []string) db.Q {
	return db.Query(bson.D{
		{
			Key: IdKey,
			Value: bson.D{
				{
					Key:   "$in",
					Value: ids,
				},
			},
		},
	})
}

// FindByJasperCredentialsID finds a host with the given Jasper credentials ID.
func FindOneByJasperCredentialsID(id string) (*Host, error) {
	h := &Host{}
	query := bson.M{JasperCredentialsIDKey: id}
	if err := db.FindOneQ(Collection, db.Query(query), h); err != nil {
		return nil, errors.Wrapf(err, "finding host with Jasper credentials ID '%s'", id)
	}
	return h, nil
}

// IsIdle is a query that returns all running Evergreen hosts with no task.
var IsIdle = db.Query(
	bson.M{
		RunningTaskKey: bson.M{"$exists": false},
		StatusKey:      evergreen.HostRunning,
		StartedByKey:   evergreen.User,
	},
)

// ByNotMonitoredSince produces a query that returns all hosts whose
// last reachability check was before the specified threshold,
// filtering out user-spawned hosts and hosts currently running tasks.
func ByNotMonitoredSince(threshold time.Time) db.Q {
	bootstrapKey := bsonutil.GetDottedKeyName(DistroKey, distro.BootstrapSettingsKey, distro.BootstrapSettingsMethodKey)
	return db.Query(bson.M{
		"$and": []bson.M{
			{RunningTaskKey: bson.M{"$exists": false}},
			{StartedByKey: evergreen.User},
			{"$and": []bson.M{
				{"$or": []bson.M{
					{StatusKey: evergreen.HostRunning},
					// Hosts provisioned with user data which have not started
					// the agent yet may be externally terminated without our
					// knowledge, so they should be monitored.
					{
						StatusKey:      evergreen.HostStarting,
						ProvisionedKey: true,
						bootstrapKey:   distro.BootstrapMethodUserData,
					},
				}},
				{"$or": []bson.M{
					{LastCommunicationTimeKey: bson.M{"$lte": threshold}},
					{LastCommunicationTimeKey: bson.M{"$exists": false}},
				}},
			}},
		},
	})
}

// ByExpiringBetween produces a query that returns any host not running tasks
// that will expire between the specified times.
func ByExpiringBetween(lowerBound time.Time, upperBound time.Time) db.Q {
	return db.Query(bson.M{
		StartedByKey: bson.M{"$ne": evergreen.User},
		StatusKey: bson.M{
			"$nin": []string{evergreen.HostTerminated, evergreen.HostQuarantined},
		},
		ExpirationTimeKey: bson.M{"$gte": lowerBound, "$lte": upperBound},
	})
}

// NeedsAgentDeploy finds hosts which need the agent to be deployed because
// either they do not have an agent yet or their agents have not communicated
// recently.
func NeedsAgentDeploy(currentTime time.Time) bson.M {
	cutoffTime := currentTime.Add(-MaxLCTInterval)
	bootstrapKey := bsonutil.GetDottedKeyName(DistroKey, distro.BootstrapSettingsKey, distro.BootstrapSettingsMethodKey)
	return bson.M{
		StartedByKey:     evergreen.User,
		HasContainersKey: bson.M{"$ne": true},
		ParentIDKey:      bson.M{"$exists": false},
		RunningTaskKey:   bson.M{"$exists": false},
		bootstrapKey:     distro.BootstrapMethodLegacySSH,
		"$and": []bson.M{
			{"$or": []bson.M{
				{StatusKey: evergreen.HostRunning},
				{"$and": []bson.M{
					{StatusKey: evergreen.HostProvisioning},
					{NeedsReprovisionKey: bson.M{"$exists": true, "$ne": ""}},
				}},
			}},
			{"$or": []bson.M{
				{LastCommunicationTimeKey: utility.ZeroTime},
				{LastCommunicationTimeKey: bson.M{"$lte": cutoffTime}},
				{LastCommunicationTimeKey: bson.M{"$exists": false}},
			}},
		},
	}
}

// NeedsAgentMonitorDeploy finds hosts which do not have an agent monitor yet or
// which should have an agent monitor but their agent has not communicated
// recently.
func NeedsAgentMonitorDeploy(currentTime time.Time) bson.M {
	bootstrapKey := bsonutil.GetDottedKeyName(DistroKey, distro.BootstrapSettingsKey, distro.BootstrapSettingsMethodKey)
	return bson.M{
		StartedByKey:     evergreen.User,
		HasContainersKey: bson.M{"$ne": true},
		ParentIDKey:      bson.M{"$exists": false},
		RunningTaskKey:   bson.M{"$exists": false},
		"$and": []bson.M{
			{"$or": []bson.M{
				{StatusKey: evergreen.HostRunning},
				{"$and": []bson.M{
					{StatusKey: evergreen.HostProvisioning},
					{NeedsReprovisionKey: bson.M{"$exists": true, "$ne": ""}},
				}},
			}},
			{"$or": []bson.M{
				{LastCommunicationTimeKey: utility.ZeroTime},
				{LastCommunicationTimeKey: bson.M{"$lte": currentTime.Add(-MaxUncommunicativeInterval)}},
				{LastCommunicationTimeKey: bson.M{"$exists": false}},
			}},
		},
		bootstrapKey: bson.M{"$in": []string{distro.BootstrapMethodSSH, distro.BootstrapMethodUserData}},
	}
}

// ShouldDeployAgent returns legacy hosts with NeedsNewAgent set to true and are
// in a state in which they can deploy agents.
func ShouldDeployAgent() db.Q {
	bootstrapKey := bsonutil.GetDottedKeyName(DistroKey, distro.BootstrapSettingsKey, distro.BootstrapSettingsMethodKey)
	return db.Query(bson.M{
		bootstrapKey:        distro.BootstrapMethodLegacySSH,
		StatusKey:           evergreen.HostRunning,
		StartedByKey:        evergreen.User,
		HasContainersKey:    bson.M{"$ne": true},
		ParentIDKey:         bson.M{"$exists": false},
		RunningTaskKey:      bson.M{"$exists": false},
		NeedsNewAgentKey:    true,
		NeedsReprovisionKey: bson.M{"$exists": false},
	})
}

// ShouldDeployAgentMonitor returns running hosts that need a new agent
// monitor.
func ShouldDeployAgentMonitor() db.Q {
	bootstrapKey := bsonutil.GetDottedKeyName(DistroKey, distro.BootstrapSettingsKey, distro.BootstrapSettingsMethodKey)
	return db.Query(bson.M{
		bootstrapKey:            bson.M{"$in": []string{distro.BootstrapMethodSSH, distro.BootstrapMethodUserData}},
		StatusKey:               evergreen.HostRunning,
		StartedByKey:            evergreen.User,
		HasContainersKey:        bson.M{"$ne": true},
		ParentIDKey:             bson.M{"$exists": false},
		RunningTaskKey:          bson.M{"$exists": false},
		NeedsNewAgentMonitorKey: true,
		NeedsReprovisionKey:     bson.M{"$exists": false},
	})
}

// FindUserDataSpawnHostsProvisioning finds all spawn hosts that have been
// provisioned by the app server but are still being provisioned by user data.
func FindUserDataSpawnHostsProvisioning() ([]Host, error) {
	bootstrapKey := bsonutil.GetDottedKeyName(DistroKey, distro.BootstrapSettingsKey, distro.BootstrapSettingsMethodKey)
	provisioningCutoff := time.Now().Add(-30 * time.Minute)

	hosts, err := Find(db.Query(bson.M{
		StatusKey:      evergreen.HostStarting,
		ProvisionedKey: true,
		// Ignore hosts that have failed to provision within the cutoff.
		ProvisionTimeKey: bson.M{"$gte": provisioningCutoff},
		StartedByKey:     bson.M{"$ne": evergreen.User},
		bootstrapKey:     distro.BootstrapMethodUserData,
	}))
	if err != nil {
		return nil, errors.Wrap(err, "finding user data spawn hosts that are still provisioning themselves")
	}
	return hosts, nil
}

// Removes host intents that have been initializing for more than 3 minutes.
//
// If you pass the empty string as a distroID, it will remove stale
// host intents for *all* distros.
func RemoveStaleInitializing(distroID string) error {
	query := bson.M{
		UserHostKey: false,
		bsonutil.GetDottedKeyName(SpawnOptionsKey, SpawnOptionsSpawnedByTaskKey): bson.M{"$ne": true},
		ProviderKey:   bson.M{"$in": evergreen.ProviderSpawnable},
		StatusKey:     evergreen.HostUninitialized,
		CreateTimeKey: bson.M{"$lt": time.Now().Add(-3 * time.Minute)},
	}

	if distroID != "" {
		key := bsonutil.GetDottedKeyName(DistroKey, distro.IdKey)
		query[key] = distroID
	}

	return db.RemoveAll(Collection, query)
}

// MarkStaleBuildingAsFailed marks building hosts that have been stuck building
// for too long as failed in order to indicate that they're stale and should be
// terminated.
func MarkStaleBuildingAsFailed(distroID string) error {
	distroIDKey := bsonutil.GetDottedKeyName(DistroKey, distro.IdKey)
	spawnedByTaskKey := bsonutil.GetDottedKeyName(SpawnOptionsKey, SpawnOptionsSpawnedByTaskKey)
	query := bson.M{
		distroIDKey:      distroID,
		UserHostKey:      false,
		spawnedByTaskKey: bson.M{"$ne": true},
		ProviderKey:      bson.M{"$in": evergreen.ProviderSpawnable},
		StatusKey:        evergreen.HostBuilding,
		CreateTimeKey:    bson.M{"$lt": time.Now().Add(-15 * time.Minute)},
	}

	if distroID != "" {
		key := bsonutil.GetDottedKeyName(DistroKey, distro.IdKey)
		query[key] = distroID
	}

	q := db.Query(query).Project(bson.M{IdKey: 1})
	hosts := []Host{}
	if err := db.FindAllQ(Collection, q, &hosts); err != nil {
		return errors.WithStack(err)
	}
	var ids []string
	for _, h := range hosts {
		ids = append(ids, h.Id)
	}
	if len(ids) == 0 {
		return nil
	}

	if _, err := db.UpdateAll(Collection, bson.M{
		IdKey: bson.M{"$in": ids},
	}, bson.M{
		"$set": bson.M{StatusKey: evergreen.HostBuildingFailed},
	}); err != nil {
		return errors.Wrap(err, "marking stale building hosts as failed")
	}

	for _, id := range ids {
		event.LogHostStartError(id, "stale building host took too long to start")
		grip.Info(message.Fields{
			"message": "stale building host took too long to start",
			"host_id": id,
			"distro":  distroID,
		})
	}

	return nil
}

// === DB Logic ===

// FindOne gets one Host for the given query.
func FindOne(query db.Q) (*Host, error) {
	host := &Host{}
	err := db.FindOneQ(Collection, query, host)
	if adb.ResultsNotFound(err) {
		return nil, nil
	}
	return host, err
}

func FindOneId(id string) (*Host, error) {
	return FindOne(ById(id))
}

// FindOneByTaskIdAndExecution returns a single host with the given running task ID and execution.
func FindOneByTaskIdAndExecution(id string, execution int) (*Host, error) {
	h := &Host{}
	query := db.Query(bson.M{
		RunningTaskKey:          id,
		RunningTaskExecutionKey: execution,
	})
	err := db.FindOneQ(Collection, query, h)
	if adb.ResultsNotFound(err) {
		return nil, nil
	}
	return h, err
}

// FindOneByIdOrTag finds a host where the given id is stored in either the _id or tag field.
// (The tag field is used for the id from the host's original intent host.)
func FindOneByIdOrTag(id string) (*Host, error) {
	query := db.Query(bson.M{
		"$or": []bson.M{
			{TagKey: id},
			{IdKey: id},
		},
	})
	host, err := FindOne(query)
	if err != nil {
		return nil, errors.Wrapf(err, "finding host with ID or tag '%s'", id)
	}
	return host, nil
}

// Find gets all Hosts for the given query.
func Find(query db.Q) ([]Host, error) {
	hosts := []Host{}
	return hosts, errors.WithStack(db.FindAllQ(Collection, query, &hosts))
}

// Count returns the number of hosts that satisfy the given query.
func Count(query db.Q) (int, error) {
	return db.CountQ(Collection, query)
}

// UpdateOne updates one host.
func UpdateOne(query interface{}, update interface{}) error {
	return db.Update(
		Collection,
		query,
		update,
	)
}

// UpdateAll updates all hosts.
func UpdateAll(query interface{}, update interface{}) error {
	_, err := db.UpdateAll(
		Collection,
		query,
		update,
	)
	return err
}

// UpsertOne upserts a host.
func UpsertOne(query interface{}, update interface{}) (*adb.ChangeInfo, error) {
	return db.Upsert(
		Collection,
		query,
		update,
	)
}

func GetHostsByFromIDWithStatus(id, status, user string, limit int) ([]Host, error) {
	var statusMatch interface{}
	if status != "" {
		statusMatch = status
	} else {
		statusMatch = bson.M{"$in": evergreen.UpHostStatus}
	}

	filter := bson.M{
		IdKey:     bson.M{"$gte": id},
		StatusKey: statusMatch,
	}

	if user != "" {
		filter[StartedByKey] = user
	}

	var query db.Q
	hosts, err := Find(query.Filter(filter).Sort([]string{IdKey}).Limit(limit))
	if err != nil {
		return nil, errors.Wrapf(err, "finding hosts with an ID of '%s' or greater, status '%s', and user '%s'", id, status, user)
	}
	return hosts, nil
}

type HostsInRangeParams struct {
	CreatedBefore time.Time
	CreatedAfter  time.Time
	User          string
	Distro        string
	Status        string
	Region        string
	UserSpawned   bool
}

// FindHostsInRange is a method to find a filtered list of hosts
func FindHostsInRange(params HostsInRangeParams) ([]Host, error) {
	var statusMatch interface{}
	if params.Status != "" {
		statusMatch = params.Status
	} else {
		statusMatch = bson.M{"$in": evergreen.UpHostStatus}
	}

	createTimeFilter := bson.M{"$gt": params.CreatedAfter}
	if !utility.IsZeroTime(params.CreatedBefore) {
		createTimeFilter["$lt"] = params.CreatedBefore
	}

	filter := bson.M{
		StatusKey:     statusMatch,
		CreateTimeKey: createTimeFilter,
	}

	if params.User != "" {
		filter[StartedByKey] = params.User
	}
	if params.Distro != "" {
		filter[bsonutil.GetDottedKeyName(DistroKey, distro.IdKey)] = params.Distro
	}

	if params.UserSpawned {
		filter[UserHostKey] = true
	}

	if params.Region != "" {
		filter[bsonutil.GetDottedKeyName(DistroKey, distro.ProviderSettingsListKey, awsRegionKey)] = params.Region
	}
	hosts, err := Find(db.Query(filter))
	if err != nil {
		return nil, errors.Wrap(err, "finding hosts by filters")
	}
	return hosts, nil
}

type InactiveHostCounts struct {
	HostType string `bson:"_id"`
	Count    int    `bson:"count"`
}

func inactiveHostCountPipeline() []bson.M {
	return []bson.M{
		{
			"$match": bson.M{
				StatusKey: bson.M{
					"$in": []string{evergreen.HostDecommissioned, evergreen.HostQuarantined},
				},
			},
		},
		{
			"$project": bson.M{
				IdKey:       0,
				StatusKey:   1,
				ProviderKey: 1,
			},
		},
		{
			"$group": bson.M{
				"_id": "$" + ProviderKey,
				"count": bson.M{
					"$sum": 1,
				},
			},
		},
	}
}

// FinishTime is a struct for storing pairs of host IDs and last container finish times
type FinishTime struct {
	Id         string    `bson:"_id"`
	FinishTime time.Time `bson:"finish_time"`
}

// aggregation pipeline to compute latest finish time for running hosts with child containers
func lastContainerFinishTimePipeline() []bson.M {
	const output string = "finish_time"
	return []bson.M{
		{
			// matches all running containers
			"$match": bson.M{
				ParentIDKey: bson.M{"$exists": true},
				StatusKey:   evergreen.HostRunning,
			},
		},
		{
			// joins hosts and tasks collections on task ID
			"$lookup": bson.M{
				"from":         task.Collection,
				"localField":   RunningTaskKey,
				"foreignField": IdKey,
				"as":           "task",
			},
		},
		{
			// deconstructs $lookup array
			"$unwind": "$task",
		},
		{
			// groups containers by parent host ID
			"$group": bson.M{
				"_id": "$" + ParentIDKey,
				output: bson.M{
					// computes last container finish time for each host
					"$max": bson.M{
						"$add": []interface{}{bsonutil.GetDottedKeyName("$task", "start_time"),
							// divide by 1000000 to treat duration as milliseconds rather than as nanoseconds
							bson.M{"$divide": []interface{}{bsonutil.GetDottedKeyName("$task", "duration_prediction", "value"), 1000000}},
						},
					},
				},
			},
		},
		{
			// projects only ID and finish time
			"$project": bson.M{
				output: 1,
			},
		},
	}
}

// AggregateLastContainerFinishTimes returns the latest finish time for each host with containers
func AggregateLastContainerFinishTimes() ([]FinishTime, error) {
	var times []FinishTime
	err := db.Aggregate(Collection, lastContainerFinishTimePipeline(), &times)
	if err != nil {
		return nil, errors.Wrap(err, "aggregating parent finish times")
	}
	return times, nil

}

func (h *Host) SetVolumes(volumes []VolumeAttachment) error {
	err := UpdateOne(
		bson.M{
			IdKey: h.Id,
		},
		bson.M{
			"$set": bson.M{
				VolumesKey: volumes,
			},
		})
	if err != nil {
		return errors.Wrap(err, "updating host volumes")
	}
	h.Volumes = volumes
	return nil
}

func (h *Host) AddVolumeToHost(newVolume *VolumeAttachment) error {
	_, err := db.FindAndModify(Collection,
		bson.M{
			IdKey: h.Id,
		}, nil,
		adb.Change{
			Update: bson.M{
				"$push": bson.M{
					VolumesKey: newVolume,
				},
			},
			ReturnNew: true,
		},
		&h,
	)
	if err != nil {
		return errors.Wrap(err, "finding host and adding volume")
	}

	grip.Error(message.WrapError((&Volume{ID: newVolume.VolumeID}).SetHost(h.Id),
		message.Fields{
			"host_id":   h.Id,
			"volume_id": newVolume.VolumeID,
			"op":        "host volume accounting",
			"message":   "problem setting host info on volume records",
		}))

	return nil
}

func (h *Host) RemoveVolumeFromHost(volumeId string) error {
	_, err := db.FindAndModify(Collection,
		bson.M{
			IdKey: h.Id,
		}, nil,
		adb.Change{
			Update: bson.M{
				"$pull": bson.M{
					VolumesKey: bson.M{VolumeAttachmentIDKey: volumeId},
				},
			},
			ReturnNew: true,
		},
		&h,
	)
	if err != nil {
		return errors.Wrap(err, "finding host and removing volume")
	}

	grip.Error(message.WrapError(UnsetVolumeHost(volumeId),
		message.Fields{
			"host_id":   h.Id,
			"volume_id": volumeId,
			"op":        "host volume accounting",
			"message":   "problem un-setting host info on volume records",
		}))

	return nil
}

// FindOne gets one Volume for the given query.
func FindOneVolume(query interface{}) (*Volume, error) {
	v := &Volume{}
	err := db.FindOneQ(VolumesCollection, db.Query(query), v)
	if adb.ResultsNotFound(err) {
		return nil, nil
	}
	return v, err
}

func FindDistroForHost(hostID string) (string, error) {
	h, err := FindOne(ById(hostID))
	if err != nil {
		return "", err
	}
	if h == nil {
		return "", errors.New("host not found")
	}
	return h.Distro.Id, nil
}

func findVolumes(q bson.M) ([]Volume, error) {
	volumes := []Volume{}
	return volumes, db.FindAllQ(VolumesCollection, db.Query(q), &volumes)
}

type ClientOptions struct {
	Provider string `bson:"provider"`
	Region   string `bson:"region"`
	Key      string `bson:"key"`
	Secret   string `bson:"secret"`
}

type EC2ProviderSettings struct {
	Region string `bson:"region"`
	Key    string `bson:"aws_access_key_id"`
	Secret string `bson:"aws_secret_access_key"`
}

var (
	awsRegionKey = bsonutil.MustHaveTag(EC2ProviderSettings{}, "Region")
	awsKeyKey    = bsonutil.MustHaveTag(EC2ProviderSettings{}, "Key")
	awsSecretKey = bsonutil.MustHaveTag(EC2ProviderSettings{}, "Secret")
)

func StartingHostsByClient(limit int) (map[ClientOptions][]Host, error) {
	if limit <= 0 {
		limit = 500
	}
	results := []struct {
		Options ClientOptions `bson:"_id"`
		Hosts   []Host        `bson:"hosts"`
	}{}

	pipeline := []bson.M{
		{
			"$match": bson.M{
				StatusKey:      evergreen.HostStarting,
				ProvisionedKey: false,
			},
		},
		{
			"$sort": bson.M{
				CreateTimeKey: 1,
			},
		},
		{
			"$limit": limit,
		},
		{
			"$project": bson.M{
				"host":          "$$ROOT",
				"settings_list": "$" + bsonutil.GetDottedKeyName(DistroKey, distro.ProviderSettingsListKey),
			},
		},
		{
			"$unwind": "$settings_list",
		},
		{
			"$group": bson.M{
				"_id": bson.M{
					"provider": bsonutil.GetDottedKeyName("$host", DistroKey, distro.ProviderKey),
					"region":   bsonutil.GetDottedKeyName("$settings_list", awsRegionKey),
					"key":      bsonutil.GetDottedKeyName("$settings_list", awsKeyKey),
					"secret":   bsonutil.GetDottedKeyName("$settings_list", awsSecretKey),
				},
				"hosts": bson.M{"$push": "$host"},
			},
		},
	}

	if err := db.Aggregate(Collection, pipeline, &results); err != nil {
		return nil, errors.Wrap(err, "aggregating starting hosts by client options")
	}

	optionsMap := make(map[ClientOptions][]Host)
	for _, result := range results {
		optionsMap[result.Options] = result.Hosts
	}

	return optionsMap, nil
}

// UnsafeReplace atomically removes the host given by the idToRemove and inserts
// a new host toInsert. This is typically done to replace the old host with a
// new one. While the atomic swap is safer than doing it non-atomically, it is
// not sufficient to guarantee application correctness, because other threads
// may still be using the old host document.
// TODO (EVG-15875): set a field containing the external identifier on the host
// document rather than do this host document swap logic.
func UnsafeReplace(ctx context.Context, env evergreen.Environment, idToRemove string, toInsert *Host) error {
	if idToRemove == toInsert.Id {
		return nil
	}

	sess, err := env.Client().StartSession()
	if err != nil {
		return errors.Wrap(err, "starting transaction session")
	}
	defer sess.EndSession(ctx)

	replaceHost := func(sessCtx mongo.SessionContext) (interface{}, error) {
		if err := RemoveStrict(idToRemove); err != nil {
			return nil, errors.Wrapf(err, "removing old host '%s'", idToRemove)
		}

		if err := toInsert.Insert(); err != nil {
			return nil, errors.Wrapf(err, "inserting new host '%s'", toInsert.Id)
		}
		return nil, nil
	}

	txnStart := time.Now()
	_, err = sess.WithTransaction(ctx, replaceHost)
	if err != nil {
		grip.Error(message.WrapError(err, message.Fields{
			"message":     "replacing old host with new host in a transaction",
			"jira_ticket": "EVG-15022",
			"old_host_id": idToRemove,
			"new_host_id": toInsert.Id,
			"distro_id":   toInsert.Distro.Id,
			"duration":    time.Since(txnStart),
		}))
		return errors.Wrap(err, "atomic removal of old host and insertion of new host")
	}

	grip.Info(message.Fields{
		"message":              "replaced host document",
		"host_id":              toInsert.Id,
		"host_tag":             toInsert.Tag,
		"distro":               toInsert.Distro.Id,
		"old_host_id":          idToRemove,
		"transaction_duration": time.Since(txnStart),
	})

	return nil
}
