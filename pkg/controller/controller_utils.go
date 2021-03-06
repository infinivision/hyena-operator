// Copyright 2018 infinivision, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// See the License for the specific language governing permissions and
// limitations under the License.

package controller

import (
	"fmt"
	"math"

	"github.com/golang/glog"
	"github.com/infinivision/hyena-operator/pkg/apis/infinivision.com/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	// controllerKind contains the schema.GroupVersionKind for this controller type.
	controllerKind = v1alpha1.SchemeGroupVersion.WithKind("HyenaCluster")
	// DefaultStorageClassName is the default storageClassName
	DefaultStorageClassName string
	// ClusterScoped controls whether operator should manage kubernetes cluster wide Hyena clusters
	ClusterScoped bool
)

// RequeueError is used to requeue the item, this error type should't be considered as a real error
type RequeueError struct {
	s string
}

func (re *RequeueError) Error() string {
	return re.s
}

// RequeueErrorf returns a RequeueError
func RequeueErrorf(format string, a ...interface{}) error {
	return &RequeueError{fmt.Sprintf(format, a...)}
}

// IsRequeueError returns whether err is a RequeueError
func IsRequeueError(err error) bool {
	_, ok := err.(*RequeueError)
	return ok
}

// GetOwnerRef returns HyenaCluster's OwnerReference
func GetOwnerRef(hc *v1alpha1.HyenaCluster) metav1.OwnerReference {
	controller := true
	blockOwnerDeletion := true
	return metav1.OwnerReference{
		APIVersion:         controllerKind.GroupVersion().String(),
		Kind:               controllerKind.Kind,
		Name:               hc.GetName(),
		UID:                hc.GetUID(),
		Controller:         &controller,
		BlockOwnerDeletion: &blockOwnerDeletion,
	}
}

// GetServiceType returns member's service type
func GetServiceType(services []v1alpha1.Service, serviceName string) corev1.ServiceType {
	for _, svc := range services {
		if svc.Name == serviceName {
			switch svc.Type {
			case "NodePort":
				return corev1.ServiceTypeNodePort
			case "LoadBalancer":
				return corev1.ServiceTypeLoadBalancer
			default:
				return corev1.ServiceTypeClusterIP
			}
		}
	}
	return corev1.ServiceTypeClusterIP
}

// StoreCapacity returns string resource requirement,
// store uses GB, TB as unit suffix, but it actually means GiB, TiB
func StoreCapacity(limits *v1alpha1.ResourceRequirement) string {
	defaultArgs := "0"
	if limits == nil || limits.Storage == "" {
		return defaultArgs
	}
	q, err := resource.ParseQuantity(limits.Storage)
	if err != nil {
		glog.Errorf("failed to parse quantity %s: %v", limits.Storage, err)
		return defaultArgs
	}
	i, b := q.AsInt64()
	if !b {
		glog.Errorf("quantity %s can't be converted to int64", q.String())
		return defaultArgs
	}
	return fmt.Sprintf("%dGB", int(float64(i)/math.Pow(2, 30)))
}

// DefaultPushGatewayRequest for the Store sidecar
func DefaultPushGatewayRequest() corev1.ResourceRequirements {
	rr := corev1.ResourceRequirements{}
	rr.Requests = make(map[corev1.ResourceName]resource.Quantity)
	rr.Limits = make(map[corev1.ResourceName]resource.Quantity)
	rr.Requests[corev1.ResourceCPU] = resource.MustParse("50m")
	rr.Requests[corev1.ResourceMemory] = resource.MustParse("50Mi")
	rr.Limits[corev1.ResourceCPU] = resource.MustParse("100m")
	rr.Limits[corev1.ResourceMemory] = resource.MustParse("100Mi")
	return rr
}

// ProphetMemberName returns pd member name
func ProphetMemberName(clusterName string) string {
	return fmt.Sprintf("%s-prophet", clusterName)
}

// ProphetPeerMemberName returns prophet peer service name
func ProphetPeerMemberName(clusterName string) string {
	return fmt.Sprintf("%s-prophet-peer", clusterName)
}

// StoreMemberName returns store member name
func StoreMemberName(clusterName string) string {
	return fmt.Sprintf("%s-store", clusterName)
}

// StorePeerName returns store peer name
func StorePeerName(clusterName string) string {
	return fmt.Sprintf("%s-store-peer", clusterName)
}

// setIfNotEmpty set the value into map when value in not empty
func setIfNotEmpty(container map[string]string, key, value string) {
	if value != "" {
		container[key] = value
	}
}

// requestTracker is used by unit test for mocking request error
type requestTracker struct {
	requests int
	err      error
	after    int
}

func (rt *requestTracker) errorReady() bool {
	return rt.err != nil && rt.requests >= rt.after
}

func (rt *requestTracker) inc() {
	rt.requests++
}

func (rt *requestTracker) reset() {
	rt.err = nil
	rt.after = 0
}
