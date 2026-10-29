package controllers

import (
	"context"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	orgsv1 "github.com/zele-space/hello_zele/api/v1"
)

func TestGroupCreatesNamespaceAndQuota(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = corev1.AddToScheme(scheme)
	_ = orgsv1.AddToScheme(scheme)

	group := &orgsv1.Group{
		TypeMeta: metav1.TypeMeta{
			APIVersion: orgsv1.GroupVersion.String(),
			Kind:       "Group",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "demo-group",
		},
		Spec: orgsv1.GroupSpec{
			Namespace: "demo-group-ns",
			Quotas: map[string]string{
				"requests.cpu": "2",
				"limits.cpu":   "4",
			},
		},
	}

	cl := fake.NewClientBuilder().
		WithScheme(scheme).
		WithStatusSubresource(group).
		WithObjects(group).
		Build()
	reconciler := &GroupReconciler{
		Client:   cl,
		Scheme:   scheme,
		Recorder: record.NewFakeRecorder(eventBufferSize),
	}

	_, err := reconciler.Reconcile(context.Background(), reconcile.Request{
		NamespacedName: client.ObjectKeyFromObject(group),
	})
	if err != nil {
		t.Fatalf("reconcile failed: %v", err)
	}

	var ns corev1.Namespace
	if err := cl.Get(context.Background(), client.ObjectKey{Name: "demo-group-ns"}, &ns); err != nil {
		t.Fatalf("namespace not created: %v", err)
	}
	if ns.Labels["orgs.zele.space/group"] != "demo-group" {
		t.Fatalf("group label missing")
	}

	var quota corev1.ResourceQuota
	if err := cl.Get(context.Background(), client.ObjectKey{Name: "group-quota", Namespace: "demo-group-ns"}, &quota); err != nil {
		t.Fatalf("resourcequota not created: %v", err)
	}

	if cpu, ok := quota.Spec.Hard[corev1.ResourceRequestsCPU]; !ok || cpu.IsZero() {
		t.Fatalf("requests.cpu not set")
	}
	if cpu, ok := quota.Spec.Hard[corev1.ResourceLimitsCPU]; !ok || cpu.IsZero() {
		t.Fatalf("limits.cpu not set")
	}
}
