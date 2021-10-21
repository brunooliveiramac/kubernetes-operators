/*
Copyright 2021.

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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	appv1alpha1 "brunooliveiramac/stack-instance-operator/api/v1alpha1"

	corev1 "k8s.io/api/core/v1"
)

// StackInstanceReconciler reconciles a StackInstance object
type StackInstanceReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=app.platform.com,resources=stackinstances,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=app.platform.com,resources=stackinstances/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=app.platform.com,resources=stackinstances/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the StackInstance object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.9.2/pkg/reconcile
func (r *StackInstanceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	// your logic here
	log := ctrllog.FromContext(ctx)

	// Fetch the PodSet instance
	instance := &appv1alpha1.StackInstance{}
	err := r.Get(context.TODO(), req.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return ctrl.Result{}, err

	}

	// List all pods owned by this PodSet instance
	podSet := instance
	podList := &corev1.PodList{}
	lbs := map[string]string{
		"app":     podSet.Name,
		"version": "v0.1",
	}
	labelSelector := labels.SelectorFromSet(lbs)
	listOps := &client.ListOptions{Namespace: podSet.Namespace, LabelSelector: labelSelector}
	if err = r.List(context.TODO(), podList, listOps); err != nil {
		return ctrl.Result{}, err
	}

	// Count the pods that are pending or running as available
	var available []corev1.Pod
	for _, pod := range podList.Items {
		if pod.ObjectMeta.DeletionTimestamp != nil {
			continue
		}
		if pod.Status.Phase == corev1.PodRunning || pod.Status.Phase == corev1.PodPending {
			available = append(available, pod)
		}
	}
	numAvailable := int32(len(available))
	availableNames := []string{}
	for _, pod := range available {
		availableNames = append(availableNames, pod.ObjectMeta.Name)
	}

	// Update the status if necessary
	status := appv1alpha1.StackInstanceStatus{
		PodNames:          availableNames,
		AvailableReplicas: numAvailable,
	}
	if !reflect.DeepEqual(podSet.Status, status) {
		podSet.Status = status
		err = r.Status().Update(context.TODO(), podSet)
		if err != nil {
			log.Error(err, "Failed to update PodSet status")
			return ctrl.Result{}, err
		}
	}

	if numAvailable > podSet.Spec.Replicas {
		log.Info("Scaling down pods", "Currently available", numAvailable, "Required replicas", podSet.Spec.Replicas)
		diff := numAvailable - podSet.Spec.Replicas
		dpods := available[:diff]
		for _, dpod := range dpods {
			err = r.Delete(context.TODO(), &dpod)
			if err != nil {
				log.Error(err, "Failed to delete pod", "pod.name", dpod.Name)
				return ctrl.Result{}, err
			}
		}
		return ctrl.Result{Requeue: true}, nil
	}

	if numAvailable < podSet.Spec.Replicas {
		log.Info("Scaling up pods", "Currently available", numAvailable)
		// Define a new Pod object
		pod := newPodForCR(podSet)
		// Set PodSet instance as the owner and controller
		if err := controllerutil.SetControllerReference(podSet, pod, r.Scheme); err != nil {
			return ctrl.Result{}, err
		}
		err = r.Create(context.TODO(), pod)
		if err != nil {
			log.Error(err, "Failed to create pod", "pod.name", pod.Name)
			return ctrl.Result{}, err
		}
		return ctrl.Result{Requeue: true}, nil
	}

	return ctrl.Result{}, nil
}

func newPodForCR(cr *appv1alpha1.StackInstance) *corev1.Pod {
	labels := map[string]string{
		"app":     cr.Name,
		"version": "v0.1",
	}
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: cr.Name + "-pod",
			Namespace:    cr.Namespace,
			Labels:       labels,
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:    "stack",
					Image:   "docker.io/silviosilva/azure/example:0.1.0",
					Command: []string{"apply"},
					Args:    []string{"-auto-approve"},
				},
			},
		},
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *StackInstanceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&appv1alpha1.StackInstance{}).
		Complete(r)
}
