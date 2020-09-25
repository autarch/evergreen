// Code generated by rest/model/codegen.go. DO NOT EDIT.

package model

import "github.com/evergreen-ci/evergreen/model/patch"

type APIDisplayTask struct {
	Name      string    `json:"name"`
	ExecTasks []*string `json:"exec_tasks"`
}

// APIDisplayTaskBuildFromService takes the patch.DisplayTask DB struct and
// returns the REST struct *APIDisplayTask with the corresponding fields populated
func APIDisplayTaskBuildFromService(t patch.DisplayTask) *APIDisplayTask {
	m := APIDisplayTask{}
	m.ExecTasks = ArrstringArrstringPtr(t.ExecTasks)
	m.Name = StringString(t.Name)
	return &m
}

// APIDisplayTaskToService takes the APIDisplayTask REST struct and returns the DB struct
// *patch.DisplayTask with the corresponding fields populated
func APIDisplayTaskToService(m APIDisplayTask) *patch.DisplayTask {
	out := &patch.DisplayTask{}
	out.ExecTasks = ArrstringPtrArrstring(m.ExecTasks)
	out.Name = StringString(m.Name)
	return out
}
