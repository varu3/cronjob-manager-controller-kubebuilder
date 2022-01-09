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

package controllers

import (
	"context"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	cronjobmanagerv1beta1 "github.com/varu3/cronjob-manager-controller-kubebuilder/api/v1beta1"
	batchv1 "k8s.io/api/batch/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CronjobManagerReconciler reconciles a CronjobManager object
type CronJobManagerReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

//+kubebuilder:rbac:groups=cronjobmanager.varu3.me,resources=cronjobmanagers,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=cronjobmanager.varu3.me,resources=cronjobmanagers/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=cronjobmanager.varu3.me,resources=cronjobmanagers/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the CronjobManager object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.10.0/pkg/reconcile
func (r *CronJobManagerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	var cjMng cronjobmanagerv1beta1.CronJobManager
	logger.Info("Fetching CronJobManager resource.")
	err := r.Get(ctx, req.NamespacedName, &cjMng)
	if errors.IsNotFound(err) {
		return ctrl.Result{}, nil
	}

	if err != nil {
		logger.Error(err, "unable to get CronJobManager", "name", req.NamespacedName)
		return ctrl.Result{}, err
	}

	if !cjMng.ObjectMeta.DeletionTimestamp.IsZero() {
		return ctrl.Result{}, nil
	}

	err = r.cleanupOwnedResources(ctx, cjMng)
	if err != nil {
		logger.Error(err, "unnable to clean up old cronjob resource for this CronJobManager")
		return ctrl.Result{}, err
	}

	err = r.reconcileCronJob(ctx, cjMng)
	if err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *CronJobManagerReconciler) reconcileCronJob(ctx context.Context, cjMng cronjobmanagerv1beta1.CronJobManager) error {
	logger := log.FromContext(ctx)

	for _, content := range cjMng.Spec.CronJobs {
		cj := &batchv1.CronJob{}
		cj.SetNamespace(cjMng.Namespace)
		cj.SetName(content.Name)

		op, err := ctrl.CreateOrUpdate(ctx, r.Client, cj, func() error {
			cj.Spec.Schedule = content.Schedule

			podTemplate := corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"app": content.Name},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:    content.Name,
							Image:   cjMng.Spec.Image,
							Command: content.Command,
						},
					},
					RestartPolicy: "OnFailure",
				},
			}

			cj.Spec.JobTemplate.Spec.Template = podTemplate
			return ctrl.SetControllerReference(&cjMng, cj, r.Scheme)
		})

		if err != nil {
			logger.Error(err, "unable to create or update CronJob")
			return err
		}

		if op != controllerutil.OperationResultNone {
			logger.Info("reconcile CronJob successfully", "op", op)
		}
	}

	return nil
}

func (r *CronJobManagerReconciler) cleanupOwnedResources(ctx context.Context, cjMng cronjobmanagerv1beta1.CronJobManager) error {
	logger := log.FromContext(ctx)
	logger.Info("finding existing CronJob for CronJobManager resource")

	var cronjobs batchv1.CronJobList
	if err := r.List(ctx, &cronjobs, client.InNamespace(cjMng.Namespace), client.MatchingFields(map[string]string{cronjobOwnerKey: cjMng.Name})); err != nil {
		return err
	}

	for _, existsCronjob := range cronjobs.Items {
		contain := false
		for _, cjMngCronjob := range cjMng.Spec.CronJobs {
			if existsCronjob.Name == cjMngCronjob.Name {
				// containしていたら
				contain = true
			}
		}

		if !contain {
			if err := r.Delete(ctx, &existsCronjob); err != nil {
				logger.Error(err, "failed to delete CronJob resources.")
				return err
			}
			logger.Info("delete cronjob resource: " + existsCronjob.Name)
		}
	}

	return nil
}

var (
	cronjobOwnerKey = ".metadata.controller"
	apiGVStr        = cronjobmanagerv1beta1.GroupVersion.String()
)

// SetupWithManager sets up the controller with the Manager.
func (r *CronJobManagerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	ctx := context.Background()

	err := mgr.GetFieldIndexer().IndexField(ctx, &batchv1.CronJob{}, cronjobOwnerKey, func(rawObj client.Object) []string {
		cronjob := rawObj.(*batchv1.CronJob)
		owner := metav1.GetControllerOf(cronjob)
		if owner == nil {
			return nil
		}

		if owner.APIVersion != apiGVStr || owner.Kind != "CronJobManager" {
			return nil
		}

		return []string{owner.Name}
	})

	if err != nil {
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&cronjobmanagerv1beta1.CronJobManager{}).
		Owns(&batchv1.CronJob{}).
		Complete(r)
}
