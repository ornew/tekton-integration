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
	"github.com/ornew/tekton-integration/api/v1alpha1"
)

// ProviderReconciler reconciles a Provider object
type ProviderReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=integrations.tekton.ornew.io,resources=providers,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=integrations.tekton.ornew.io,resources=providers/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=integrations.tekton.ornew.io,resources=providers/finalizers,verbs=update

func (r *ProviderReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logr.FromContext(ctx)
	log.V(2).Info("start")

	var provider v1alpha1.Provider
	if err := r.Get(ctx, req.NamespacedName, &provider); err != nil {
		if apierrors.IsNotFound(err) {
			log.V(2).Info("deleted")
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	if err := r.validate(ctx, provider); err != nil {
		patch := client.MergeFrom(provider.DeepCopy())
		r.setStatusCondition(&provider, v1alpha1.ReadyCondition, metav1.ConditionFalse, v1alpha1.ReconcileFailedReason, err.Error())
		if err := r.Status().Patch(ctx, &provider, patch); err != nil {
			return ctrl.Result{Requeue: true}, err
		}
		return ctrl.Result{Requeue: true}, err
	}

	if !apimeta.IsStatusConditionTrue(provider.Status.Conditions, v1alpha1.ReadyCondition) || provider.Status.ObservedGeneration != provider.Generation {
		patch := client.MergeFrom(provider.DeepCopy())
		r.setStatusCondition(&provider, v1alpha1.ReadyCondition, metav1.ConditionTrue, v1alpha1.InitializedReason, v1alpha1.InitializedReason)
		if err := r.Status().Patch(ctx, &provider, patch); err != nil {
			return ctrl.Result{Requeue: true}, err
		}
		log.Info("initialized")
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ProviderReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.Provider{}).
		Complete(r)
}

func (r *ProviderReconciler) validate(ctx context.Context, provider v1alpha1.Provider) error {
	// TODO
	return nil
}

func (r *ProviderReconciler) setStatusCondition(provider *v1alpha1.Provider, condition string, status metav1.ConditionStatus, reason, message string) {
	newCondition := metav1.Condition{
		Type:    condition,
		Status:  status,
		Reason:  reason,
		Message: message,
	}
	apimeta.SetStatusCondition(&provider.Status.Conditions, newCondition)
}
