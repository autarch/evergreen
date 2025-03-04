package graphql

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"time"

	"github.com/evergreen-ci/evergreen"
	"github.com/evergreen-ci/evergreen/db"
	"github.com/evergreen-ci/evergreen/model"
	"github.com/evergreen-ci/evergreen/model/distro"
	"github.com/evergreen-ci/evergreen/model/event"
	"github.com/evergreen-ci/evergreen/model/host"
	"github.com/evergreen-ci/evergreen/model/patch"
	"github.com/evergreen-ci/evergreen/model/task"
	"github.com/evergreen-ci/evergreen/model/testresult"
	"github.com/evergreen-ci/evergreen/model/user"
	"github.com/evergreen-ci/evergreen/rest/data"
	restModel "github.com/evergreen-ci/evergreen/rest/model"
	"github.com/evergreen-ci/evergreen/thirdparty"
	"github.com/evergreen-ci/plank"
	"github.com/evergreen-ci/utility"
	"github.com/mongodb/anser/bsonutil"
	"github.com/mongodb/grip"
	"github.com/mongodb/grip/message"
	werrors "github.com/pkg/errors"
)

// BbGetCreatedTickets is the resolver for the bbGetCreatedTickets field.
func (r *queryResolver) BbGetCreatedTickets(ctx context.Context, taskID string) ([]*thirdparty.JiraTicket, error) {
	createdTickets, err := bbGetCreatedTicketsPointers(taskID)
	if err != nil {
		return nil, err
	}

	return createdTickets, nil
}

// BuildBaron is the resolver for the buildBaron field.
func (r *queryResolver) BuildBaron(ctx context.Context, taskID string, execution int) (*BuildBaron, error) {
	execString := strconv.Itoa(execution)

	searchReturnInfo, bbConfig, err := model.GetSearchReturnInfo(taskID, execString)
	if err != nil {
		return nil, InternalServerError.Send(ctx, err.Error())
	}

	return &BuildBaron{
		SearchReturnInfo:        searchReturnInfo,
		BuildBaronConfigured:    bbConfig.ProjectFound && bbConfig.SearchConfigured,
		BbTicketCreationDefined: bbConfig.TicketCreationDefined,
	}, nil
}

// AwsRegions is the resolver for the awsRegions field.
func (r *queryResolver) AwsRegions(ctx context.Context) ([]string, error) {
	return evergreen.GetEnvironment().Settings().Providers.AWS.AllowedRegions, nil
}

// ClientConfig is the resolver for the clientConfig field.
func (r *queryResolver) ClientConfig(ctx context.Context) (*restModel.APIClientConfig, error) {
	envClientConfig := evergreen.GetEnvironment().ClientConfig()
	clientConfig := restModel.APIClientConfig{}
	clientConfig.BuildFromService(*envClientConfig)
	return &clientConfig, nil
}

// InstanceTypes is the resolver for the instanceTypes field.
func (r *queryResolver) InstanceTypes(ctx context.Context) ([]string, error) {
	config, err := evergreen.GetConfig()
	if err != nil {
		return nil, InternalServerError.Send(ctx, "unable to retrieve server config")
	}
	return config.Providers.AWS.AllowedInstanceTypes, nil
}

// SpruceConfig is the resolver for the spruceConfig field.
func (r *queryResolver) SpruceConfig(ctx context.Context) (*restModel.APIAdminSettings, error) {
	config, err := evergreen.GetConfig()
	if err != nil {
		return nil, InternalServerError.Send(ctx, fmt.Sprintf("Error Fetching evergreen settings: %s", err.Error()))
	}

	spruceConfig := restModel.APIAdminSettings{}
	err = spruceConfig.BuildFromService(config)
	if err != nil {
		return nil, InternalServerError.Send(ctx, fmt.Sprintf("Error building api admin settings from service: %s", err.Error()))
	}
	return &spruceConfig, nil
}

// SubnetAvailabilityZones is the resolver for the subnetAvailabilityZones field.
func (r *queryResolver) SubnetAvailabilityZones(ctx context.Context) ([]string, error) {
	zones := []string{}
	for _, subnet := range evergreen.GetEnvironment().Settings().Providers.AWS.Subnets {
		zones = append(zones, subnet.AZ)
	}
	return zones, nil
}

// Distros is the resolver for the distros field.
func (r *queryResolver) Distros(ctx context.Context, onlySpawnable bool) ([]*restModel.APIDistro, error) {
	apiDistros := []*restModel.APIDistro{}

	var distros []distro.Distro
	if onlySpawnable {
		d, err := distro.Find(distro.BySpawnAllowed())
		if err != nil {
			return nil, InternalServerError.Send(ctx, fmt.Sprintf("Error while fetching spawnable distros: %s", err.Error()))
		}
		distros = d
	} else {
		d, err := distro.FindAll()
		if err != nil {
			return nil, InternalServerError.Send(ctx, fmt.Sprintf("Error while fetching distros: %s", err.Error()))
		}
		distros = d
	}
	for _, d := range distros {
		apiDistro := restModel.APIDistro{}
		apiDistro.BuildFromService(d)
		apiDistros = append(apiDistros, &apiDistro)
	}
	return apiDistros, nil
}

