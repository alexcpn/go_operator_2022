/*
Copyright 2022.

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

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	log "sigs.k8s.io/controller-runtime/pkg/log"

	appsv1 "k8s.io/api/apps/v1" //ADDED
	"k8s.io/api/core/v1"
	corev1 "k8s.io/api/core/v1" //ADDED
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1" //ADDED
	"k8s.io/apimachinery/pkg/labels"
	ktypes "k8s.io/apimachinery/pkg/types"
	grpcappv1 "mytest.io/testoperator/api/v1"
)

// TestoperartorReconciler reconciles a Testoperartor object
type TestoperartorReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=grpcapp.mytest.io,resources=testoperartors,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=grpcapp.mytest.io,resources=testoperartors/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=grpcapp.mytest.io,resources=testoperartors/finalizers,verbs=update

//ADDED extra for creating deployment
// generate rbac to get,list, and watch pods
// +kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;watch
// generate rbac to get, list, watch, create, update, patch, and delete deployments
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Testoperartor object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.12.2/pkg/reconcile
func (r *TestoperartorReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	// TODO(user): your logic here
	var testOperator grpcappv1.Testoperartor
	if err := r.Get(ctx, req.NamespacedName, &testOperator); err != nil {
		log.Log.Error(err, "unable to fetch Test Operator")
		// we'll ignore not-found errors, since they can't be fixed by an immediate
		// requeue (we'll need to wait for a new notification), and we can get them
		// on deleted requests.
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	// ADDED - Block below
	log.Log.Info("Reconciling Test Operator", "Test Operator", testOperator)
	log.FromContext(ctx).Info("Pod Image is ", "PodImageName", testOperator.Spec.PodImage)
	// check if the PodImage is set
	if testOperator.Spec.PodImage == "" {
		log.Log.Info("Pod Image is not set")
	} else {
		log.Log.Info("Pod Image is set", "PodImageName", testOperator.Spec.PodImage)
	}

	// Let's check if a deployment is present

	found := &appsv1.Deployment{}
	namespaceName := ktypes.NamespacedName{Namespace: testOperator.Namespace, Name: testOperator.Name + "-deployment"}
	err := r.Get(ctx, namespaceName, found)
	if err == nil && !errors.IsNotFound(err) {
		log.Log.Info("Deployment Already  exists", namespaceName.Name, "Namespace", namespaceName.Namespace, "ok")

	} else {
		log.Log.Info("Deployment Does not exist", namespaceName.Name, "Namespace", namespaceName.Namespace, "ok")
		//Let's create a deployment
		one := int32(1)
		deployment := &appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      testOperator.Name + "-deployment",
				Namespace: testOperator.Namespace,
			},
			Spec: appsv1.DeploymentSpec{
				Replicas: &one,
				Selector: &metav1.LabelSelector{
					MatchLabels: map[string]string{
						"app": testOperator.Name,
					},
				},
				Template: corev1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{
							"app": testOperator.Name,
						},
					},
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{
								Name:  testOperator.Name,
								Image: testOperator.Spec.PodImage,
							},
						},
					},
				},
			},
		}
		if err := r.Create(ctx, deployment); err != nil {
			log.Log.Error(err, "unable to create Deployment", deployment.Namespace, "Deployment Name", deployment.Name)
			return ctrl.Result{}, err
		}
		log.Log.Info("Created Deployment", "Deployment.Namespace", deployment.Namespace, "Deployment.Name", deployment.Name)
	}
	// Lets wait for the deployment to be created and up and then update the status
	// Return all pods in the request namespace with a label of `instance=<name>`
	// and phase `Running`.
	podList := &v1.PodList{}
	labelSelector, err := labels.Parse("app=testoperartor-sample")
	opts := []client.ListOption{
		client.InNamespace(testOperator.Namespace),
		client.MatchingLabelsSelector{Selector: labelSelector},
	}
	if err := r.List(ctx, podList, opts...); err != nil {
		log.Log.Error(err, "unable to Get the Pod List")
		return ctrl.Result{}, err
	}
	var runningStatus bool

	for _, pod := range podList.Items {
		log.Log.Info("Pod Name", "", pod.Name, "Pod Status", pod.Status.Phase)
		if pod.Status.Phase == "Running" {
			runningStatus = true
		} else {
			runningStatus = false
		}
	}

	if runningStatus == true { //update the status
		testOperator.Status.CommonStatus.Healthy = true
		testOperator.Status.CommonStatus.Phase = "Running"
		if err := r.Status().Update(ctx, &testOperator); err != nil {
			log.Log.Error(err, "unable to update TestOperator status")
			return ctrl.Result{}, err
		} else {
			log.Log.Info("Test", "Updated Test Operator Status", &testOperator.Status.CommonStatus)
		}
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *TestoperartorReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&grpcappv1.Testoperartor{}).
		Complete(r)
}
