/*
Copyright 2018 The Volcano Authors.

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

package validate

import (
	"fmt"

	admissionv1 "k8s.io/api/admission/v1"
	whv1 "k8s.io/api/admissionregistration/v1"
	"k8s.io/klog/v2"
	"k8s.io/utils/ptr"

	jobflowv1alpha1 "volcano.sh/apis/pkg/apis/flow/v1alpha1"
	"volcano.sh/volcano/pkg/webhooks/router"
	"volcano.sh/volcano/pkg/webhooks/schema"
	"volcano.sh/volcano/pkg/webhooks/util"
)

const (
	// DefaulGlobaltMaxRetry is the default number of retries.
	DefaulGlobaltMaxRetry = 3
)

func init() {
	router.RegisterAdmission(service)
}

var service = &router.AdmissionService{
	Path: "/jobflows/validate",
	Func: AdmitJobFlow,

	Config: config,

	ValidatingConfig: &whv1.ValidatingWebhookConfiguration{
		Webhooks: []whv1.ValidatingWebhook{{
			Name: "validatejobflow.volcano.sh",
			Rules: []whv1.RuleWithOperations{
				{
					Operations: []whv1.OperationType{whv1.Create, whv1.Update, whv1.Delete},
					Rule: whv1.Rule{
						APIGroups:   []string{jobflowv1alpha1.SchemeGroupVersion.Group},
						APIVersions: []string{jobflowv1alpha1.SchemeGroupVersion.Version},
						Resources:   []string{"jobflows"},
					},
				},
			},
		}},
	},
}

var config = &router.AdmissionServiceConfig{}

// AdmitJobFlow is to admit jobflow and return response.
func AdmitJobFlow(ar admissionv1.AdmissionReview) *admissionv1.AdmissionResponse {
	klog.V(3).Infof("Admitting %s jobflw %s.", ar.Request.Operation, ar.Request.Name)

	jobflow, err := schema.DecodeJobFlow(ar.Request.Object, ar.Request.Resource)
	if err != nil {
		return util.ToAdmissionResponse(err)
	}

	switch ar.Request.Operation {
	case admissionv1.Create, admissionv1.Update:
		err := validateJobFlow(jobflow)
		if err != nil {
			return util.ToAdmissionResponse(err)
		}
	default:
		return util.ToAdmissionResponse(fmt.Errorf("invalid operation `%s`, "+
			"expect operation to be `CREATE` or `UPDATE`", ar.Request.Operation))
	}

	return &admissionv1.AdmissionResponse{
		Allowed: true,
	}
}

func validateJobFlow(jobflow *jobflowv1alpha1.JobFlow) error {
	if jobflow == nil {
		return nil
	}

	if ptr.Deref(jobflow.Spec.MaxRetry, 0) <= 0 {
		return fmt.Errorf("'maxRetry' cannot be less than zero.")
	}

	return validateDAGNoCycles(jobflow)
}

func validateDAGNoCycles(jobflow *jobflowv1alpha1.JobFlow) error {
	if jobflow == nil || len(jobflow.Spec.Flows) <= 0 {
		return nil
	}

	visited := make(map[string]bool)
	recStack := make(map[string]bool)

	var visit func(string) bool

	visit = func(name string) bool {
		if recStack[name] {
			return true // cycle detected
		}
		if visited[name] {
			return false
		}
		visited[name] = true
		recStack[name] = true

		for _, task := range jobflow.Spec.Flows {
			if task.Name == name {
				for _, target := range task.DependsOn.Targets {
					if visit(target) {
						return true
					}
				}
			}
		}
		recStack[name] = false
		return false
	}

	for _, task := range jobflow.Spec.Flows {
		if visit(task.Name) {
			return fmt.Errorf("jobflow flow cyclic dependency detected")
		}
	}
	return nil
}
