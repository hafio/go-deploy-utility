package nsgate

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
)

func TestProtected(t *testing.T) {
	for ns, want := range map[string]bool{
		"default": true, "kube-system": true, "kube-public": true, "kube-node-lease": true, "": true,
		"solace": false, "prod": false, "kube-other": false,
	} {
		if got := Protected(ns); got != want {
			t.Errorf("Protected(%q) = %v, want %v", ns, got, want)
		}
	}
}

// TestListArgv: one query, JSON, and no --ignore-not-found, with which kubectl
// prints nothing for an empty list (both tools' one-query tests).
func TestListArgv(t *testing.T) {
	base := make([]string, 1, 8)
	base[0] = "kubectl"
	got := List(base, "prod").Argv
	if want := []string{"kubectl", "get", "all,persistentvolumeclaims,secrets,configmaps", "-n", "prod", "-o", "json"}; !reflect.DeepEqual(got, want) {
		t.Errorf("argv = %v, want %v", got, want)
	}
	if base[:2][1] != "" {
		t.Errorf("the base was written through: %v", base[:2])
	}
}

func TestDeleteProbe(t *testing.T) {
	if p := DeleteProbe(); p.Verb != "delete" || p.Resource != "namespaces" || p.Namespace != "" {
		t.Errorf("DeleteProbe = %+v", p)
	}
}

// obj is one listing item as JSON: kind, name, then extra metadata or top-level
// fields as raw JSON fragments.
func obj(kind, name, metadata, top string) string {
	if metadata != "" {
		metadata = ", " + metadata
	}
	if top != "" {
		top = ", " + top
	}
	return fmt.Sprintf(`{"kind": %q, "metadata": {"name": %q%s}%s}`, kind, name, metadata, top)
}

// listing wraps items in a List.
func listing(items ...string) string {
	return `{"apiVersion": "v1", "kind": "List", "items": [` + strings.Join(items, ", ") + `]}`
}

func occupants(t *testing.T, out string, owned []Object) []string {
	t.Helper()
	got, err := Occupants(out, owned)
	if err != nil {
		t.Fatalf("Occupants: %v", err)
	}
	return got
}

// TestOccupantsCountsForeignObjects carries solmq-conn-util's
// TestNamespaceOccupantsRules and TestNamespaceOccupantsAreSortedAndLabelled, and
// solace-util's TestNamespaceContentsReportsRealOccupants.
func TestOccupantsCountsForeignObjects(t *testing.T) {
	got := occupants(t, listing(
		obj("StatefulSet", "postgres", "", ""),
		obj("Deployment", "billing-api", "", ""),
		obj("Secret", "their-token", "", `"type": "Opaque"`),
		obj("Service", "billing-api", "", ""),
	), nil)
	want := []string{"deployment/billing-api", "secret/their-token", "service/billing-api", "statefulset/postgres"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Occupants = %v, want %v", got, want)
	}
	if got := occupants(t, `{"items": []}`, nil); got != nil {
		t.Errorf("an empty namespace = %v, want no occupants", got)
	}
}

// TestOccupantsDiscountsClusterDefaults carries both tools' default lists: what
// Kubernetes puts in every namespace, and the namespace's own policy objects.
func TestOccupantsDiscountsClusterDefaults(t *testing.T) {
	got := occupants(t, listing(
		obj("ConfigMap", "kube-root-ca.crt", "", ""),
		obj("ServiceAccount", "default", "", ""),
		obj("Secret", "default-token-abcde", `"annotations": {"kubernetes.io/service-account.name": "default"}`,
			`"type": "kubernetes.io/service-account-token"`),
		obj("LimitRange", "mem-default", "", ""),
		obj("ResourceQuota", "compute", "", ""),
		obj("Event", "something.17", "", ""),
		obj("Secret", "builder-token-xyz", `"annotations": {"kubernetes.io/service-account.name": "builder"}`,
			`"type": "kubernetes.io/service-account-token"`),
		obj("ConfigMap", "kube-root-ca.crt-copy", "", ""),
	), nil)
	if want := []string{"configmap/kube-root-ca.crt-copy", "secret/builder-token-xyz"}; !reflect.DeepEqual(got, want) {
		t.Errorf("Occupants = %v, want only %v: another service account's token is someone else's", got, want)
	}
}