// DistroTaskQueue is the resolver for the distroTaskQueue field.
func (r *queryResolver) DistroTaskQueue(ctx context.Context, distroID string) ([]*restModel.APITaskQueueItem, error) {
	distroQueue, err := model.LoadTaskQueue(distroID)
	if err != nil {
		return nil, InternalServerError.Send(ctx, fmt.Sprintf("Error getting task queue for distro %v: %v", distroID, err.Error()))
	}
	if distroQueue == nil {
		return nil, ResourceNotFound.Send(ctx, fmt.Sprintf("cannot find queue with distro ID `%s`", distroID))
	}

	taskQueue := []*restModel.APITaskQueueItem{}

	for _, taskQueueItem := range distroQueue.Queue {
		apiTaskQueueItem := restModel.APITaskQueueItem{}
		apiTaskQueueItem.BuildFromService(taskQueueItem)
		taskQueue = append(taskQueue, &apiTaskQueueItem)
	}

	return taskQueue, nil
}

// Host is the resolver for the host field.
func (r *queryResolver) Host(ctx context.Context, hostID string) (*restModel.APIHost, error) {
	host, err := host.GetHostByIdOrTagWithTask(hostID)
	if err != nil {
		return nil, InternalServerError.Send(ctx, fmt.Sprintf("Error Fetching host: %s", err.Error()))
	}
	if host == nil {
		return nil, werrors.Errorf("unable to find host %s", hostID)
	}

	apiHost := &restModel.APIHost{}
	apiHost.BuildFromService(host, host.RunningTaskFull)
	return apiHost, nil
}

// HostEvents is the resolver for the hostEvents field.
func (r *queryResolver) HostEvents(ctx context.Context, hostID string, hostTag *string, limit *int, page *int) (*HostEvents, error) {
	h, err := host.FindOneByIdOrTag(hostID)
	if err != nil {
		return nil, InternalServerError.Send(ctx, fmt.Sprintf("finding host '%s': %s", hostID, err.Error()))
	}
	if h == nil {
		return nil, ResourceNotFound.Send(ctx, fmt.Sprintf("host '%s' not found", hostID))
	}
	events, count, err := event.MostRecentPaginatedHostEvents(h.Id, h.Tag, *limit, *page)
	if err != nil {
		return nil, InternalServerError.Send(ctx, fmt.Sprintf("fetching host events: %s", err.Error()))
	}
	// populate eventlogs pointer arrays
	apiEventLogPointers := []*restModel.HostAPIEventLogEntry{}
	for _, e := range events {
		apiEventLog := restModel.HostAPIEventLogEntry{}
		err = apiEventLog.BuildFromService(e)
		if err != nil {
			return nil, InternalServerError.Send(ctx, fmt.Sprintf("building APIEventLogEntry from EventLog: %s", err.Error()))
		}
		apiEventLogPointers = append(apiEventLogPointers, &apiEventLog)
	}
	hostevents := HostEvents{
		EventLogEntries: apiEventLogPointers,
		Count:           count,
	}
	return &hostevents, nil
}

// Hosts is the resolver for the hosts field.
func (r *queryResolver) Hosts(ctx context.Context, hostID *string, distroID *string, currentTaskID *string, statuses []string, startedBy *string, sortBy *HostSortBy, sortDir *SortDirection, page *int, limit *int) (*HostsResponse, error) {
	hostIDParam := ""
	if hostID != nil {
		hostIDParam = *hostID
	}
	distroParam := ""
	if distroID != nil {
		distroParam = *distroID
	}
	currentTaskParam := ""
	if currentTaskID != nil {
		currentTaskParam = *currentTaskID
	}
	startedByParam := ""
	if startedBy != nil {
		startedByParam = *startedBy
	}
	sorter := host.StatusKey
	if sortBy != nil {
		switch *sortBy {
		case HostSortByCurrentTask:
			sorter = host.RunningTaskKey
		case HostSortByDistro:
			sorter = bsonutil.GetDottedKeyName(host.DistroKey, distro.IdKey)
		case HostSortByElapsed:
			sorter = "task_full.start_time"
		case HostSortByID:
			sorter = host.IdKey
		case HostSortByIDLeTime:
			sorter = host.TotalIdleTimeKey
		case HostSortByOwner:
			sorter = host.StartedByKey
		case HostSortByStatus:
			sorter = host.StatusKey
		case HostSortByUptime:
			sorter = host.CreateTimeKey
		default:
			sorter = host.StatusKey
		}

	}
	sortDirParam := 1
	if *sortDir == SortDirectionDesc {
		sortDirParam = -1
	}
	pageParam := 0
	if page != nil {
		pageParam = *page
	}
	limitParam := 0
	if limit != nil {
		limitParam = *limit
	}

	hosts, filteredHostsCount, totalHostsCount, err := host.GetPaginatedRunningHosts(hostIDParam, distroParam, currentTaskParam, statuses, startedByParam, sorter, sortDirParam, pageParam, limitParam)
	if err != nil {
		return nil, InternalServerError.Send(ctx, fmt.Sprintf("Error getting hosts: %s", err.Error()))
	}

	apiHosts := []*restModel.APIHost{}

	for _, h := range hosts {
		apiHost := restModel.APIHost{}
		apiHost.BuildFromService(&h, h.RunningTaskFull)
		apiHosts = append(apiHosts, &apiHost)
	}

	return &HostsResponse{
		Hosts:              apiHosts,
		FilteredHostsCount: filteredHostsCount,
		TotalHostsCount:    totalHostsCount,
	}, nil
}

