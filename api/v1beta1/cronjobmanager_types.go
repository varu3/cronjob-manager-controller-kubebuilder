/*
Copyright 2021 varu3.

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

package v1beta1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// CronjobManagerSpec defines the desired state of CronjobManager
type CronJobManagerSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of CronjobManager. Edit cronjobmanager_types.go to remove/update
	Image    string            `json:"image,omitempty"`
	TZ       string            `json:"tz,omitempty"`
	Cronjobs []CronJobSettings `json:"cronjobSettings,omitempty"`
}

type CronJobSettings struct {
	Name     string   `json:"name,omitempty"`
	Schedule string   `json:"schedule"`
	Command  []string `json:"command"`
	Type     string   `json:"type"`
}

// CronjobManagerStatus defines the observed state of CronjobManager
type CronJobManagerStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// CronjobManager is the Schema for the cronjobmanagers API
type CronJobManager struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CronJobManagerSpec   `json:"spec,omitempty"`
	Status CronJobManagerStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// CronJobManagerList contains a list of CronJobManager
type CronJobManagerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CronJobManager `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CronJobManager{}, &CronJobManagerList{})
}