// TestOccupantsDiscountsOwnedAndTerminating: the deployment's own objects by kind and
// name, what they own followed by uid or kind and name, a root that is not listed,
// and anything already terminating with its children.
func TestOccupantsDiscountsOwnedAndTerminating(t *testing.T) {
	owned := []Object{{"Deployment", "solmq-connector"}, {"ConfigMap", "solmq-connector-config"},
		{"pubsubpluseventbroker", "dev-broker"}}
	got := occupants(t, listing(
		obj("Deployment", "solmq-connector", `"uid": "d1"`, ""),
		obj("ConfigMap", "solmq-connector-config", "", ""),
		// uid chain: Deployment -> ReplicaSet -> Pod.
		obj("ReplicaSet", "solmq-connector-7c9", `"uid": "r1", "ownerReferences": [{"kind": "Deployment", "name": "solmq-connector", "uid": "d1"}]`, ""),
		obj("Pod", "solmq-connector-7c9-x", `"ownerReferences": [{"kind": "ReplicaSet", "name": "solmq-connector-7c9", "uid": "r1"}]`, ""),
		// CR root, not listed: CR -> StatefulSet -> Pod, by kind and name.
		obj("StatefulSet", "dev-broker-pubsubplus-p", `"ownerReferences": [{"kind": "PubSubPlusEventBroker", "name": "dev-broker"}]`, ""),
		obj("Pod", "dev-broker-pubsubplus-p-0", `"ownerReferences": [{"kind": "StatefulSet", "name": "dev-broker-pubsubplus-p"}]`, ""),
		// Terminating, and its child.
		obj("Deployment", "old-app", `"uid": "o1", "deletionTimestamp": "2026-11-05T01:00:00Z"`, ""),
		obj("ReplicaSet", "old-app-1", `"ownerReferences": [{"kind": "Deployment", "name": "old-app", "uid": "o1"}]`, ""),
		// A terminating owner's uid differs from the one this ReplicaSet names, so the
		// ReplicaSet belongs to a newer object of that name and still counts.
		obj("Deployment", "web", `"uid": "w1", "deletionTimestamp": "2026-11-05T01:00:00Z"`, ""),
		obj("ReplicaSet", "web-2", `"ownerReferences": [{"kind": "Deployment", "name": "web", "uid": "w2"}]`, ""),
	), owned)
	if want := []string{"replicaset/web-2"}; !reflect.DeepEqual(got, want) {
		t.Errorf("Occupants = %v, want %v", got, want)
	}
}

// TestOccupantsNameCollisionAcrossKinds: owning a Deployment named broker says
// nothing about a PersistentVolumeClaim of the same name, or what that claim owns.
func TestOccupantsNameCollisionAcrossKinds(t *testing.T) {
	got := occupants(t, listing(
		obj("Deployment", "broker", "", ""),
		obj("PersistentVolumeClaim", "broker", "", ""),
		obj("Pod", "broker-x", `"ownerReferences": [{"kind": "PersistentVolumeClaim", "name": "broker"}]`, ""),
	), []Object{{"deployment", "broker"}})
	if want := []string{"persistentvolumeclaim/broker", "pod/broker-x"}; !reflect.DeepEqual(got, want) {
		t.Errorf("Occupants = %v, want %v", got, want)
	}
}

// TestOccupantsStatefulSetPVCsCount: a StatefulSet's claims carry no ownerReference,
// so they hold their data, and the namespace, after the StatefulSet goes
// (solace-util's TestRetainedPVCsKeepTheNamespace).
func TestOccupantsStatefulSetPVCsCount(t *testing.T) {
	got := occupants(t, listing(
		obj("StatefulSet", "dev-broker-pubsubplus-p", "", ""),
		obj("PersistentVolumeClaim", "data-dev-broker-pubsubplus-p-0", "", ""),
	), []Object{{"StatefulSet", "dev-broker-pubsubplus-p"}})
	if want := []string{"persistentvolumeclaim/data-dev-broker-pubsubplus-p-0"}; !reflect.DeepEqual(got, want) {
		t.Errorf("Occupants = %v, want %v", got, want)
	}
}

// TestOccupantsOwnershipBlind: with no owned objects, a deployment's own leftovers
// count, as solace-util's gate counted them.
func TestOccupantsOwnershipBlind(t *testing.T) {
	got := occupants(t, listing(obj("Deployment", "solmq-connector", "", "")), nil)
	if want := []string{"deployment/solmq-connector"}; !reflect.DeepEqual(got, want) {
		t.Errorf("Occupants = %v, want %v", got, want)
	}
}

// TestOccupantsRefusesUnreadable carries solmq-conn-util's
// TestNamespaceOccupantsRejectsUnreadableOutput and solace-util's
// TestNamespaceContentsErrorMeansOccupied: anything that cannot be read keeps the
// namespace.
func TestOccupantsRefusesUnreadable(t *testing.T) {
	for name, out := range map[string]string{
		"empty":           "",
		"blank":           "  \n",
		"not JSON":        "No resources found in prod namespace.",
		"no items key":    "{}",
		"null items":      `{"items": null}`,
		"a single object": obj("Deployment", "x", "", ""),
		"no kind":         `{"items": [{"metadata": {"name": "x"}}]}`,
		"no name":         `{"items": [{"kind": "Pod", "metadata": {}}]}`,
	} {
		got, err := Occupants(out, nil)
		if err == nil {
			t.Errorf("%s: Occupants = %v, want an error", name, got)
			continue
		}
		if !strings.Contains(err.Error(), "kept") {
			t.Errorf("%s: err = %v, want it to say the namespace is kept", name, err)
		}
	}
}