// TaskQueueDistros is the resolver for the taskQueueDistros field.
func (r *queryResolver) TaskQueueDistros(ctx context.Context) ([]*TaskQueueDistro, error) {
	queues, err := model.FindAllTaskQueues()
	if err != nil {
		return nil, InternalServerError.Send(ctx, fmt.Sprintf("Error getting all task queues: %v", err.Error()))
	}

	distros := []*TaskQueueDistro{}

	for _, distro := range queues {
		numHosts, err := host.CountRunningHosts(distro.Distro)
		if err != nil {
			return nil, InternalServerError.Send(ctx, fmt.Sprintf("Error getting associated hosts: %s", err.Error()))
		}
		tqd := TaskQueueDistro{
			ID:        distro.Distro,
			TaskCount: len(distro.Queue),
			HostCount: numHosts,
		}
		distros = append(distros, &tqd)
	}

	// sort distros by task count in descending order
	sort.SliceStable(distros, func(i, j int) bool {
		return distros[i].TaskCount > distros[j].TaskCount
	})

	return distros, nil
}

// Pod is the resolver for the pod field.
func (r *queryResolver) Pod(ctx context.Context, podID string) (*restModel.APIPod, error) {
	pod, err := data.FindAPIPodByID(podID)
	if err != nil {
		return nil, InternalServerError.Send(ctx, fmt.Sprintf("Error finding pod: %s", err.Error()))
	}
	return pod, nil
}

// Patch is the resolver for the patch field.
func (r *queryResolver) Patch(ctx context.Context, id string) (*restModel.APIPatch, error) {
	patch, err := data.FindPatchById(id)
	if err != nil {
		return nil, InternalServerError.Send(ctx, err.Error())
	}

	if evergreen.IsFinishedPatchStatus(*patch.Status) {
		failedAndAbortedStatuses := append(evergreen.TaskFailureStatuses, evergreen.TaskAborted)
		opts := task.GetTasksByVersionOptions{
			Statuses:                       failedAndAbortedStatuses,
			FieldsToProject:                []string{task.DisplayStatusKey},
			IncludeBaseTasks:               false,
			IncludeBuildVariantDisplayName: false,
		}
		tasks, _, err := task.GetTasksByVersion(ctx, id, opts)
		if err != nil {
			return nil, InternalServerError.Send(ctx, fmt.Sprintf("Could not fetch tasks for patch: %s ", err.Error()))
		}

		if len(patch.ChildPatches) > 0 {
			for _, cp := range patch.ChildPatches {
				// add the child patch tasks to tasks so that we can consider their status
				childPatchTasks, _, err := task.GetTasksByVersion(ctx, *cp.Id, opts)
				if err != nil {
					return nil, InternalServerError.Send(ctx, fmt.Sprintf("Could not fetch tasks for child patch: %s ", err.Error()))
				}
				tasks = append(tasks, childPatchTasks...)
			}
		}
		statuses := getAllTaskStatuses(tasks)

		// If theres an aborted task we should set the patch status to aborted if there are no other failures
		if utility.StringSliceContains(statuses, evergreen.TaskAborted) {
			if len(utility.StringSliceIntersection(statuses, evergreen.TaskFailureStatuses)) == 0 {
				patch.Status = utility.ToStringPtr(evergreen.PatchAborted)
			}
		}
	}
	return patch, nil
}

// GithubProjectConflicts is the resolver for the githubProjectConflicts field.
func (r *queryResolver) GithubProjectConflicts(ctx context.Context, projectID string) (*model.GithubProjectConflicts, error) {
	pRef, err := model.FindMergedProjectRef(projectID, "", false)
	if err != nil {
		return nil, InternalServerError.Send(ctx, fmt.Sprintf("error getting project: %v", err.Error()))
	}
	if pRef == nil {
		return nil, InternalServerError.Send(ctx, fmt.Sprintf("project '%s' not found", projectID))
	}

	conflicts, err := pRef.GetGithubProjectConflicts()
	if err != nil {
		return nil, InternalServerError.Send(ctx, fmt.Sprintf("error getting project conflicts: %v", err.Error()))
	}
	return &conflicts, nil
}

// Project is the resolver for the project field.
func (r *queryResolver) Project(ctx context.Context, projectIdentifier string) (*restModel.APIProjectRef, error) {
	project, err := data.FindProjectById(projectIdentifier, true, false)
	if err != nil {
		return nil, InternalServerError.Send(ctx, fmt.Sprintf("Error finding project by id %s: %s", projectIdentifier, err.Error()))
	}
	apiProjectRef := restModel.APIProjectRef{}
	err = apiProjectRef.BuildFromService(*project)
	if err != nil {
		return nil, InternalServerError.Send(ctx, fmt.Sprintf("error building APIProject from service: %s", err.Error()))
	}
	return &apiProjectRef, nil
}

// Projects is the resolver for the projects field.
func (r *queryResolver) Projects(ctx context.Context) ([]*GroupedProjects, error) {
	allProjects, err := model.FindAllMergedTrackedProjectRefs()
	if err != nil {
		return nil, ResourceNotFound.Send(ctx, err.Error())
	}
	// We have to iterate over the merged project refs to verify if they are enabled
	enabledProjects := []model.ProjectRef{}
	for _, p := range allProjects {
		if p.Enabled {
			enabledProjects = append(enabledProjects, p)
		}
	}
	groupedProjects, err := groupProjects(enabledProjects, false)
	if err != nil {
		return nil, InternalServerError.Send(ctx, fmt.Sprintf("error grouping project: %s", err.Error()))
	}
	return groupedProjects, nil
}

