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

package hyenacluster

import (
	"testing"
	"time"

	"github.com/infinivision/hyena-operator/pkg/apis/infinivision.com/v1alpha1"
	"github.com/infinivision/hyena-operator/pkg/client/clientset/versioned/fake"
	informers "github.com/infinivision/hyena-operator/pkg/client/informers/externalversions"
	"github.com/infinivision/hyena-operator/pkg/controller"
	mm "github.com/infinivision/hyena-operator/pkg/manager/member"
	"github.com/infinivision/hyena-operator/pkg/manager/meta"
	. "github.com/onsi/gomega"
	apps "k8s.io/api/apps/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	kubeinformers "k8s.io/client-go/informers"
	kubefake "k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
)

func TestHyenaClusterControllerEnqueueHyenaCluster(t *testing.T) {
	g := NewGomegaWithT(t)
	hc := newHyenaCluster()
	hcc, _, _ := newFakeHyenaClusterController()

	hcc.enqueueHyenaCluster(hc)
	g.Expect(hcc.queue.Len()).To(Equal(1))
}

func TestHyenaClusterControllerAddStatefuSet(t *testing.T) {
	g := NewGomegaWithT(t)
	type testcase struct {
		name                     string
		modifySet                func(*v1alpha1.HyenaCluster) *apps.StatefulSet
		addHyenaClusterToIndexer bool
		expectedLen              int
	}

	testFn := func(test *testcase, t *testing.T) {
		cc := newHyenaCluster()
		set := test.modifySet(cc)

		ccc, ccIndexer, _ := newFakeHyenaClusterController()

		if test.addHyenaClusterToIndexer {
			err := ccIndexer.Add(cc)
			g.Expect(err).NotTo(HaveOccurred())
		}
		ccc.addStatefulSet(set)
		g.Expect(ccc.queue.Len()).To(Equal(test.expectedLen))
	}

	tests := []testcase{
		{
			name: "normal",
			modifySet: func(cc *v1alpha1.HyenaCluster) *apps.StatefulSet {
				return newStatefuSet(cc)
			},
			addHyenaClusterToIndexer: true,
			expectedLen:              1,
		},
		{
			name: "have deletionTimestamp",
			modifySet: func(cc *v1alpha1.HyenaCluster) *apps.StatefulSet {
				set := newStatefuSet(cc)
				set.DeletionTimestamp = &metav1.Time{Time: time.Now().Add(30 * time.Second)}
				return set
			},
			addHyenaClusterToIndexer: true,
			expectedLen:              1,
		},
		{
			name: "without controllerRef",
			modifySet: func(cc *v1alpha1.HyenaCluster) *apps.StatefulSet {
				set := newStatefuSet(cc)
				set.OwnerReferences = nil
				return set
			},
			addHyenaClusterToIndexer: true,
			expectedLen:              0,
		},
		{
			name: "without HyenaCluster",
			modifySet: func(cc *v1alpha1.HyenaCluster) *apps.StatefulSet {
				return newStatefuSet(cc)
			},
			addHyenaClusterToIndexer: false,
			expectedLen:              0,
		},
	}

	for i := range tests {
		testFn(&tests[i], t)
	}
}

