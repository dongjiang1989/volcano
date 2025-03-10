/*
Copyright 2022 The Volcano Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package jobflow

import (
	"fmt"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/klog/v2"

	batch "volcano.sh/apis/pkg/apis/batch/v1alpha1"
	v1alpha1flow "volcano.sh/apis/pkg/apis/flow/v1alpha1"
)

func getJobName(jobFlowName string, jobTemplateName string) string {
	return jobFlowName + "-" + jobTemplateName
}

// GenerateObjectString generates the object information string using namespace and name
func GenerateObjectString(namespace, name string) string {
	return namespace + "." + name
}

func isControlledBy(obj metav1.Object, gvk schema.GroupVersionKind) bool {
	controllerRef := metav1.GetControllerOf(obj)
	if controllerRef == nil {
		return false
	}
	if controllerRef.APIVersion == gvk.GroupVersion().String() && controllerRef.Kind == gvk.Kind {
		return true
	}
	return false
}

func getJobFlowNameByJob(job *batch.Job) string {
	for _, owner := range job.OwnerReferences {
		if owner.Kind == JobFlow && strings.Contains(owner.APIVersion, Volcano) {
			return owner.Name
		}
	}
	return ""
}

func getFlowsPatchByName(jobFlow *v1alpha1flow.JobFlow, name string) (*v1alpha1flow.Patch, error) {
	for _, fw := range jobFlow.Spec.Flows {
		if fw.Name == name {
			return fw.Patch, nil
		}
	}

	return nil, fmt.Errorf("not found flows patch.")
}

func patchObjects(job, patch batch.JobSpec) (batch.JobSpec, error) {
	if patch.SchedulerName != "" {
		job.SchedulerName = patch.SchedulerName
	}

	if patch.MinAvailable > 0 {
		job.MinAvailable = patch.MinAvailable
	}

	if len(patch.Tasks) > 0 {
		tasks, err := patchTasks(job.Tasks, patch.Tasks)
		if err != nil {
			return batch.JobSpec{}, err
		}
		job.Tasks = tasks
	}

	if len(patch.Policies) > 0 {
		job.Policies = patch.Policies
	}

	if patch.RunningEstimate != nil {
		job.RunningEstimate = patch.RunningEstimate
	}

	if patch.Queue != "" {
		job.Queue = patch.Queue
	}

	if patch.MaxRetry != job.MaxRetry {
		job.MaxRetry = patch.MaxRetry
	}

	if patch.TTLSecondsAfterFinished != nil {
		job.TTLSecondsAfterFinished = patch.TTLSecondsAfterFinished
	}

	if patch.PriorityClassName != "" {
		job.PriorityClassName = patch.PriorityClassName
	}

	if patch.MinSuccess != nil {
		job.MinSuccess = patch.MinSuccess
	}

	return job, nil
}

func patchTasks(tasks, patchs []batch.TaskSpec) ([]batch.TaskSpec, error) {
	ret := make([]batch.TaskSpec, len(tasks))
	if len(patchs) > len(tasks) {
		klog.Errorf("patch.spec.task not in job template taskSpec: patchs %v, tasks %v",
			patchs, tasks)
		return ret, fmt.Errorf("patch.spec.task not in job template taskSpec: patchs %v, tasks %v", patchs, tasks)
	}

	for index, task := range tasks {
		ret[index] = *task.DeepCopy()
	}

	for index, patch := range patchs {
		if patch.Name != "" {
			ret[index].Name = patch.Name
		}

		if patch.Replicas > 0 {
			ret[index].Replicas = patch.Replicas
		}

		// container
		if len(patch.Template.Spec.Containers) > 0 {
			for i, container := range patch.Template.Spec.Containers {
				if container.Image != "" {
					ret[index].Template.Spec.Containers[i].Image = container.Image
				}

				if len(container.Command) > 0 {
					ret[index].Template.Spec.Containers[i].Command = container.Command
				}

				if len(container.Args) > 0 {
					ret[index].Template.Spec.Containers[i].Args = container.Args
				}

				if len(container.Env) > 0 {
					ret[index].Template.Spec.Containers[i].Env = container.Env
				}
			}

		}

		// init container
		if len(patch.Template.Spec.InitContainers) > 0 {
			for i, container := range patch.Template.Spec.InitContainers {
				if container.Image != "" {
					ret[index].Template.Spec.InitContainers[i].Image = container.Image
				}

				if len(container.Command) > 0 {
					ret[index].Template.Spec.Containers[i].Command = container.Command
				}

				if len(container.Args) > 0 {
					ret[index].Template.Spec.Containers[i].Args = container.Args
				}

				if len(container.Env) > 0 {
					ret[index].Template.Spec.Containers[i].Env = container.Env
				}

			}
		}

		if patch.MinAvailable != nil {
			ret[index].MinAvailable = patch.MinAvailable
		}

		if len(patch.Policies) > 0 {
			ret[index].Policies = patch.Policies
		}

		if patch.MaxRetry > 0 {
			ret[index].MaxRetry = patch.MaxRetry
		}

		if patch.DependsOn != nil {
			ret[index].DependsOn = patch.DependsOn
		}
	}

	return ret, nil
}