// ProjectEvents is the resolver for the projectEvents field.
func (r *queryResolver) ProjectEvents(ctx context.Context, identifier string, limit *int, before *time.Time) (*ProjectEvents, error) {
	timestamp := time.Now()
	if before != nil {
		timestamp = *before
	}
	events, err := data.GetProjectEventLog(identifier, timestamp, utility.FromIntPtr(limit))
	res := &ProjectEvents{
		EventLogEntries: getPointerEventList(events),
		Count:           len(events),
	}
	return res, err
}

// ProjectSettings is the resolver for the projectSettings field.
func (r *queryResolver) ProjectSettings(ctx context.Context, identifier string) (*restModel.APIProjectSettings, error) {
	projectRef, err := model.FindBranchProjectRef(identifier)
	if err != nil {
		return nil, InternalServerError.Send(ctx, fmt.Sprintf("error looking in project collection: %s", err.Error()))
	}
	if projectRef == nil {
		return nil, ResourceNotFound.Send(ctx, "project doesn't exist")
	}

	res := &restModel.APIProjectSettings{
		ProjectRef: restModel.APIProjectRef{},
	}
	if err = res.ProjectRef.BuildFromService(*projectRef); err != nil {
		return nil, InternalServerError.Send(ctx, fmt.Sprintf("error building APIProjectRef from service: %s", err.Error()))
	}
	if !projectRef.UseRepoSettings() {
		// Default values so the UI understands what to do with nil values.
		res.ProjectRef.DefaultUnsetBooleans()
	}
	return res, nil
}

// RepoEvents is the resolver for the repoEvents field.
func (r *queryResolver) RepoEvents(ctx context.Context, id string, limit *int, before *time.Time) (*ProjectEvents, error) {
	timestamp := time.Now()
	if before != nil {
		timestamp = *before
	}
	events, err := data.GetEventsById(id, timestamp, utility.FromIntPtr(limit))
	res := &ProjectEvents{
		EventLogEntries: getPointerEventList(events),
		Count:           len(events),
	}
	return res, err
}

// RepoSettings is the resolver for the repoSettings field.
func (r *queryResolver) RepoSettings(ctx context.Context, id string) (*restModel.APIProjectSettings, error) {
	repoRef, err := model.FindOneRepoRef(id)
	if err != nil {
		return nil, InternalServerError.Send(ctx, fmt.Sprintf("error looking in repo collection: %s", err.Error()))
	}
	if repoRef == nil {
		return nil, ResourceNotFound.Send(ctx, "repo doesn't exist")
	}

	res := &restModel.APIProjectSettings{
		ProjectRef: restModel.APIProjectRef{},
	}
	if err = res.ProjectRef.BuildFromService(repoRef.ProjectRef); err != nil {
		return nil, InternalServerError.Send(ctx, fmt.Sprintf("error building APIProjectRef from service: %s", err.Error()))
	}

	// Default values so the UI understands what to do with nil values.
	res.ProjectRef.DefaultUnsetBooleans()
	return res, nil
}

// ViewableProjectRefs is the resolver for the viewableProjectRefs field.
func (r *queryResolver) ViewableProjectRefs(ctx context.Context) ([]*GroupedProjects, error) {
	usr := mustHaveUser(ctx)
	projectIds, err := usr.GetViewableProjectSettings()
	if err != nil {
		return nil, InternalServerError.Send(ctx, fmt.Sprintf("error getting viewable projects for '%s': '%s'", usr.DispName, err.Error()))
	}

	projects, err := model.FindProjectRefsByIds(projectIds...)
	if err != nil {
		return nil, InternalServerError.Send(ctx, fmt.Sprintf("Error getting projects: %v", err.Error()))
	}

	groupedProjects, err := groupProjects(projects, true)
	if err != nil {
		return nil, InternalServerError.Send(ctx, fmt.Sprintf("error grouping project: %s", err.Error()))
	}
	return groupedProjects, nil
}

// MyHosts is the resolver for the myHosts field.
func (r *queryResolver) MyHosts(ctx context.Context) ([]*restModel.APIHost, error) {
	usr := mustHaveUser(ctx)
	hosts, err := host.Find(host.ByUserWithRunningStatus(usr.Username()))
	if err != nil {
		return nil, InternalServerError.Send(ctx,
			fmt.Sprintf("Error finding running hosts for user %s : %s", usr.Username(), err))
	}
	duration := time.Duration(5) * time.Minute
	timestamp := time.Now().Add(-duration) // within last 5 minutes
	recentlyTerminatedHosts, err := host.Find(host.ByUserRecentlyTerminated(usr.Username(), timestamp))
	if err != nil {
		return nil, InternalServerError.Send(ctx,
			fmt.Sprintf("Error finding recently terminated hosts for user %s : %s", usr.Username(), err))
	}
	hosts = append(hosts, recentlyTerminatedHosts...)

	var apiHosts []*restModel.APIHost
	for _, h := range hosts {
		apiHost := restModel.APIHost{}
		apiHost.BuildFromService(&h, nil)
		apiHosts = append(apiHosts, &apiHost)
	}
	return apiHosts, nil
}

// MyVolumes is the resolver for the myVolumes field.
func (r *queryResolver) MyVolumes(ctx context.Context) ([]*restModel.APIVolume, error) {
	usr := mustHaveUser(ctx)
	volumes, err := host.FindSortedVolumesByUser(usr.Username())
	if err != nil {
		return nil, InternalServerError.Send(ctx, err.Error())
	}
	return getAPIVolumeList(volumes)
}