func TestHyenaClusterControllerUpdateStatefuSet(t *testing.T) {
	g := NewGomegaWithT(t)
	type testcase struct {
		name                     string
		initial                  func() *v1alpha1.HyenaCluster
		initialSet               func(*v1alpha1.HyenaCluster) *apps.StatefulSet
		updateSet                func(*apps.StatefulSet) *apps.StatefulSet
		addHyenaClusterToIndexer bool
		expectedLen              int
	}

	testFn := func(test *testcase, t *testing.T) {
		cc := test.initial()
		set1 := test.initialSet(cc)
		set2 := test.updateSet(set1)

		ccc, ccIndexer, _ := newFakeHyenaClusterController()

		if test.addHyenaClusterToIndexer {
			err := ccIndexer.Add(cc)
			g.Expect(err).NotTo(HaveOccurred())
		}
		ccc.updateStatefuSet(set1, set2)
		g.Expect(ccc.queue.Len()).To(Equal(test.expectedLen))
	}

	tests := []testcase{
		{
			name: "normal",
			initial: func() *v1alpha1.HyenaCluster {
				return newHyenaCluster()
			},
			initialSet: func(cc *v1alpha1.HyenaCluster) *apps.StatefulSet {
				return newStatefuSet(cc)
			},
			updateSet: func(set1 *apps.StatefulSet) *apps.StatefulSet {
				set2 := *set1
				set2.ResourceVersion = "1000"
				return &set2
			},
			addHyenaClusterToIndexer: true,
			expectedLen:              1,
		},
		{
			name: "same resouceVersion",
			initial: func() *v1alpha1.HyenaCluster {
				return newHyenaCluster()
			},
			initialSet: func(cc *v1alpha1.HyenaCluster) *apps.StatefulSet {
				return newStatefuSet(cc)
			},
			updateSet: func(set1 *apps.StatefulSet) *apps.StatefulSet {
				set2 := *set1
				return &set2
			},
			addHyenaClusterToIndexer: true,
			expectedLen:              0,
		},
		{
			name: "without controllerRef",
			initial: func() *v1alpha1.HyenaCluster {
				return newHyenaCluster()
			},
			initialSet: func(cc *v1alpha1.HyenaCluster) *apps.StatefulSet {
				return newStatefuSet(cc)
			},
			updateSet: func(set1 *apps.StatefulSet) *apps.StatefulSet {
				set2 := *set1
				set2.ResourceVersion = "1000"
				set2.OwnerReferences = nil
				return &set2
			},
			addHyenaClusterToIndexer: true,
			expectedLen:              0,
		},
		{
			name: "without HyenaCluster",
			initial: func() *v1alpha1.HyenaCluster {
				return newHyenaCluster()
			},
			initialSet: func(cc *v1alpha1.HyenaCluster) *apps.StatefulSet {
				return newStatefuSet(cc)
			},
			updateSet: func(set1 *apps.StatefulSet) *apps.StatefulSet {
				set2 := *set1
				set2.ResourceVersion = "1000"
				return &set2
			},
			addHyenaClusterToIndexer: false,
			expectedLen:              0,
		},
	}

	for i := range tests {
		testFn(&tests[i], t)
	}
}

func alwaysReady() bool { return true }

