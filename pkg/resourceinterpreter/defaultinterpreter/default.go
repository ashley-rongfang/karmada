package defaultinterpreter

import (
	"fmt"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/klog/v2"

	configv1alpha1 "github.com/karmada-io/karmada/pkg/apis/config/v1alpha1"
	workv1alpha2 "github.com/karmada-io/karmada/pkg/apis/work/v1alpha2"
)

// DefaultInterpreter contains all default operation interpreter factory
// for interpreting common resource.
type DefaultInterpreter struct {
	replicaHandlers       map[schema.GroupVersionKind]replicaInterpreter
	reviseReplicaHandlers map[schema.GroupVersionKind]reviseReplicaInterpreter
	retentionHandlers     map[schema.GroupVersionKind]retentionInterpreter
}

// NewDefaultInterpreter return a new DefaultInterpreter.
func NewDefaultInterpreter() *DefaultInterpreter {
	return &DefaultInterpreter{
		replicaHandlers:       getAllDefaultReplicaInterpreter(),
		reviseReplicaHandlers: getAllDefaultReviseReplicaInterpreter(),
		retentionHandlers:     getAllDefaultRetentionInterpreter(),
	}
}

// HookEnabled tells if any hook exist for specific resource type and operation type.
func (e *DefaultInterpreter) HookEnabled(kind schema.GroupVersionKind, operationType configv1alpha1.InterpreterOperation) bool {
	switch operationType {
	case configv1alpha1.InterpreterOperationInterpretReplica:
		if _, exist := e.replicaHandlers[kind]; exist {
			return true
		}
	case configv1alpha1.InterpreterOperationReviseReplica:
		if _, exist := e.reviseReplicaHandlers[kind]; exist {
			return true
		}
	case configv1alpha1.InterpreterOperationRetain:
		if _, exist := e.retentionHandlers[kind]; exist {
			return true
		}

		// TODO(RainbowMango): more cases should be added here
	}

	klog.V(4).Infof("Default interpreter is not enabled for kind %q with operation %q.", kind, operationType)
	return false
}

// GetReplicas returns the desired replicas of the object as well as the requirements of each replica.
func (e *DefaultInterpreter) GetReplicas(object *unstructured.Unstructured) (int32, *workv1alpha2.ReplicaRequirements, error) {
	handler, exist := e.replicaHandlers[object.GroupVersionKind()]
	if !exist {
		return 0, &workv1alpha2.ReplicaRequirements{}, fmt.Errorf("defalut %s interpreter for %q not found", configv1alpha1.InterpreterOperationInterpretReplica, object.GroupVersionKind())
	}
	return handler(object)
}

// ReviseReplica revises the replica of the given object.
func (e *DefaultInterpreter) ReviseReplica(object *unstructured.Unstructured, replica int64) (*unstructured.Unstructured, error) {
	handler, exist := e.reviseReplicaHandlers[object.GroupVersionKind()]
	if !exist {
		return nil, fmt.Errorf("defalut %s interpreter for %q not found", configv1alpha1.InterpreterOperationReviseReplica, object.GroupVersionKind())
	}
	return handler(object, replica)
}

// Retain returns the objects that based on the "desired" object but with values retained from the "observed" object.
func (e *DefaultInterpreter) Retain(desired *unstructured.Unstructured, observed *unstructured.Unstructured) (retained *unstructured.Unstructured, err error) {
	handler, exist := e.retentionHandlers[desired.GroupVersionKind()]
	if !exist {
		return nil, fmt.Errorf("default %s interpreter for %q not found", configv1alpha1.InterpreterOperationRetain, desired.GroupVersionKind())
	}

	return handler(desired, observed)
}