// LogkeeperBuildMetadata is the resolver for the logkeeperBuildMetadata field.
func (r *queryResolver) LogkeeperBuildMetadata(ctx context.Context, buildID string) (*plank.Build, error) {
	client := plank.NewLogkeeperClient(plank.NewLogkeeperClientOptions{
		BaseURL: evergreen.GetEnvironment().Settings().LoggerConfig.LogkeeperURL,
	})
	build, err := client.GetBuildMetadata(ctx, buildID)
	if err != nil {
		return nil, InternalServerError.Send(ctx, err.Error())
	}
	return &build, nil
}

// Task is the resolver for the task field.
func (r *queryResolver) Task(ctx context.Context, taskID string, execution *int) (*restModel.APITask, error) {
	return getTask(ctx, taskID, execution, r.sc.GetURL())
}

// TaskAllExecutions is the resolver for the taskAllExecutions field.
func (r *queryResolver) TaskAllExecutions(ctx context.Context, taskID string) ([]*restModel.APITask, error) {
	latestTask, err := task.FindOneId(taskID)
	if err != nil {
		return nil, ResourceNotFound.Send(ctx, err.Error())
	}
	if latestTask == nil {
		return nil, werrors.Errorf("unable to find task %s", taskID)
	}
	allTasks := []*restModel.APITask{}
	for i := 0; i < latestTask.Execution; i++ {
		var dbTask *task.Task
		dbTask, err = task.FindByIdExecution(taskID, &i)
		if err != nil {
			return nil, ResourceNotFound.Send(ctx, err.Error())
		}
		if dbTask == nil {
			return nil, werrors.Errorf("unable to find task %s", taskID)
		}
		var apiTask *restModel.APITask
		apiTask, err = getAPITaskFromTask(ctx, r.sc.GetURL(), *dbTask)
		if err != nil {
			return nil, InternalServerError.Send(ctx, "error converting task")
		}
		allTasks = append(allTasks, apiTask)
	}
	apiTask, err := getAPITaskFromTask(ctx, r.sc.GetURL(), *latestTask)
	if err != nil {
		return nil, InternalServerError.Send(ctx, "error converting task")
	}
	allTasks = append(allTasks, apiTask)
	return allTasks, nil
}

// TaskTests is the resolver for the taskTests field.
func (r *queryResolver) TaskTests(ctx context.Context, taskID string, execution *int, sortCategory *TestSortCategory, sortDirection *SortDirection, page *int, limit *int, testName *string, statuses []string, groupID *string) (*TaskTestResult, error) {
	dbTask, err := task.FindByIdExecution(taskID, execution)
	if dbTask == nil || err != nil {
		return nil, ResourceNotFound.Send(ctx, fmt.Sprintf("finding task with id '%s' and execution %d", taskID, utility.FromIntPtr(execution)))
	}
	baseTaskOpts, err := getBaseTaskTestResultsOptions(ctx, dbTask)
	if err != nil {
		return nil, err
	}

	var sortBy string
	if sortCategory != nil {
		switch *sortCategory {
		case TestSortCategoryStatus:
			sortBy = testresult.SortByStatus
		case TestSortCategoryDuration:
			sortBy = testresult.SortByDuration
		case TestSortCategoryTestName:
			sortBy = testresult.SortByTestName
		case TestSortCategoryStartTime:
			sortBy = testresult.SortByStart
		case TestSortCategoryBaseStatus:
			if len(baseTaskOpts) > 0 {
				// Only sort by base status if we know there
				// are base task options we can send to the
				// results service.
				sortBy = testresult.SortByBaseStatus
			}
		}
	}

	taskResults, err := dbTask.GetTestResults(
		ctx,
		evergreen.GetEnvironment(),
		&testresult.FilterOptions{
			TestName:     utility.FromStringPtr(testName),
			Statuses:     statuses,
			GroupID:      utility.FromStringPtr(groupID),
			SortBy:       sortBy,
			SortOrderDSC: sortDirection != nil && *sortDirection == SortDirectionDesc,
			Limit:        utility.FromIntPtr(limit),
			Page:         utility.FromIntPtr(page),
			BaseTasks:    baseTaskOpts,
		},
	)
	if err != nil {
		return nil, InternalServerError.Send(ctx, fmt.Sprintf("finding test results for task '%s': %s", taskID, err))
	}

	apiResults := make([]*restModel.APITest, len(taskResults.Results))
	for i, t := range taskResults.Results {
		apiTest := &restModel.APITest{}
		if err = apiTest.BuildFromService(t.TaskID); err != nil {
			return nil, InternalServerError.Send(ctx, err.Error())
		}
		if err = apiTest.BuildFromService(&t); err != nil {
			return nil, InternalServerError.Send(ctx, err.Error())
		}

		apiResults[i] = apiTest
	}

	return &TaskTestResult{
		TestResults:       apiResults,
		TotalTestCount:    taskResults.Stats.TotalCount,
		FilteredTestCount: utility.FromIntPtr(taskResults.Stats.FilteredCount),
	}, nil
}