func newFakeHyenaClusterController() (*Controller, cache.Indexer, cache.Indexer) {
	cli := fake.NewSimpleClientset()
	kubeCli := kubefake.NewSimpleClientset()
	informerFactory := informers.NewSharedInformerFactory(cli, 0)
	kubeInformerFactory := kubeinformers.NewSharedInformerFactory(kubeCli, 0)

	setInformer := kubeInformerFactory.Apps().V1beta1().StatefulSets()
	svcInformer := kubeInformerFactory.Core().V1().Services()
	pvcInformer := kubeInformerFactory.Core().V1().PersistentVolumeClaims()
	pvInformer := kubeInformerFactory.Core().V1().PersistentVolumes()
	tcInformer := informerFactory.Deepfabric().V1alpha1().HyenaClusters()
	podInformer := kubeInformerFactory.Core().V1().Pods()
	nodeInformer := kubeInformerFactory.Core().V1().Nodes()
	autoFailover := true

	ccc := NewController(
		kubeCli,
		cli,
		informerFactory,
		kubeInformerFactory,
	)
	ccc.ccListerSynced = alwaysReady
	ccc.setListerSynced = alwaysReady
	recorder := record.NewFakeRecorder(10)

	pdControl := controller.NewFakePDControl()
	proxyControl := controller.NewFakeProxyControl()
	svcControl := controller.NewRealServiceControl(
		kubeCli,
		svcInformer.Lister(),
		recorder,
	)
	setControl := controller.NewRealStatefuSetControl(
		kubeCli,
		setInformer.Lister(),
		recorder,
	)
	pvControl := controller.NewRealPVControl(kubeCli, pvcInformer.Lister(), pvInformer.Lister(), recorder)
	pvcControl := controller.NewRealPVCControl(kubeCli, recorder, pvcInformer.Lister())
	podControl := controller.NewRealPodControl(kubeCli, pdControl, podInformer.Lister(), recorder)
	storeScaler := mm.NewStoreScaler(pdControl, pvcInformer.Lister(), pvcControl, podInformer.Lister())
	storeFailover := mm.NewFakeStoreFailover()
	pdUpgrader := mm.NewFakePDUpgrader()
	storeUpgrader := mm.NewFakeStoreUpgrader()
	proxyUpgrader := mm.NewFakeProxyUpgrader()

	ccc.control = NewDefaultHyenaClusterControl(
		controller.NewRealHyenaClusterControl(cli, tcInformer.Lister(), recorder),
		mm.NewPDMemberManager(
			pdControl,
			setControl,
			svcControl,
			setInformer.Lister(),
			svcInformer.Lister(),
			podInformer.Lister(),
			podControl,
			pvcInformer.Lister(),
			pdUpgrader,
		),
		mm.NewStoreMemberManager(
			pdControl,
			setControl,
			svcControl,
			setInformer.Lister(),
			svcInformer.Lister(),
			podInformer.Lister(),
			nodeInformer.Lister(),
			autoFailover,
			storeFailover,
			storeScaler,
			storeUpgrader,
		),
		mm.NewProxyMemberManager(
			controller.NewRealStatefuSetControl(
				kubeCli,
				setInformer.Lister(),
				recorder,
			),
			svcControl,
			proxyControl,
			setInformer.Lister(),
			svcInformer.Lister(),
			podInformer.Lister(),
			proxyUpgrader,
		),
		meta.NewReclaimPolicyManager(
			pvcInformer.Lister(),
			pvInformer.Lister(),
			pvControl,
		),
		meta.NewMetaManager(
			pvcInformer.Lister(),
			pvcControl,
			pvInformer.Lister(),
			pvControl,
			podInformer.Lister(),
			podControl,
		),
		mm.NewFakeOrphanPodsCleaner(),
		recorder,
	)

	return ccc, tcInformer.Informer().GetIndexer(), setInformer.Informer().GetIndexer()
}

func newHyenaCluster() *v1alpha1.HyenaCluster {
	return &v1alpha1.HyenaCluster{
		TypeMeta: metav1.TypeMeta{
			Kind:       "HyenaCluster",
			APIVersion: "deepfabric.com/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-prophet",
			Namespace: corev1.NamespaceDefault,
			UID:       types.UID("test"),
		},
		Spec: v1alpha1.HyenaClusterSpec{
			Prophet: v1alpha1.ProphetSpec{
				ContainerSpec: v1alpha1.ContainerSpec{
					Image: "prophet-test-image",
				},
			},
			Store: v1alpha1.StoreSpec{
				ContainerSpec: v1alpha1.ContainerSpec{
					Image: "store-test-image",
				},
			},
		},
	}
}

func newStatefuSet(cc *v1alpha1.HyenaCluster) *apps.StatefulSet {
	return &apps.StatefulSet{
		TypeMeta: metav1.TypeMeta{
			Kind:       "StatefulSet",
			APIVersion: "apps/v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-statefuset",
			Namespace: corev1.NamespaceDefault,
			UID:       types.UID("test"),
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(cc, controllerKind),
			},
			ResourceVersion: "1",
		},
		Spec: apps.StatefulSetSpec{
			Replicas: &cc.Spec.PD.Replicas,
		},
	}
}
