package ebpf

import (
	"context"
	"fmt"
	"strconv"

	"github.com/netobserv/network-observability-operator/controllers/ebpf/internal/permissions"
	"github.com/netobserv/network-observability-operator/pkg/discover"
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/log"

	flowsv1alpha1 "github.com/netobserv/network-observability-operator/api/v1alpha1"
	"github.com/netobserv/network-observability-operator/controllers/constants"
	"github.com/netobserv/network-observability-operator/controllers/reconcilers"
	"github.com/netobserv/network-observability-operator/pkg/helper"
)

const (
	flowsTargetHostEnvVar = "FLOWS_TARGET_HOST"
	flowsTargetPortEnvVar = "FLOWS_TARGET_PORT"
)

type reconcileAction int

const (
	actionNone = iota
	actionCreate
	actionUpdate
)

// AgentController reconciles the status of the eBPF agent Daemonset, as well as the
// associated objects that are required to bind the proper permissions: namespace, service
// accounts, SecurityContextConstraints...
type AgentController struct {
	client              reconcilers.ClientHelper
	baseNamespace       string
	privilegedNamespace string
	permissions         permissions.Reconciler
}

func NewAgentController(
	client reconcilers.ClientHelper,
	baseNamespace string,
	permissionsVendor *discover.Permissions,
) *AgentController {
	pns := baseNamespace + constants.EBPFPrivilegedNSSuffix
	return &AgentController{
		client:              client,
		baseNamespace:       baseNamespace,
		privilegedNamespace: pns,
		permissions:         permissions.NewReconciler(client, pns, permissionsVendor),
	}
}

func (c *AgentController) Reconcile(
	ctx context.Context, target *flowsv1alpha1.FlowCollector) error {
	rlog := log.FromContext(ctx).WithName("AgentController")
	ctx = log.IntoContext(ctx, rlog)

	if err := c.permissions.Reconcile(ctx); err != nil {
		return fmt.Errorf("reconciling permissions: %w", err)
	}
	current, err := c.current(ctx)
	if err != nil {
		return fmt.Errorf("can't fetch current EBPF Agent: %w", err)
	}
	desired := c.desired(target)
	switch c.requiredAction(current, desired) {
	case actionCreate:
		rlog.Info("action: create agent")
		return c.client.CreateOwned(ctx, desired)
	case actionUpdate:
		rlog.Info("action: update agent")
		return c.client.UpdateOwned(ctx, current, desired)
	default:
		rlog.Info("action: nonthing to do")
		return nil
	}
}

func (c *AgentController) current(ctx context.Context) (*v1.DaemonSet, error) {
	agentDS := v1.DaemonSet{}
	if err := c.client.Get(ctx, types.NamespacedName{
		Name:      constants.EBPFAgentName,
		Namespace: c.privilegedNamespace,
	}, &agentDS); err != nil {
		if errors.IsNotFound(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("can't read DaemonSet %s/%s: %w",
			c.privilegedNamespace, constants.EBPFAgentName, err)
	}
	return &agentDS, nil
}

func (c *AgentController) desired(coll *flowsv1alpha1.FlowCollector) *v1.DaemonSet {
	if coll == nil || coll.Spec.EBPF == nil {
		return nil
	}
	trueVal := true
	version := helper.ExtractVersion(coll.Spec.EBPF.Image)
	return &v1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      constants.EBPFAgentName,
			Namespace: c.privilegedNamespace,
			Labels: map[string]string{
				"app":     constants.EBPFAgentName,
				"version": version,
			},
		},
		Spec: v1.DaemonSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"app": constants.EBPFAgentName},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"app": constants.EBPFAgentName},
				},
				Spec: corev1.PodSpec{
					ServiceAccountName: constants.EBPFServiceAccount,
					HostNetwork:        true,
					DNSPolicy:          corev1.DNSClusterFirstWithHostNet,
					Containers: []corev1.Container{{
						Name:            constants.EBPFAgentName,
						Image:           coll.Spec.EBPF.Image,
						ImagePullPolicy: corev1.PullPolicy(coll.Spec.EBPF.ImagePullPolicy),
						Resources:       coll.Spec.EBPF.Resources,
						// TODO: other parameters when NETOBSERV-201 is implemented
						SecurityContext: &corev1.SecurityContext{
							Privileged: &trueVal,
						},
						Env: c.flpEndpoint(coll),
					}},
				},
			},
		},
	}
}

func (c *AgentController) flpEndpoint(coll *flowsv1alpha1.FlowCollector) []corev1.EnvVar {
	switch coll.Spec.FlowlogsPipeline.Kind {
	case constants.DaemonSetKind:
		// When flowlogs-pipeline is deployed as a daemonset, each agent must send
		// data to the pod that is deployed in the same host
		return []corev1.EnvVar{{
			Name: flowsTargetHostEnvVar,
			ValueFrom: &corev1.EnvVarSource{
				FieldRef: &corev1.ObjectFieldSelector{
					FieldPath: "status.hostIP",
				},
			},
		}, {
			Name:  flowsTargetPortEnvVar,
			Value: strconv.Itoa(int(coll.Spec.FlowlogsPipeline.Port)),
		}}
	case constants.DeploymentKind:
		return []corev1.EnvVar{{
			Name:  flowsTargetHostEnvVar,
			Value: constants.FLPName + "." + c.baseNamespace,
		}, {
			Name:  flowsTargetPortEnvVar,
			Value: strconv.Itoa(int(coll.Spec.FlowlogsPipeline.Port)),
		}}
	}
	return nil
}

func (c *AgentController) requiredAction(current, desired *v1.DaemonSet) reconcileAction {
	if desired == nil {
		return actionNone
	}
	if current == nil && desired != nil {
		return actionCreate
	}
	dspec, cspec := &desired.Spec.Template.Spec, &current.Spec.Template.Spec
	equal := helper.IsSubSet(current.ObjectMeta.Labels, desired.ObjectMeta.Labels) &&
		dspec.ServiceAccountName == cspec.ServiceAccountName &&
		dspec.HostNetwork == cspec.HostNetwork &&
		dspec.DNSPolicy == cspec.DNSPolicy &&
		len(dspec.Containers) == len(cspec.Containers)
	if equal {
		return actionNone
	}
	return actionUpdate
}