// TaskTestSample is the resolver for the taskTestSample field.
func (r *queryResolver) TaskTestSample(ctx context.Context, tasks []string, filters []*TestFilter) ([]*TaskTestResultSample, error) {
	if len(tasks) == 0 {
		return nil, nil
	}
	dbTasks, err := task.FindAll(db.Query(task.ByIds(tasks)))
	if err != nil {
		return nil, InternalServerError.Send(ctx, fmt.Sprintf("Error finding tasks %s: %s", tasks, err))
	}
	if len(dbTasks) == 0 {
		return nil, ResourceNotFound.Send(ctx, fmt.Sprintf("Tasks %s not found", tasks))
	}

	failingTests := []string{}
	for _, f := range filters {
		failingTests = append(failingTests, f.TestName)
	}

	var allTaskOpts []testresult.TaskOptions
	apiSamples := make([]*TaskTestResultSample, len(dbTasks))
	apiSamplesByTaskID := map[string]*TaskTestResultSample{}
	for i, dbTask := range dbTasks {
		taskOpts, err := dbTask.CreateTestResultsTaskOptions()
		if err != nil {
			return nil, InternalServerError.Send(ctx, fmt.Sprintf("Error creating test results task options for task '%s': %s", dbTask.Id, err))
		}

		apiSamples[i] = &TaskTestResultSample{TaskID: dbTask.Id, Execution: dbTask.Execution}
		for _, o := range taskOpts {
			apiSamplesByTaskID[o.TaskID] = apiSamples[i]
		}
		allTaskOpts = append(allTaskOpts, taskOpts...)
	}

	if len(allTaskOpts) > 0 {
		samples, err := testresult.GetFailedTestSamples(ctx, evergreen.GetEnvironment(), allTaskOpts, failingTests)
		if err != nil {
			return nil, InternalServerError.Send(ctx, fmt.Sprintf("Error getting test results sample: %s", err))
		}

		for _, sample := range samples {
			apiSample, ok := apiSamplesByTaskID[sample.TaskID]
			if !ok {
				return nil, InternalServerError.Send(ctx, fmt.Sprintf("Error: unexpected task '%s' in task test sample result", sample.TaskID))
			}

			apiSample.MatchingFailedTestNames = append(apiSample.MatchingFailedTestNames, sample.MatchingFailedTestNames...)
			apiSample.TotalTestCount += sample.TotalFailedNames
		}
	}

	return apiSamples, nil
}

// MyPublicKeys is the resolver for the myPublicKeys field.
func (r *queryResolver) MyPublicKeys(ctx context.Context) ([]*restModel.APIPubKey, error) {
	publicKeys := getMyPublicKeys(ctx)
	return publicKeys, nil
}

// User is the resolver for the user field.
func (r *queryResolver) User(ctx context.Context, userID *string) (*restModel.APIDBUser, error) {
	usr := mustHaveUser(ctx)
	var err error
	if userID != nil {
		usr, err = user.FindOneById(*userID)
		if err != nil {
			return nil, ResourceNotFound.Send(ctx, fmt.Sprintf("Error getting user from user ID: %s", err.Error()))
		}
		if usr == nil {
			return nil, ResourceNotFound.Send(ctx, "Could not find user from user ID")
		}
	}
	displayName := usr.DisplayName()
	username := usr.Username()
	email := usr.Email()
	user := restModel.APIDBUser{
		DisplayName:  &displayName,
		UserID:       &username,
		EmailAddress: &email,
	}
	return &user, nil
}

// UserConfig is the resolver for the userConfig field.
func (r *queryResolver) UserConfig(ctx context.Context) (*UserConfig, error) {
	usr := mustHaveUser(ctx)
	settings := evergreen.GetEnvironment().Settings()
	config := &UserConfig{
		User:          usr.Username(),
		APIKey:        usr.GetAPIKey(),
		UIServerHost:  settings.Ui.Url,
		APIServerHost: settings.ApiUrl + "/api",
	}
	return config, nil
}

// UserSettings is the resolver for the userSettings field.
func (r *queryResolver) UserSettings(ctx context.Context) (*restModel.APIUserSettings, error) {
	usr := mustHaveUser(ctx)
	userSettings := restModel.APIUserSettings{}
	userSettings.BuildFromService(usr.Settings)
	return &userSettings, nil
}

// CommitQueue is the resolver for the commitQueue field.
func (r *queryResolver) CommitQueue(ctx context.Context, projectIdentifier string) (*restModel.APICommitQueue, error) {
	commitQueue, err := data.FindCommitQueueForProject(projectIdentifier)
	if err != nil {
		if werrors.Cause(err) == err {
			return nil, ResourceNotFound.Send(ctx, fmt.Sprintf("error finding commit queue for %s: %s", projectIdentifier, err.Error()))
		}
		return nil, InternalServerError.Send(ctx, fmt.Sprintf("error finding commit queue for %s: %s", projectIdentifier, err.Error()))
	}
	project, err := data.FindProjectById(projectIdentifier, true, true)
	if err != nil {
		return nil, InternalServerError.Send(ctx, fmt.Sprintf("error finding project %s: %s", projectIdentifier, err.Error()))
	}
	if project.CommitQueue.Message != "" {
		commitQueue.Message = &project.CommitQueue.Message
	}
	commitQueue.Owner = &project.Owner
	commitQueue.Repo = &project.Repo

	for i, item := range commitQueue.Queue {
		patchId := ""
		if utility.FromStringPtr(item.Version) != "" {
			patchId = utility.FromStringPtr(item.Version)
		} else if utility.FromStringPtr(item.PatchId) != "" {
			patchId = utility.FromStringPtr(item.PatchId)
		}

		if patchId != "" {
			p, err := data.FindPatchById(patchId)
			if err != nil {
				return nil, ResourceNotFound.Send(ctx, fmt.Sprintf("error finding patch: %s", err.Error()))
			}
			commitQueue.Queue[i].Patch = p
		}
	}
	return commitQueue, nil
}

