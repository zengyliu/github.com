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
	"bytes"
	"context"
	"fmt"
	"net/http"

	"encoding/json"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

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
	reqLogger.Info("Enter ipconf Reconciling")

	var ipConfiguration dynamicconfigipbetav1.Ipconf
	if err := r.Get(ctx, req.NamespacedName, &ipConfiguration); err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		reqLogger.Error(err, "unable to fetch ipconf")
		// we'll ignore not-found errors, since they can't be fixed by an immediate
		// requeue (we'll need to wait for a new notification), and we can get them
		// on deleted requests.
		return ctrl.Result{}, nil
	}
	// Check if the Pod exists
	var pod corev1.Pod
	err := r.Get(ctx, client.ObjectKey{Name: ipConfiguration.Spec.Owner, Namespace: req.Namespace}, &pod)
	if err != nil {
		if client.IgnoreNotFound(err) == nil {
			return ctrl.Result{}, nil
		}
		reqLogger.Error(err, "unable to fetch Pod")
		return ctrl.Result{}, nil
	}

	if ipConfiguration.Spec.IpItems == nil || len(ipConfiguration.Spec.IpItems) == 0 {
		reqLogger.Info("IpItems is empty")
		return ctrl.Result{}, nil
	}
	// Create NetworkUpdateRequest array from IpItems
	networkUpdateRequests := make([]NetworkUpdateRequest, len(ipConfiguration.Spec.IpItems))
	for i, ipItem := range ipConfiguration.Spec.IpItems {
		networkUpdateRequests[i] = NetworkUpdateRequest{
			Interface: ipItem.Iface,
			IPAddress: ipItem.Ipaddress,
			Netmask:   ipItem.Netmask,
			IpType:    ipItem.Type,
		}
	}

	// Serialize the NetworkUpdateRequests to JSON
	networkUpdateRequestsJSON, err := json.Marshal(networkUpdateRequests)
	if err != nil {
		reqLogger.Error(err, "error serializing NetworkUpdateRequests to JSON")
		return ctrl.Result{}, err
	}

	// Send the NetworkUpdateRequest to the server
	serverURL := fmt.Sprintf("http://%s:8080/network-update", ipConfiguration.Spec.Owner)
	reqLogger.Info("Sending NetworkUpdateRequest to server", "URL", serverURL)

	httpClient := &http.Client{}
	httpRequest, err := http.NewRequest("POST", serverURL, bytes.NewBuffer(networkUpdateRequestsJSON))
	if err != nil {
		reqLogger.Error(err, "error creating HTTP request")
		return ctrl.Result{}, err
	}
	httpRequest.Header.Set("Content-Type", "application/json")

	httpResponse, err := httpClient.Do(httpRequest)
	if err != nil {
		reqLogger.Error(err, "error sending HTTP request")
		return ctrl.Result{}, err
	}
	defer httpResponse.Body.Close()

	if httpResponse.StatusCode != http.StatusOK {
		reqLogger.Error(fmt.Errorf("received non-OK HTTP status: %s", httpResponse.Status), "error from server")
		return ctrl.Result{}, fmt.Errorf("received non-OK HTTP status: %s", httpResponse.Status)
	}

	reqLogger.Info("NetworkUpdateRequest sent successfully")
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *IpconfReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&dynamicconfigipbetav1.Ipconf{}).
		Watches(&corev1.Pod{}, &handler.EnqueueRequestForObject{}).
		Named("ipconf controller").
		Complete(r)
}
