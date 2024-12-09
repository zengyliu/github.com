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
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/dynamicConfigIp/api/betav1"
	dynamicconfigipbetav1 "github.com/dynamicConfigIp/api/betav1"
)

// IpconfReconciler reconciles a Ipconf object
type IpconfReconciler struct {
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
func (r *IpconfReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	reqLogger := log.FromContext(ctx)
	reqLogger.Info("enter Reconciling")

	var ipConfiguration betav1.Ipconf
	if err := r.Get(ctx, req.NamespacedName, &ipConfiguration); err != nil {
		reqLogger.Error(err, "unable to fetch CronJob")
		// we'll ignore not-found errors, since they can't be fixed by an immediate
		// requeue (we'll need to wait for a new notification), and we can get them
		// on deleted requests.
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Fetch the Pods with the specified label
	var pods corev1.PodList
	if err := r.List(ctx, &pods, client.InNamespace(req.Namespace), client.MatchingLabels{"ubuntu": "ubuntu"}); err != nil {
		reqLogger.Error(err, "unable to list Pods")
		return ctrl.Result{}, err
	}

	// TODO: Add your logic here to handle the Pods with the specified label
	for _, pod := range pods.Items {
		reqLogger.Info("Pod details", "Name", pod.Name, "Namespace", pod.Namespace, "Labels", pod.Labels)
	}

	reqLogger.Info("Ipconf details", "Spec", ipConfiguration.Spec, "Status", ipConfiguration.Status)

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *IpconfReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&dynamicconfigipbetav1.Ipconf{}).
		Named("ipconf").
		Complete(r)
}
