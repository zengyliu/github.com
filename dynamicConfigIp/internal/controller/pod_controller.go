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

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

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
	reqLogger.Info("pod enter Reconciling")

	var pod corev1.Pod
	if err := r.Get(ctx, req.NamespacedName, &pod); err != nil {
		reqLogger.Error(err, "unable to get Pod")
		return ctrl.Result{}, err
	}

	// Fetch the Pods with the specified label
	var ipConfigurations betav1.IpconfList
	if err := r.List(ctx, &ipConfigurations, client.InNamespace(req.Namespace)); err != nil {
		reqLogger.Error(err, "unable to list ipConfigurations")
		return ctrl.Result{}, err
	}

	var errorCode error
	var result ctrl.Result
	// TODO: Add your logic here to handle the Pods with the specified label
	for _, ipConf := range ipConfigurations.Items {
		if pod.Name == ipConf.Spec.Owner {
			reqLogger.Info("Pod found with ipconf owner", "Pod", pod.Name)
			returnResult, returnEc := UpdatePodAnnotations(r.Client, ctx, pod, ipConf)
			if returnEc != nil {
				errorCode = returnEc
				result = returnResult
				reqLogger.Info("Pod annotations updated failed", "Error", returnResult)
			}
		}
	}

	return result, errorCode
}

// SetupWithManager sets up the controller with the Manager.
func (r *PodReconciler) SetupWithManager(mgr ctrl.Manager) error {
	podLabelSelector := labels.SelectorFromSet(map[string]string{"app.kubernetes.io/instance": "mynginx"})

	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Pod{}).
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
		Named("pods with ubuntu").
		Complete(r)
}
