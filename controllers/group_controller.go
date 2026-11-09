package controllers

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	orgsv1 "github.com/zele-space/hello_zele/api/v1"
)

// GroupReconciler reconciles a Group object.
type GroupReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

// +kubebuilder:rbac:groups=orgs.zele.space,resources=groups,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=orgs.zele.space,resources=groups/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=orgs.zele.space,resources=groups/finalizers,verbs=update
// +kubebuilder:rbac:groups="",resources=namespaces,verbs=get;list;watch;create;update;patch
// +kubebuilder:rbac:groups="",resources=resourcequotas,verbs=get;list;watch;create;update;patch

func (r *GroupReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	var group orgsv1.Group
	if err := r.Get(ctx, req.NamespacedName, &group); err != nil {
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}
	original := group.DeepCopy()

	targetNamespace := group.Spec.Namespace
	if targetNamespace == "" {
		targetNamespace = sanitizeName(group.Name)
	}

	ns := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: targetNamespace}}
	nsOp, err := controllerutil.CreateOrUpdate(ctx, r.Client, ns, func() error {
		if ns.Labels == nil {
			ns.Labels = map[string]string{}
		}
		ns.Labels["orgs.zele.space/group"] = group.Name
		return controllerutil.SetControllerReference(&group, ns, r.Scheme)
	})
	if err != nil {
		failed := group.DeepCopy()
		meta.SetStatusCondition(&failed.Status.Conditions, metav1.Condition{
			Type:               "Ready",
			Status:             metav1.ConditionFalse,
			Reason:             "ReconcileError",
			Message:            err.Error(),
			LastTransitionTime: metav1.Now(),
		})
		_ = r.Status().Patch(ctx, failed, client.MergeFrom(original))
		return ctrl.Result{}, err
	}
	if nsOp != controllerutil.OperationResultNone {
		logger.Info("reconciled namespace", "namespace", ns.Name, "operation", nsOp)
		r.Recorder.Eventf(&group, corev1.EventTypeNormal, "Reconciled", "Namespace %s %s", ns.Name, nsOp)
	}

	quota := &corev1.ResourceQuota{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "group-quota",
			Namespace: ns.Name,
		},
	}
	quotaOp, err := controllerutil.CreateOrUpdate(ctx, r.Client, quota, func() error {
		quota.Spec.Hard = corev1.ResourceList{}
		for k, v := range group.Spec.Quotas {
			qty, parseErr := resource.ParseQuantity(v)
			if parseErr != nil {
				return parseErr
			}
			quota.Spec.Hard[corev1.ResourceName(k)] = qty
		}
		return controllerutil.SetControllerReference(&group, quota, r.Scheme)
	})
	if err != nil {
		failed := group.DeepCopy()
		meta.SetStatusCondition(&failed.Status.Conditions, metav1.Condition{
			Type:               "Ready",
			Status:             metav1.ConditionFalse,
			Reason:             "ReconcileError",
			Message:            err.Error(),
			LastTransitionTime: metav1.Now(),
		})
		_ = r.Status().Patch(ctx, failed, client.MergeFrom(original))
		return ctrl.Result{}, err
	}
	if quotaOp != controllerutil.OperationResultNone {
		logger.Info("reconciled resourcequota", "namespace", ns.Name, "operation", quotaOp)
		r.Recorder.Eventf(&group, corev1.EventTypeNormal, "Reconciled", "ResourceQuota %s/%s %s", ns.Name, quota.Name, quotaOp)
	}

	updated := group.DeepCopy()
	updated.Status.NamespaceCreated = true
	updated.Status.ObservedNamespace = ns.Name
	updated.Status.QuotaCreated = true
	meta.SetStatusCondition(&updated.Status.Conditions, metav1.Condition{
		Type:               "Ready",
		Status:             metav1.ConditionTrue,
		Reason:             "QuotaAvailable",
		Message:            "Namespace and ResourceQuota are ready",
		LastTransitionTime: metav1.Now(),
	})

	if err := r.Status().Patch(ctx, updated, client.MergeFrom(original)); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *GroupReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&orgsv1.Group{}).
		Owns(&corev1.Namespace{}).
		Owns(&corev1.ResourceQuota{}).
		Complete(r)
}
