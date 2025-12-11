package controllers

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	orgsv1 "github.com/zele-space/hello_zele/api/v1"
)

// OrganizationReconciler reconciles an Organization object.
type OrganizationReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

// +kubebuilder:rbac:groups=orgs.zele.space,resources=organizations,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=orgs.zele.space,resources=organizations/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=orgs.zele.space,resources=organizations/finalizers,verbs=update
// +kubebuilder:rbac:groups="",resources=namespaces,verbs=get;list;watch;create;update;patch

func (r *OrganizationReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	var org orgsv1.Organization
	if err := r.Get(ctx, req.NamespacedName, &org); err != nil {
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	targetNamespace := org.Spec.Namespace
	if targetNamespace == "" {
		targetNamespace = fmt.Sprintf("org-%s", org.Name)
	}

	ns := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: targetNamespace}}
	op, err := controllerutil.CreateOrUpdate(ctx, r.Client, ns, func() error {
		if ns.Labels == nil {
			ns.Labels = map[string]string{}
		}
		ns.Labels["orgs.zele.space/organization"] = org.Name
		for k, v := range org.Spec.Labels {
			ns.Labels[k] = v
		}
		return controllerutil.SetControllerReference(&org, ns, r.Scheme)
	})
	if err != nil {
		return ctrl.Result{}, err
	}

	if op != controllerutil.OperationResultNone {
		logger.Info("reconciled namespace", "namespace", ns.Name, "operation", op)
		r.Recorder.Eventf(&org, corev1.EventTypeNormal, "Reconciled", "Namespace %s %s", ns.Name, op)
	}

	updated := org.DeepCopy()
	updated.Status.NamespaceCreated = true
	updated.Status.ObservedNamespace = ns.Name
	meta.SetStatusCondition(&updated.Status.Conditions, metav1.Condition{
		Type:               "Ready",
		Status:             metav1.ConditionTrue,
		Reason:             "NamespaceAvailable",
		Message:            "Namespace is created and labeled",
		LastTransitionTime: metav1.Now(),
	})

	if err := r.Status().Patch(ctx, updated, client.MergeFrom(&org)); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *OrganizationReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&orgsv1.Organization{}).
		Owns(&corev1.Namespace{}).
		Complete(r)
}
