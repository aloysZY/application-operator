/*
Copyright 2024 Aloys.Zhou.

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
	"time"

	dappsv1 "github.com/aloys.zy/application-operator/api/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// ApplicationReconciler reconciles a Application object
type ApplicationReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=apps.aloys.cn,resources=applications,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps.aloys.cn,resources=applications/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=apps.aloys.cn,resources=applications/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Application object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.18.4/pkg/reconcile
func (r *ApplicationReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	// _ = log.FromContext(ctx)

	// TODO(user): your logic here
	// 写具体的调谐逻辑

	l := log.FromContext(ctx)

	// 声明一个*application类型的实例app用来接收CR
	app := &dappsv1.Application{}

	// namespaceName在这里也就是 default/application-sample
	if err := r.Get(ctx, req.NamespacedName, app); err != nil {
		// err分很多情况，如果找不到，一半不需要进一步处理，只是这个CR被删除了
		if errors.IsNotFound(err) {
			l.Info("the Application is not found.")
			return ctrl.Result{}, nil
		}
		// 其他错误还有很多，比如连接apiserver等打印错误信息，然后一分钟后重试
		l.Error(err, "failed to get the Application.")
		return ctrl.Result{RequeueAfter: 1 * time.Minute}, err
	}

	// 根据副本数循环创建pod
	for i := 0; i < int(app.Spec.Replicas); i++ {
		pod := &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      fmt.Sprintf("%s-%d", app.Name, i),
				Namespace: app.Namespace,
				Labels:    app.Labels,
			},
			Spec: app.Spec.Template.Spec,
		}
		if err := r.Create(ctx, pod); err != nil {
			l.Error(err, "failed to create pod for the Application.")
			return ctrl.Result{RequeueAfter: 1 * time.Minute}, err
		}
		l.Info("created pod for the Application.", "Pod.Namespace", pod.Namespace, "Pod.Name", pod.Name)
	}
	l.Info("all pods has created")
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ApplicationReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&dappsv1.Application{}).
		Complete(r)
}
