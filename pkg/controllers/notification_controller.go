/*
Copyright 2021 Arata Furukawa.

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

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/go-logr/logr"
	"github.com/ornew/tekton-integration/pkg/api/v1alpha1"
)

// NotificationReconciler reconciles a Notification object
type NotificationReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=integrations.tekton.ornew.io,resources=notifications,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=integrations.tekton.ornew.io,resources=notifications/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=integrations.tekton.ornew.io,resources=notifications/finalizers,verbs=update

func (r *NotificationReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logr.FromContext(ctx)
	log.V(2).Info("start")

	var notif v1alpha1.Notification
	if err := r.Get(ctx, req.NamespacedName, &notif); err != nil {
		if apierrors.IsNotFound(err) {
			log.V(2).Info("deleted")
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	if err := r.validate(ctx, notif); err != nil {
		patch := client.MergeFrom(notif.DeepCopy())
		r.setStatusCondition(&notif, v1alpha1.ReadyCondition, metav1.ConditionFalse, v1alpha1.ReconcileFailedReason, err.Error())
		if err := r.Status().Patch(ctx, &notif, patch); err != nil {
			return ctrl.Result{Requeue: true}, err
		}
		return ctrl.Result{Requeue: true}, err
	}

	if !apimeta.IsStatusConditionTrue(notif.Status.Conditions, v1alpha1.ReadyCondition) || notif.Status.ObservedGeneration != notif.Generation {
		patch := client.MergeFrom(notif.DeepCopy())
		r.setStatusCondition(&notif, v1alpha1.ReadyCondition, metav1.ConditionTrue, v1alpha1.InitializedReason, v1alpha1.InitializedReason)
		if err := r.Status().Patch(ctx, &notif, patch); err != nil {
			return ctrl.Result{Requeue: true}, err
		}
		log.Info("initialized")
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *NotificationReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.Notification{}).
		Complete(r)
}

func (r *NotificationReconciler) validate(ctx context.Context, notif v1alpha1.Notification) error {
	// TODO
	return nil
}

func (r *NotificationReconciler) setStatusCondition(notif *v1alpha1.Notification, condition string, status metav1.ConditionStatus, reason, message string) {
	newCondition := metav1.Condition{
		Type:    condition,
		Status:  status,
		Reason:  reason,
		Message: message,
	}
	apimeta.SetStatusCondition(&notif.Status.Conditions, newCondition)
}
