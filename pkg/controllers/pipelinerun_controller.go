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
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/go-logr/logr"
	pipelinesv1beta1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	knativeapis "knative.dev/pkg/apis"

	"github.com/ornew/tekton-integration/internal/providers"
	"github.com/ornew/tekton-integration/pkg/api/v1alpha1"
)

const (
	annotationLastStatus = "integrations.tekton.ornew.io/last-status"
)

// PipelineRunReconciler reconciles a PipelineRun object
type PipelineRunReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=tekton.dev,resources=pipelineruns,verbs=get;list;watch;update;patch
//+kubebuilder:rbac:groups=tekton.dev,resources=pipelineruns/status,verbs=get;update;patch

func (r *PipelineRunReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logr.FromContext(ctx)
	log.V(2).Info("start")

	var pr pipelinesv1beta1.PipelineRun
	if err := r.Get(ctx, req.NamespacedName, &pr); err != nil {
		if apierrors.IsNotFound(err) {
			log.V(2).Info("deleted")
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	cond := pr.Status.GetCondition(knativeapis.ConditionSucceeded)
	if cond == nil {
		// we won't handle if missing conditions
		return ctrl.Result{}, nil
	}

	status := string(cond.Status)
	last := pr.Annotations[annotationLastStatus]
	if last == "" || status != last {
		log.Info("PipelineRun status changed", "status", status, "last", last)

		if pr.Annotations == nil {
			log.Info("annotation is nil")
			pr.SetAnnotations(make(map[string]string))
		}

		pr.Annotations[annotationLastStatus] = status
		if err := r.Update(ctx, &pr); err != nil {
			if apierrors.IsConflict(err) {
				return ctrl.Result{Requeue: true}, nil
			}
			if apierrors.IsNotFound(err) {
				return ctrl.Result{Requeue: true}, nil
			}
			log.Error(err, "unable to update PipelineRun")
			return ctrl.Result{}, err
		}

		// TODO no blocking reconcile
		//ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		//defer cancel()

		var allNotif v1alpha1.NotificationList
		err := r.Client.List(ctx, &allNotif)
		if err != nil {
			log.Error(err, "failed to list notifications")
			return ctrl.Result{Requeue: true}, nil
		}
		notifs := make([]v1alpha1.Notification, 0)
		for _, notif := range allNotif.Items {
			isReady := apimeta.IsStatusConditionTrue(notif.Status.Conditions, v1alpha1.ReadyCondition)
			if notif.Spec.Suspend || !isReady {
				continue
			}
			// TODO filtering
			notifs = append(notifs, notif)
		}
		if len(notifs) == 0 {
			log.Info("matched notifications are not found")
			return ctrl.Result{}, nil
		}
		for _, notif := range notifs {
			var provider v1alpha1.Provider
			providerRef := types.NamespacedName{
				Namespace: notif.Namespace,
				Name:      notif.Spec.ProviderRef.Name,
			}
			if err := r.Client.Get(ctx, providerRef, &provider); err != nil {
				log.Error(err, "failed to get Provider", "provider", providerRef)
				continue
			}
			logp := log.WithValues("provider", providerRef, "type", provider.Spec.Type)
			app, err := providers.ResolveProvider(ctx, &provider, r.Client)
			if err != nil {
				logp.Error(err, "failed to create the provider app")
				continue
			}
			log.Info("get provider app", "app", app)
			if err := app.Notify(ctx, &pr); err != nil {
				logp.Error(err, "failed to notify")
			}
		}
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *PipelineRunReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&pipelinesv1beta1.PipelineRun{}).
		Complete(r)
}
