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

func TestOrganizationCreatesNamespace(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = corev1.AddToScheme(scheme)
	_ = orgsv1.AddToScheme(scheme)

	org := &orgsv1.Organization{
		TypeMeta: metav1.TypeMeta{
			APIVersion: orgsv1.GroupVersion.String(),
			Kind:       "Organization",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "demo",
		},
		Spec: orgsv1.OrganizationSpec{
			Namespace: "demo-ns",
			Labels: map[string]string{
				"team": "platform",
			},
		},
	}

	cl := fake.NewClientBuilder().
		WithScheme(scheme).
		WithStatusSubresource(org).
		WithObjects(org).
		Build()
	reconciler := &OrganizationReconciler{
		Client:   cl,
		Scheme:   scheme,
		Recorder: record.NewFakeRecorder(16),
	}

	var fetched orgsv1.Organization
	if err := cl.Get(context.Background(), client.ObjectKeyFromObject(org), &fetched); err != nil {
		t.Fatalf("organization not persisted in fake client: %v", err)
	}

	_, err := reconciler.Reconcile(context.Background(), reconcile.Request{
		NamespacedName: client.ObjectKeyFromObject(org),
	})
	if err != nil {
		t.Fatalf("reconcile failed: %v", err)
	}

	var ns corev1.Namespace
	if err := cl.Get(context.Background(), client.ObjectKey{Name: "demo-ns"}, &ns); err != nil {
		t.Fatalf("namespace not created: %v", err)
	}
	if ns.Labels["orgs.zele.space/organization"] != "demo" {
		t.Fatalf("organization label missing")
	}
	if ns.Labels["team"] != "platform" {
		t.Fatalf("custom label not propagated")
	}
}