// BuildVariantsForTaskName is the resolver for the buildVariantsForTaskName field.
func (r *queryResolver) BuildVariantsForTaskName(ctx context.Context, projectIdentifier string, taskName string) ([]*task.BuildVariantTuple, error) {
	pid, err := model.GetIdForProject(projectIdentifier)
	if err != nil {
		return nil, ResourceNotFound.Send(ctx, fmt.Sprintf("Could not find project with id: %s", projectIdentifier))
	}
	repo, err := model.FindRepository(pid)
	if err != nil {
		return nil, InternalServerError.Send(ctx, fmt.Sprintf("getting repository for '%s': %s", pid, err.Error()))
	}
	if repo == nil {
		return nil, ResourceNotFound.Send(ctx, fmt.Sprintf("could not find repository '%s'", pid))
	}
	taskBuildVariants, err := task.FindUniqueBuildVariantNamesByTask(pid, taskName, repo.RevisionOrderNumber, false)
	if err != nil {
		return nil, InternalServerError.Send(ctx, fmt.Sprintf("getting build variant tasks for task '%s': %s", taskName, err.Error()))
	}
	// use legacy lookup pipeline if no results are found
	if len(taskBuildVariants) == 0 {
		taskBuildVariants, err = task.FindUniqueBuildVariantNamesByTask(pid, taskName, repo.RevisionOrderNumber, true)
		if err != nil {
			return nil, InternalServerError.Send(ctx, fmt.Sprintf("getting build variant tasks using legacy pipeline for task '%s': %s", taskName, err.Error()))
		}
	}
	if taskBuildVariants == nil {
		return nil, nil
	}
	return taskBuildVariants, nil
}

// MainlineCommits is the resolver for the mainlineCommits field.
func (r *queryResolver) MainlineCommits(ctx context.Context, options MainlineCommitsOptions, buildVariantOptions *BuildVariantOptions) (*MainlineCommits, error) {
	projectId, err := model.GetIdForProject(options.ProjectIdentifier)
	if err != nil {
		return nil, ResourceNotFound.Send(ctx, fmt.Sprintf("Could not find project with id: %s", options.ProjectIdentifier))
	}
	limit := model.DefaultMainlineCommitVersionLimit
	if utility.FromIntPtr(options.Limit) != 0 {
		limit = utility.FromIntPtr(options.Limit)
	}
	requesters := options.Requesters
	if len(requesters) == 0 {
		requesters = evergreen.SystemVersionRequesterTypes
	}
	opts := model.MainlineCommitVersionOptions{
		Limit:           limit,
		SkipOrderNumber: utility.FromIntPtr(options.SkipOrderNumber),
		Requesters:      requesters,
	}

	versions, err := model.GetMainlineCommitVersionsWithOptions(ctx, projectId, opts)
	if err != nil {
		return nil, ResourceNotFound.Send(ctx, fmt.Sprintf("Error getting activated versions: %s", err.Error()))
	}

	var mainlineCommits MainlineCommits
	matchingVersionCount := 0

	// We only want to return the PrevPageOrderNumber if a user is not on the first page
	if options.SkipOrderNumber != nil {
		prevPageCommit, err := model.GetPreviousPageCommitOrderNumber(ctx, projectId, utility.FromIntPtr(options.SkipOrderNumber), limit, requesters)

		if err != nil {
			// This shouldn't really happen, but if it does, we should return an error and log it
			grip.Warning(message.WrapError(err, message.Fields{
				"message":    "Error getting most recent version",
				"project_id": projectId,
			}))
			return nil, InternalServerError.Send(ctx, fmt.Sprintf("Error getting most recent mainline commit: %s", err.Error()))
		}

		if prevPageCommit != nil {
			mainlineCommits.PrevPageOrderNumber = prevPageCommit
		}
	}

	index := 0
	versionsCheckedCount := 0

	// We will loop through each version returned from GetMainlineCommitVersionsWithOptions and see if there is a commit that matches the filter parameters (if any).
	// If there is a match, we will add it to the array of versions to be returned to the user.
	// If there are not enough matches to satisfy our limit, we will call GetMainlineCommitVersionsWithOptions again with the next order number to check and repeat the process.
	for matchingVersionCount < limit {
		// If we no longer have any more versions to check break out and return what we have.
		if len(versions) == 0 {
			break
		}
		// If we have checked more versions than the MaxMainlineCommitVersionLimit then break out and return what we have.
		if versionsCheckedCount >= model.MaxMainlineCommitVersionLimit {
			// Return an error if we did not find any versions that match.
			if matchingVersionCount == 0 {
				return nil, ResourceNotFound.Send(ctx, fmt.Sprintf("Matching version not found in %d most recent versions", model.MaxMainlineCommitVersionLimit))
			}
			break
		}
		versionsCheckedCount++
		v := versions[index]
		apiVersion := restModel.APIVersion{}
		apiVersion.BuildFromService(v)

		// If the version was created before we started caching activation status we must manually verify it and cache that value.
		if v.Activated == nil {
			err = setVersionActivationStatus(ctx, &v)
			if err != nil {
				return nil, InternalServerError.Send(ctx, fmt.Sprintf("Error setting version activation status: %s", err.Error()))
			}
		}
		mainlineCommitVersion := MainlineCommitVersion{}
		shouldCollapse := false
		if !utility.FromBoolPtr(v.Activated) {
			shouldCollapse = true
		} else if isPopulated(buildVariantOptions) && utility.FromBoolPtr(options.ShouldCollapse) {
			opts := task.HasMatchingTasksOptions{
				TaskNames:                  buildVariantOptions.Tasks,
				Variants:                   buildVariantOptions.Variants,
				Statuses:                   getValidTaskStatusesFilter(buildVariantOptions.Statuses),
				IncludeNeverActivatedTasks: true,
			}
			hasTasks, err := task.HasMatchingTasks(ctx, v.Id, opts)
			if err != nil {
				return nil, InternalServerError.Send(ctx, fmt.Sprintf("Error checking if version has tasks: %s", err.Error()))
			}
			if !hasTasks {
				shouldCollapse = true
			}
		}
		// If a version matches our filter criteria we append it directly to our returned list of mainlineCommits
		if !shouldCollapse {
			matchingVersionCount++
			mainlineCommits.NextPageOrderNumber = utility.ToIntPtr(v.RevisionOrderNumber)
			mainlineCommitVersion.Version = &apiVersion

		} else {
			// If a version does not match our filter criteria roll up all the unactivated versions that are sequentially near each other into a single MainlineCommitVersion,
			// and then append it to our returned list.
			// If we have any versions already we should check the most recent one first otherwise create a new one
			if len(mainlineCommits.Versions) > 0 {
				lastMainlineCommit := mainlineCommits.Versions[len(mainlineCommits.Versions)-1]

				// If the previous mainlineCommit contains rolled up unactivated versions append the latest RolledUp unactivated version
				if lastMainlineCommit.RolledUpVersions != nil {
					lastMainlineCommit.RolledUpVersions = append(lastMainlineCommit.RolledUpVersions, &apiVersion)
				} else {
					mainlineCommitVersion.RolledUpVersions = []*restModel.APIVersion{&apiVersion}
				}
			} else {
				mainlineCommitVersion.RolledUpVersions = []*restModel.APIVersion{&apiVersion}
			}
		}

		// Only add a mainlineCommit if a new one was added and it's not a modified existing RolledUpVersion
		if mainlineCommitVersion.Version != nil || mainlineCommitVersion.RolledUpVersions != nil {
			mainlineCommits.Versions = append(mainlineCommits.Versions, &mainlineCommitVersion)
		}
		index++
		// If we have exhausted all of our versions we should fetch some more.
		if index == len(versions) && matchingVersionCount < limit {
			skipOrderNumber := versions[len(versions)-1].RevisionOrderNumber
			opts := model.MainlineCommitVersionOptions{
				Limit:           limit,
				SkipOrderNumber: skipOrderNumber,
				Requesters:      requesters,
			}

			versions, err = model.GetMainlineCommitVersionsWithOptions(ctx, projectId, opts)
			if err != nil {
				return nil, ResourceNotFound.Send(ctx, fmt.Sprintf("Error getting activated versions: %s", err.Error()))
			}
			index = 0
		}
	}
	return &mainlineCommits, nil
}

