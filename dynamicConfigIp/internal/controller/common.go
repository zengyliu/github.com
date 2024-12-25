package controller

import (
	"context"
	"strings"

	corev1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	dynamicconfigipbetav1 "github.com/dynamicConfigIp/api/betav1"
)

func UpdatePodAnnotations(c client.Client, ctx context.Context, pod corev1.Pod, ipConfiguration dynamicconfigipbetav1.Ipconf) (reconcile.Result, error) {
	if pod.Annotations == nil {
		pod.Annotations = make(map[string]string)
	}

	if existingNetworks, exists := pod.Annotations["k8s.v1.cni.cncf.io/networks"]; !exists || existingNetworks == "" {
		pod.Annotations["k8s.v1.cni.cncf.io/networks"] = ipConfiguration.Name
	} else {
		if !strings.Contains(existingNetworks, ipConfiguration.Name) {
			pod.Annotations["k8s.v1.cni.cncf.io/networks"] = existingNetworks + "," + ipConfiguration.Name
		}
	}
	if err := c.Update(ctx, &pod); err != nil {
		return ctrl.Result{}, err
	}
	return reconcile.Result{}, nil
}
