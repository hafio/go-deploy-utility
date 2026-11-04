// Package nsgate decides whether a namespace a deployment leaves behind may be
// offered for deletion: only when nothing but this deployment's own leftovers and
// what Kubernetes puts in every namespace remains.
//
// The gate is a convenience, so every way out of it short of a confirmed delete keeps
// the namespace, and none of them fails the remove that ran it. A tool runs it in
// this order, keeping the namespace with a warning at the first step that does not
// pass:
//
//  1. Protected: the cluster's own namespaces are never deleted.
//  2. List, then Occupants: a listing that is refused, fails or cannot be read keeps
//     the namespace, and so does any occupant, which the warning names.
//  3. rbac.Check of DeleteProbe, before asking, so nobody is asked and then refused.
//  4. Confirm: --no-prompt approves; without a terminal the namespace is kept.
//  5. kube.DeleteNamespace; a failure is a warning too.
package nsgate

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/hafio/go-deploy-utility/kube/rbac"
	"github.com/hafio/go-deploy-utility/runner"
)

// Object names one object a deployment created: its kind (any case) and name.
type Object struct{ Kind, Name string }

// matches reports whether the object is kind/name.
func (o Object) matches(kind, name string) bool {
	return strings.EqualFold(o.Kind, kind) && o.Name == name
}

// Protected reports whether ns must never be deleted: the cluster's own namespaces,
// and an empty name, which names nothing that could be checked.
func Protected(ns string) bool {
	switch ns {
	case "", "default", "kube-system", "kube-public", "kube-node-lease":
		return true
	}
	return false
}

// List asks what is in ns in one call: `get
// all,persistentvolumeclaims,secrets,configmaps -n <ns> -o json`. It has no
// --ignore-not-found, with which kubectl prints nothing at all for an empty list; a
// namespace that is already gone lists as an empty one. base is the guarded kubectl
// argv, and it is copied.
func List(base []string, ns string) runner.Cmd {
	argv := make([]string, 0, len(base)+6)
	argv = append(argv, base...)
	return runner.Cmd{Argv: append(argv, "get", "all,persistentvolumeclaims,secrets,configmaps", "-n", ns,
		"-o", "json")}
}

// DeleteProbe is the permission the delete needs, checked before the operator is
// asked.
func DeleteProbe() rbac.Probe { return rbac.Probe{Verb: "delete", Resource: "namespaces"} }

type ref struct{ Kind, Name, UID string }

type item struct {
	Kind     string
	Type     string
	Metadata struct {
		Name, UID         string
		DeletionTimestamp string
		Annotations       map[string]string
		OwnerReferences   []ref
	}
}

// Occupants reads List's answer and returns what still occupies the namespace, as
// sorted kind/name with the kind in lower case. None of these counts:
//
//   - an object being deleted (it has a deletionTimestamp);
//   - an Event, LimitRange or ResourceQuota, which describe the namespace;
//   - what Kubernetes puts in every namespace: ConfigMap kube-root-ca.crt,
//     ServiceAccount default, and the default service account's token Secret;
//   - an object in owned, the deployment's own objects by kind and name;
//   - an object whose ownerReference names an object in owned, by kind and name, or
//     one already left out, by uid when both carry one and by kind and name
//     otherwise. This is followed to the end, so a Deployment's ReplicaSet and its
//     Pods, or a custom resource's StatefulSet and its Pods, are left out with it,
//     even when the root itself is not listed.
//
// A nil owned counts every object of a deployment too. An answer that is empty, not
// JSON, has no items, or holds an object with no kind or name is an error: the
// namespace is kept rather than judged from what could not be read.
func Occupants(out string, owned []Object) ([]string, error) {
	out = strings.TrimSpace(out)
	if out == "" {
		return nil, errors.New("nsgate: the namespace listing came back empty, so what is in the namespace " +
			"is unknown; it is kept")
	}
	var doc struct{ Items *[]item }
	if err := json.Unmarshal([]byte(out), &doc); err != nil {
		return nil, fmt.Errorf("nsgate: the namespace listing is not JSON (%w); it is kept", err)
	}
	if doc.Items == nil {
		return nil, errors.New("nsgate: the namespace listing has no items, so what is in the namespace is " +
			"unknown; it is kept")
	}
	items := *doc.Items
	gone := make([]bool, len(items))
	for i, it := range items {
		if it.Kind == "" || it.Metadata.Name == "" {
			return nil, fmt.Errorf("nsgate: the namespace listing holds an object with no kind or name "+
				"(item %d); it is kept", i)
		}
		gone[i] = discounted(it, owned)
	}
	// Follow ownership until nothing more changes: each pass can only add to gone.
	for changed := true; changed; {
		changed = false
		for i, it := range items {
			if !gone[i] && ownedBy(it, items, gone, owned) {
				gone[i], changed = true, true
			}
		}
	}
	var left []string
	for i, it := range items {
		if !gone[i] {
			left = append(left, strings.ToLower(it.Kind)+"/"+it.Metadata.Name)
		}
	}
	sort.Strings(left)
	return left, nil
}

// discounted reports whether it is left out on its own account.
func discounted(it item, owned []Object) bool {
	m := it.Metadata
	switch {
	case m.DeletionTimestamp != "":
		return true
	case it.Kind == "Event" || it.Kind == "LimitRange" || it.Kind == "ResourceQuota":
		return true
	case it.Kind == "ConfigMap" && m.Name == "kube-root-ca.crt",
		it.Kind == "ServiceAccount" && m.Name == "default":
		return true
	case it.Kind == "Secret" && it.Type == "kubernetes.io/service-account-token" &&
		m.Annotations["kubernetes.io/service-account.name"] == "default":
		return true
	}
	for _, o := range owned {
		if o.matches(it.Kind, m.Name) {
			return true
		}
	}
	return false
}

// ownedBy reports whether one of it's owners is in owned or already left out.
func ownedBy(it item, items []item, gone []bool, owned []Object) bool {
	for _, r := range it.Metadata.OwnerReferences {
		for _, o := range owned {
			if o.matches(r.Kind, r.Name) {
				return true
			}
		}
		for j, owner := range items {
			if !gone[j] {
				continue
			}
			if r.UID != "" && owner.Metadata.UID != "" {
				if r.UID == owner.Metadata.UID {
					return true
				}
				continue
			}
			if strings.EqualFold(r.Kind, owner.Kind) && r.Name == owner.Metadata.Name {
				return true
			}
		}
	}
	return false
}
