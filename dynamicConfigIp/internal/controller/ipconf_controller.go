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
	"fmt"

	"encoding/json"

	netv1 "github.com/k8snetworkplumbingwg/network-attachment-definition-client/pkg/apis/k8s.cni.cncf.io/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/dynamicConfigIp/api/betav1"
	dynamicconfigipbetav1 "github.com/dynamicConfigIp/api/betav1"
)

// IpconfReconciler reconciles a Ipconf object
type IpconfReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

func (r *IpconfReconciler) createNetworkAttachmentDefinition(ctx context.Context, ipConfiguration betav1.Ipconf, namespace string) error {
	// Define the IPAM configuration
	ipamConfig := map[string]interface{}{
		"type":       ipConfiguration.Spec.Type,
		"cniVersion": ipConfiguration.Spec.CNIVersion,
		"ipam": map[string]interface{}{
			"type":      "static",
			"addresses": []map[string]string{},
		},
	}

	for _, ipaddr := range ipConfiguration.Spec.IpItems {
		if ipaddr.Ipaddress != "" {
			ipamConfig["ipam"].(map[string]interface{})["addresses"] = []map[string]string{
				{
					"address":   fmt.Sprintf("%s/%s", ipaddr.Ipaddress, ipaddr.Netmask),
					"interface": ipaddr.Iface,
				},
			}
		}
	}

	if ipConfiguration.Spec.Trust != "" {
		ipamConfig["trust"] = ipConfiguration.Spec.Trust
	}
	// Serialize the IPAM configuration to JSON
	ipamConfigJSON, err := json.Marshal(ipamConfig)
	if err != nil {
		return fmt.Errorf("error serializing IPAM configuration to JSON: %v", err)
	}

	// Create the NetworkAttachmentDefinition
	netAttachDef := &netv1.NetworkAttachmentDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ipConfiguration.Spec.Owner,
			Namespace: namespace,
		},
		Spec: netv1.NetworkAttachmentDefinitionSpec{
			Config: string(ipamConfigJSON),
		},
	}

	// Set the owner reference
	if err := controllerutil.SetControllerReference(&ipConfiguration, netAttachDef, r.Scheme); err != nil {
		return fmt.Errorf("error setting controller reference: %v", err)
	}

	// Create or update the NetworkAttachmentDefinition
	if err := r.Create(ctx, netAttachDef); err != nil && !errors.IsAlreadyExists(err) {
		return fmt.Errorf("error creating NetworkAttachmentDefinition: %v", err)
	}

	return nil
}

func (r *IpconfReconciler) updatePodAnnotations(ctx context.Context, pod corev1.Pod, ipConfiguration dynamicconfigipbetav1.Ipconf) (reconcile.Result, error) {
	if pod.Annotations == nil {
		pod.Annotations = make(map[string]string)
	}
	if existingNetworks, exists := pod.Annotations["k8s.v1.cni.cncf.io/networks"]; !exists || existingNetworks == "" {
		pod.Annotations["k8s.v1.cni.cncf.io/networks"] = ipConfiguration.Name
	} else {
		pod.Annotations["k8s.v1.cni.cncf.io/networks"] = existingNetworks + "," + ipConfiguration.Name
	}
	if err := r.Update(ctx, &pod); err != nil {
		return ctrl.Result{}, err
	}
	return reconcile.Result{}, nil
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
	reqLogger.Info("ipconf enter Reconciling")

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
	if err := r.List(ctx, &pods, client.InNamespace(req.Namespace), client.MatchingLabels{"app.kubernetes.io/instance": "mynginx"}); err != nil {
		reqLogger.Error(err, "unable to list Pods")
		return ctrl.Result{}, err
	}

	// Create NetworkAttachmentDefinition
	returnEc := r.createNetworkAttachmentDefinition(ctx, ipConfiguration, req.Namespace)
	if returnEc != nil {
		reqLogger.Info("NetworkAttachmentDefinition creation failed", "Error", returnEc)
		return ctrl.Result{}, returnEc
	}

	var errorCode error
	var result ctrl.Result
	for _, pod := range pods.Items {
		if pod.Name == ipConfiguration.Spec.Owner {
			reqLogger.Info("mached Pod details", "Name", pod.Name, "Namespace", pod.Namespace, "Labels", pod.Labels)
			returnResult, returnEc := r.updatePodAnnotations(ctx, pod, ipConfiguration)
			if returnEc != nil {
				errorCode = returnEc
				result = returnResult
				reqLogger.Info("Pod annotations updated failed", "Error", returnResult)
			}
		} else {
			reqLogger.Info("not matchec", "Name", pod.Name, "owner", ipConfiguration.Spec.Owner)
		}
	}

	return result, errorCode
}

// SetupWithManager sets up the controller with the Manager.
func (r *IpconfReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&dynamicconfigipbetav1.Ipconf{}).
		Named("ipconf").
		Complete(r)
}
