/*
Copyright 2024.

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

package controller

import (
	"context"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/dynamicConfigIp/api/betav1"
)

// IpconfReconciler reconciles a Ipconf object
type PodReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=dynamicconfigip.github.com,resources=ipconfs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=dynamicconfigip.github.com,resources=ipconfs/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=dynamicconfigip.github.com,resources=ipconfs/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Ipconf object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.19.1/pkg/reconcile
func (r *PodReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	// TODO(user): your logic here
	reqLogger := log.FromContext(ctx)
	reqLogger.Info("Enter StatefulSet Reconciling")

	var StatefulSet appsv1.StatefulSet
	if err := r.Get(ctx, req.NamespacedName, &StatefulSet); err != nil {
		if errors.IsNotFound(err) {
			reqLogger.Info("Not found StatefulSet", "error", err.Error())
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		reqLogger.Error(err, "unable to get StatefulSet")
		return ctrl.Result{}, err
	}
	//list all the sideCarContainer
	sideCarContainerList := &betav1.SideCarContainerList{}
	err := r.List(ctx, sideCarContainerList)
	if err != nil {
		reqLogger.Info("Failed to list Pods.")
		return ctrl.Result{}, err
	}

	if len(sideCarContainerList.Items) == 0 {
		return ctrl.Result{}, err
	}
	sideCarContainer := sideCarContainerList.Items[0]

	if StatefulSet.Spec.ServiceName == "" {
		StatefulSet.Spec.ServiceName = sideCarContainer.Spec.HeadlessServiceName
	}

	// Define a new container based on the Ipconf spec and SideCarContainer
	newContainer := corev1.Container{
		Name:            sideCarContainer.Spec.ContainerName,
		Image:           sideCarContainer.Spec.Repo + ":" + sideCarContainer.Spec.ImageVersion,
		ImagePullPolicy: corev1.PullIfNotPresent,
	}

	// Check if the container already exists in the StatefulSet
	containerExists := false
	for _, container := range StatefulSet.Spec.Template.Spec.Containers {
		if container.Name == newContainer.Name {
			containerExists = true
			break
		}
	}

	// Add the new container to the StatefulSet
	if !containerExists {
		StatefulSet.Spec.Template.Spec.Containers = append(StatefulSet.Spec.Template.Spec.Containers, newContainer)

		// Update the StatefulSet with the new container
		if err := r.Update(ctx, &StatefulSet); err != nil {
			reqLogger.Error(err, "unable to update StatefulSet with new container")
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *PodReconciler) SetupWithManager(mgr ctrl.Manager) error {
	podLabelSelector := labels.SelectorFromSet(map[string]string{"network-config/runtime-ip": "true"})

	return ctrl.NewControllerManagedBy(mgr).
		For(&appsv1.StatefulSet{}).
		WithEventFilter(predicate.Funcs{
			CreateFunc: func(e event.CreateEvent) bool {
				return podLabelSelector.Matches(labels.Set(e.Object.GetLabels()))
			},
			UpdateFunc: func(e event.UpdateEvent) bool {
				return podLabelSelector.Matches(labels.Set(e.ObjectNew.GetLabels()))
			},
			DeleteFunc: func(e event.DeleteEvent) bool {
				return podLabelSelector.Matches(labels.Set(e.Object.GetLabels()))
			},
			GenericFunc: func(e event.GenericEvent) bool {
				return podLabelSelector.Matches(labels.Set(e.Object.GetLabels()))
			},
		}).
		Watches(&betav1.SideCarContainer{}, &handler.EnqueueRequestForObject{}).
		Named("statefulsets controller").
		Complete(r)
}
