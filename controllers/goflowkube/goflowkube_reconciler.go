package goflowkube

import (
	"context"
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	flowsv1alpha1 "github.com/netobserv/network-observability-operator/api/v1alpha1"
)

// Reconciler reconciles the current goflow-kube state with the desired configuration
type Reconciler struct {
	client.Client
	SetControllerReference func(client.Object) error
	OperatorNamespace      string
}

const gfkName = "goflow-kube"
const deploymentKind = "Deployment"
const daemonSetKind = "DaemonSet"
const serviceKind = "Service"

// Reconcile is the reconciler entry point to reconcile the current goflow-kube state with the desired configuration
func (r *Reconciler) Reconcile(ctx context.Context, desired *flowsv1alpha1.FlowCollectorGoflowKube) error {
	// Check if goflow-kube already exists, as a deployment or as a daemon set
	nsname := types.NamespacedName{Name: gfkName, Namespace: r.OperatorNamespace}
	oldDepl, err := r.getObj(ctx, nsname, &appsv1.Deployment{}, deploymentKind)
	if err != nil {
		return err
	}
	oldDS, err := r.getObj(ctx, nsname, &appsv1.DaemonSet{}, daemonSetKind)
	if err != nil {
		return err
	}
	oldSVC, err := r.getObj(ctx, nsname, &corev1.Service{}, serviceKind)
	if err != nil {
		return err
	}

	// If none of them already exist, it must be the first setup. Thus, setup permissions.
	if oldDepl == nil && oldDS == nil {
		r.setupPermissions(ctx)
	}

	switch desired.Kind {
	case deploymentKind:
		// Kind changed: delete DaemonSet and create Deployment+Service
		if oldDS != nil {
			r.delete(ctx, oldDS, daemonSetKind)
		}
		if oldDepl == nil || deploymentNeedsUpdate(oldDepl.(*appsv1.Deployment), desired) {
			newDepl := buildDeployment(desired, r.OperatorNamespace)
			r.createOrUpdate(ctx, oldDepl, newDepl, deploymentKind)
		}
		if oldSVC == nil || serviceNeedsUpdate(oldSVC.(*corev1.Service), desired) {
			newSVC := buildService(desired, r.OperatorNamespace)
			r.createOrUpdate(ctx, oldSVC, newSVC, serviceKind)
		}
	case daemonSetKind:
		// Kind changed: delete Deployment/Service and create DaemonSet
		if oldDepl != nil {
			r.delete(ctx, oldDepl, deploymentKind)
			r.delete(ctx, oldSVC, serviceKind)
		}
		if oldDS != nil && !daemonSetNeedsUpdate(oldDS.(*appsv1.DaemonSet), desired) {
			return nil
		}
		newDS := buildDaemonSet(desired, r.OperatorNamespace)
		r.createOrUpdate(ctx, oldDS, newDS, daemonSetKind)
	default:
		return fmt.Errorf("Could not reconcile collector, invalid kind: %s", desired.Kind)
	}
	return nil
}

func (r *Reconciler) setupPermissions(ctx context.Context) {
	log := log.FromContext(ctx)
	log.Info("Setup permissions for " + gfkName)
	rbacObjects := buildRBAC(r.OperatorNamespace)
	for _, rbacObj := range rbacObjects {
		err := r.SetControllerReference(rbacObj)
		if err != nil {
			log.Error(err, "Failed to set controller reference")
		}
		err = r.Create(ctx, rbacObj)
		if err != nil {
			log.Error(err, "Failed to setup permissions for "+gfkName)
		}
	}
}

func (r *Reconciler) getObj(ctx context.Context, nsname types.NamespacedName, obj client.Object, kind string) (client.Object, error) {
	err := r.Get(ctx, nsname, obj)
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, nil
		} else {
			log.FromContext(ctx).Error(err, "Failed to get "+gfkName+" "+kind)
			return nil, err
		}
	}
	return obj, nil
}

func (r *Reconciler) createOrUpdate(ctx context.Context, old, new client.Object, kind string) {
	log := log.FromContext(ctx)
	err := r.SetControllerReference(new)
	if err != nil {
		log.Error(err, "Failed to set controller reference")
	}
	if old == nil {
		log.Info("Creating a new "+kind, "Namespace", new.GetNamespace(), "Name", new.GetName())
		err := r.Create(ctx, new)
		if err != nil {
			log.Error(err, "Failed to create new "+kind, "Namespace", new.GetNamespace(), "Name", new.GetName())
			return
		}
	} else {
		log.Info("Updating "+kind, "Namespace", new.GetNamespace(), "Name", new.GetName())
		err := r.Update(ctx, new)
		if err != nil {
			log.Error(err, "Failed to update "+kind, "Namespace", new.GetNamespace(), "Name", new.GetName())
			return
		}
	}
}

func (r *Reconciler) delete(ctx context.Context, old client.Object, kind string) {
	log := log.FromContext(ctx)
	log.Info("Deleting old "+kind, "Namespace", old.GetNamespace(), "Name", old.GetName())
	err := r.Delete(ctx, old)
	if err != nil {
		log.Error(err, "Failed to delete old "+kind, "Namespace", old.GetNamespace(), "Name", old.GetName())
	}
}

func daemonSetNeedsUpdate(ds *appsv1.DaemonSet, desired *flowsv1alpha1.FlowCollectorGoflowKube) bool {
	return containerNeedsUpdate(&ds.Spec.Template.Spec, desired)
}

func deploymentNeedsUpdate(depl *appsv1.Deployment, desired *flowsv1alpha1.FlowCollectorGoflowKube) bool {
	return containerNeedsUpdate(&depl.Spec.Template.Spec, desired) ||
		*depl.Spec.Replicas != desired.Replicas
}

func serviceNeedsUpdate(svc *corev1.Service, desired *flowsv1alpha1.FlowCollectorGoflowKube) bool {
	for _, port := range svc.Spec.Ports {
		if port.Port == desired.Port && port.Protocol == "UDP" {
			return false
		}
	}
	return true
}

func containerNeedsUpdate(podSpec *corev1.PodSpec, desired *flowsv1alpha1.FlowCollectorGoflowKube) bool {
	container := findContainer(podSpec)
	if container == nil {
		return true
	}
	if desired.Image != container.Image || desired.ImagePullPolicy != string(container.ImagePullPolicy) {
		return true
	}
	if len(container.Command) != 3 || container.Command[2] != buildMainCommand(desired) {
		return true
	}
	return false
}

func findContainer(podSpec *corev1.PodSpec) *corev1.Container {
	for _, ctnr := range podSpec.Containers {
		if ctnr.Name == gfkName {
			return &ctnr
		}
	}
	return nil
}