// TaskNamesForBuildVariant is the resolver for the taskNamesForBuildVariant field.
func (r *queryResolver) TaskNamesForBuildVariant(ctx context.Context, projectIdentifier string, buildVariant string) ([]string, error) {
	pid, err := model.GetIdForProject(projectIdentifier)
	if err != nil {
		return nil, ResourceNotFound.Send(ctx, fmt.Sprintf("Could not find project with id: %s", projectIdentifier))
	}
	repo, err := model.FindRepository(pid)
	if err != nil {
		return nil, InternalServerError.Send(ctx, fmt.Sprintf("Error while getting repository for '%s': %s", pid, err.Error()))
	}
	if repo == nil {
		return nil, ResourceNotFound.Send(ctx, fmt.Sprintf("could not find repository '%s'", pid))
	}
	buildVariantTasks, err := task.FindTaskNamesByBuildVariant(pid, buildVariant, repo.RevisionOrderNumber)
	if err != nil {
		return nil, InternalServerError.Send(ctx, fmt.Sprintf("Error while getting tasks for '%s': %s", buildVariant, err.Error()))
	}
	if buildVariantTasks == nil {
		return []string{}, nil
	}
	return buildVariantTasks, nil
}

// HasVersion is the resolver for the hasVersion field.
func (r *queryResolver) HasVersion(ctx context.Context, id string) (bool, error) {
	v, err := model.VersionFindOne(model.VersionById(id))
	if err != nil {
		return false, InternalServerError.Send(ctx, fmt.Sprintf("Error finding version %s: %s", id, err.Error()))
	}
	if v != nil {
		return true, nil
	}

	if patch.IsValidId(id) {
		p, err := patch.FindOneId(id)
		if err != nil {
			return false, InternalServerError.Send(ctx, fmt.Sprintf("Error finding patch %s: %s", id, err.Error()))
		}
		if p != nil {
			return false, nil
		}
	}
	return false, ResourceNotFound.Send(ctx, fmt.Sprintf("Unable to find patch or version %s", id))
}

// Version is the resolver for the version field.
func (r *queryResolver) Version(ctx context.Context, id string) (*restModel.APIVersion, error) {
	v, err := model.VersionFindOneId(id)
	if err != nil {
		return nil, InternalServerError.Send(ctx, fmt.Sprintf("Error while finding version with id: `%s`: %s", id, err.Error()))
	}
	if v == nil {
		return nil, ResourceNotFound.Send(ctx, fmt.Sprintf("Unable to find version with id: `%s`", id))
	}
	apiVersion := restModel.APIVersion{}
	apiVersion.BuildFromService(*v)
	return &apiVersion, nil
}

// Query returns QueryResolver implementation.
func (r *Resolver) Query() QueryResolver { return &queryResolver{r} }

type queryResolver struct{ *Resolver }
