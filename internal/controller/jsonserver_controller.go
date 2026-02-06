/*
Copyright 2026.

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

// import (
// 	"context"

// 	"k8s.io/apimachinery/pkg/runtime"
// 	ctrl "sigs.k8s.io/controller-runtime"
// 	"sigs.k8s.io/controller-runtime/pkg/client"
// 	logf "sigs.k8s.io/controller-runtime/pkg/log"

// 	examplev1 "github.com/BlueTurtle-bytes/json-server/api/v1"
// )

import (
	"context"
	"encoding/json"
	"reflect"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	examplev1 "github.com/BlueTurtle-bytes/json-server/api/v1"
)

// JsonServerReconciler reconciles a JsonServer object
type JsonServerReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// RBAC
// +kubebuilder:rbac:groups=example.com,resources=jsonservers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=example.com,resources=jsonservers/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=services;configmaps,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=coordination.k8s.io,resources=leases,verbs=get;list;watch;create;update;patch;delete

func (r *JsonServerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	var js examplev1.JsonServer
	if err := r.Get(ctx, req.NamespacedName, &js); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// -------------------- JSON Validation --------------------
	var parsed any
	if err := json.Unmarshal([]byte(js.Spec.JsonConfig), &parsed); err != nil {
		logger.Info("invalid jsonConfig detected", "name", js.Name, "error", err)

		r.updateStatus(ctx, &js, "Error", "Error: spec.jsonConfig is not a valid json object")

		// Stop reconciliation – do NOT create/update resources
		return ctrl.Result{}, nil
	}

	if err := r.reconcileConfigMap(ctx, &js); err != nil {
		logger.Error(err, "failed to reconcile ConfigMap")
		r.updateStatus(ctx, &js, "Error", "Error: unexpected failure")
		return ctrl.Result{}, err
	}

	// if err := r.reconcileDeployment(ctx, &js); err != nil {
	// 	logger.Error(err, "failed to reconcile Deployment")
	// 	r.updateStatus(ctx, &js, "Error", "Error: unexpected failure")
	// 	return ctrl.Result{}, err
	// }

	deploy, err := r.reconcileDeployment(ctx, &js)
	if err != nil {
		logger.Error(err, "failed to reconcile Deployment")
		r.updateStatus(ctx, &js, "Error", "Error: unexpected failure")
		return ctrl.Result{}, err
	}

	// ✅ Accurate replica reporting
	js.Status.Replicas = deploy.Status.ReadyReplicas

	if err := r.reconcileService(ctx, &js); err != nil {
		logger.Error(err, "failed to reconcile Service")
		r.updateStatus(ctx, &js, "Error", "Error: unexpected failure")
		return ctrl.Result{}, err
	}

	// replicas := int32(1)
	// if js.Spec.Replicas != nil {
	// 	replicas = *js.Spec.Replicas
	// }
	// // Sync replicas into status for scale subresource
	// js.Status.Replicas = replicas

	r.updateStatus(ctx, &js, "Synced", "Synced successfully!")
	return ctrl.Result{}, nil
}

// -------------------- ConfigMap --------------------

func (r *JsonServerReconciler) reconcileConfigMap(ctx context.Context, js *examplev1.JsonServer) error {
	cm := &corev1.ConfigMap{}
	err := r.Get(ctx, types.NamespacedName{
		Name:      js.Name,
		Namespace: js.Namespace,
	}, cm)

	desired := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      js.Name,
			Namespace: js.Namespace,
		},
		Data: map[string]string{
			"db.json": js.Spec.JsonConfig,
		},
	}

	if apierrors.IsNotFound(err) {
		if err := controllerutil.SetControllerReference(js, desired, r.Scheme); err != nil {
			return err
		}
		return r.Create(ctx, desired)
	}

	if err != nil {
		return err
	}

	if !reflect.DeepEqual(cm.Data, desired.Data) {
		cm.Data = desired.Data
		return r.Update(ctx, cm)
	}

	return nil
}

// -------------------- Deployment --------------------

// func (r *JsonServerReconciler) reconcileDeployment(ctx context.Context, js *examplev1.JsonServer) error {
// 	deploy := &appsv1.Deployment{}
// 	err := r.Get(ctx, types.NamespacedName{
// 		Name:      js.Name,
// 		Namespace: js.Namespace,
// 	}, deploy)

// 	replicas := int32(1)
// 	if js.Spec.Replicas != nil {
// 		replicas = *js.Spec.Replicas
// 	}

// 	desired := desiredDeployment(js, replicas)

// 	if apierrors.IsNotFound(err) {
// 		controllerutil.SetControllerReference(js, desired, r.Scheme)
// 		return r.Create(ctx, desired)
// 	}

// 	if err != nil {
// 		return err
// 	}

// 	if *deploy.Spec.Replicas != replicas {
// 		deploy.Spec.Replicas = &replicas
// 		return r.Update(ctx, deploy)
// 	}

// 	return nil
// }

func (r *JsonServerReconciler) reconcileDeployment(ctx context.Context, js *examplev1.JsonServer) (*appsv1.Deployment, error) {

	deploy := &appsv1.Deployment{}
	err := r.Get(ctx, types.NamespacedName{
		Name:      js.Name,
		Namespace: js.Namespace,
	}, deploy)

	replicas := int32(1)
	if js.Spec.Replicas != nil {
		replicas = *js.Spec.Replicas
	}

	desired := desiredDeployment(js, replicas)

	if apierrors.IsNotFound(err) {
		if err := controllerutil.SetControllerReference(js, desired, r.Scheme); err != nil {
			return err
		}
		if err := r.Create(ctx, desired); err != nil {
			return nil, err
		}
		return desired, nil
	}

	if err != nil {
		return nil, err
	}

	if *deploy.Spec.Replicas != replicas {
		deploy.Spec.Replicas = &replicas
		if err := r.Update(ctx, deploy); err != nil {
			return nil, err
		}
	}

	return deploy, nil
}

func desiredDeployment(js *examplev1.JsonServer, replicas int32) *appsv1.Deployment {
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      js.Name,
			Namespace: js.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": js.Name,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": js.Name,
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "json-server",
							Image: "backplane/json-server",
							Args:  []string{"/data/db.json"},
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: 3000,
									Protocol:      corev1.ProtocolTCP,
								},
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "json-config",
									MountPath: "/data",
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "json-config",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: js.Name,
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

// -------------------- Service --------------------

func (r *JsonServerReconciler) reconcileService(ctx context.Context, js *examplev1.JsonServer) error {
	svc := &corev1.Service{}
	err := r.Get(ctx, types.NamespacedName{
		Name:      js.Name,
		Namespace: js.Namespace,
	}, svc)

	if apierrors.IsNotFound(err) {
		desired := &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      js.Name,
				Namespace: js.Namespace,
			},
			Spec: corev1.ServiceSpec{
				Selector: map[string]string{
					"app": js.Name,
				},
				Ports: []corev1.ServicePort{
					{
						Port:       3000,
						TargetPort: intstr.FromInt(3000),
					},
				},
			},
		}

		if err := controllerutil.SetControllerReference(js, desired, r.Scheme); err != nil {
			return err
		}
		return r.Create(ctx, desired)
	}

	return err
}

// -------------------- Status --------------------

func (r *JsonServerReconciler) updateStatus(
	ctx context.Context,
	js *examplev1.JsonServer,
	state, message string,
) {
	js.Status.State = state
	js.Status.Message = message
	_ = r.Status().Update(ctx, js)
}

// -------------------- Setup --------------------

func (r *JsonServerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&examplev1.JsonServer{}).
		Owns(&appsv1.Deployment{}).
		Owns(&corev1.Service{}).
		Owns(&corev1.ConfigMap{}).
		Complete(r)
}
