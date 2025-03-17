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
	"testing"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	batch "volcano.sh/apis/pkg/apis/batch/v1alpha1"
)

func TestGetJobNameFunc(t *testing.T) {
	type args struct {
		jobFlowName     string
		jobTemplateName string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "GetJobName success case",
			args: args{
				jobFlowName:     "jobFlowA",
				jobTemplateName: "jobTemplateA",
			},
			want: "jobFlowA-jobTemplateA",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getJobName(tt.args.jobFlowName, tt.args.jobTemplateName); got != tt.want {
				t.Errorf("getJobName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetConnectionOfJobAndJobTemplate(t *testing.T) {
	type args struct {
		namespace string
		name      string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "TestGetConnectionOfJobAndJobTemplate",
			args: args{
				namespace: "default",
				name:      "flow",
			},
			want: "default.flow",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GenerateObjectString(tt.args.namespace, tt.args.name); got != tt.want {
				t.Errorf("GenerateObjectString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetJobFlowNameByJob(t *testing.T) {
	type args struct {
		job *batch.Job
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "TestGetConnectionOfJobAndJobTemplate",
			args: args{
				job: &batch.Job{
					TypeMeta: v1.TypeMeta{},
					ObjectMeta: v1.ObjectMeta{
						OwnerReferences: []v1.OwnerReference{
							{
								APIVersion:         "flow.volcano.sh/v1alpha1",
								Kind:               JobFlow,
								Name:               "jobflowtest",
								UID:                "",
								Controller:         nil,
								BlockOwnerDeletion: nil,
							},
						},
					},
					Spec:   batch.JobSpec{},
					Status: batch.JobStatus{},
				},
			},
			want: "jobflowtest",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getJobFlowNameByJob(tt.args.job); got != tt.want {
				t.Errorf("getJobFlowNameByJob() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPatchVolumes(t *testing.T) {
	type args struct {
		volumes []batch.VolumeSpec
		patchs  []batch.VolumeSpec
	}
	tests := []struct {
		name string
		args args
		want []batch.VolumeSpec
	}{
		{
			name: "TestPatchInVolumes",
			args: args{
				volumes: []batch.VolumeSpec{
					{
						MountPath:       "/aa/aa",
						VolumeClaimName: "aa",
					},
					{
						MountPath:       "/aa/bb",
						VolumeClaimName: "bb",
					},
				},
				patchs: []batch.VolumeSpec{
					{
						MountPath:       "/aa/aa",
						VolumeClaimName: "cc",
					},
				},
			},
			want: []batch.VolumeSpec{
				{
					MountPath:       "/aa/aa",
					VolumeClaimName: "cc",
				},
				{
					MountPath:       "/aa/bb",
					VolumeClaimName: "bb",
				},
			},
		},
		{
			name: "TestPatchNotInVolumes",
			args: args{
				volumes: []batch.VolumeSpec{
					{
						MountPath:       "/aa/aa",
						VolumeClaimName: "aa",
					},
					{
						MountPath:       "/aa/bb",
						VolumeClaimName: "bb",
					},
				},
				patchs: []batch.VolumeSpec{
					{
						MountPath:       "/aa/cc",
						VolumeClaimName: "cc",
					},
				},
			},
			want: []batch.VolumeSpec{
				{
					MountPath:       "/aa/aa",
					VolumeClaimName: "aa",
				},
				{
					MountPath:       "/aa/bb",
					VolumeClaimName: "bb",
				},
				{
					MountPath:       "/aa/cc",
					VolumeClaimName: "cc",
				},
			},
		},
		{
			name: "TestPatchNotInVolumesTwo",
			args: args{
				volumes: []batch.VolumeSpec{},
				patchs: []batch.VolumeSpec{
					{
						MountPath:       "/aa/cc",
						VolumeClaimName: "cc",
					},
				},
			},
			want: []batch.VolumeSpec{
				{
					MountPath:       "/aa/cc",
					VolumeClaimName: "cc",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := patchVolumes(tt.args.volumes, tt.args.patchs)
			if !equality.Semantic.DeepEqual(got, tt.want) {
				t.Errorf("PatchVolumes() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPatchTasks(t *testing.T) {
	type args struct {
		tasks  []batch.TaskSpec
		patchs []batch.TaskSpec
	}
	tests := []struct {
		name string
		args args
		want []batch.TaskSpec
		err  bool
	}{
		{
			name: "TestPatchInVolumes",
			args: args{
				tasks: []batch.TaskSpec{
					{
						Name:     "aa",
						Replicas: 3,
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Volumes: []corev1.Volume{
									{
										Name: "aaaa",
									},
								},
								Containers: []corev1.Container{
									{
										Name:  "container",
										Image: "nginx:latest1",
										VolumeMounts: []corev1.VolumeMount{
											{
												Name:      "aaaa",
												MountPath: "/aa/aa/",
											},
										},
									},
								},
							},
						},
					},
					{
						Name:     "bb",
						Replicas: 4,
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Volumes: []corev1.Volume{
									{
										Name: "bbbb",
									},
								},
								Containers: []corev1.Container{
									{
										Name:  "container",
										Image: "nginx:latest2",
										VolumeMounts: []corev1.VolumeMount{
											{
												Name:      "bbbb",
												MountPath: "/bb/bb/",
											},
										},
									},
								},
							},
						},
					},
				},
				patchs: []batch.TaskSpec{
					{
						Name:     "bb",
						Replicas: 2,
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Volumes: []corev1.Volume{
									{
										Name: "bbbb",
									},
								},
								Containers: []corev1.Container{
									{
										Name:  "container",
										Image: "nginx:latest2",
										VolumeMounts: []corev1.VolumeMount{
											{
												Name:      "bbbbc",
												MountPath: "/bb/bb/cc/",
											},
										},
									},
								},
							},
						},
					},
				},
			},
			want: []batch.TaskSpec{
				{
					Name:     "aa",
					Replicas: 3,
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Volumes: []corev1.Volume{
								{
									Name: "aaaa",
								},
							},
							Containers: []corev1.Container{
								{
									Name:  "container",
									Image: "nginx:latest1",
									VolumeMounts: []corev1.VolumeMount{
										{
											Name:      "aaaa",
											MountPath: "/aa/aa/",
										},
									},
								},
							},
						},
					},
				},
				{
					Name:     "bb",
					Replicas: 2,
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Volumes: []corev1.Volume{
								{
									Name: "bbbb",
								},
							},
							Containers: []corev1.Container{
								{
									Name:  "container",
									Image: "nginx:latest2",
									VolumeMounts: []corev1.VolumeMount{
										{
											Name:      "bbbb",
											MountPath: "/bb/bb/",
										},
										{
											Name:      "bbbbc",
											MountPath: "/bb/bb/cc/",
										},
									},
								},
							},
						},
					},
				},
			},
			err: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := patchTasks(tt.args.tasks, tt.args.patchs)
			if (err == nil) == tt.err {
				t.Errorf("patchTasks() error = %v", err)
			}
			if !equality.Semantic.DeepEqual(got, tt.want) {
				t.Errorf("patchTasks() = %v, want %v", got, tt.want)
			}
		})
	}
}
