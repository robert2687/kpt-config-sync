// Copyright 2022 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package root

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	"kpt.dev/configsync/pkg/api/kpt.dev/v1alpha1"
	"kpt.dev/configsync/pkg/resourcegroup/controllers/resourcemap"
	"kpt.dev/configsync/pkg/resourcegroup/controllers/watch"
	// +kubebuilder:scaffold:imports
)

// These tests use Ginkgo (BDD-style Go testing framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

var cfg *rest.Config
var testEnv *envtest.Environment

func TestMain(m *testing.M) {
	testEnv = &envtest.Environment{
		CRDDirectoryPaths: []string{filepath.Join("..", "..", "..", "..", "manifests")},
	}

	var err error
	cfg, err = testEnv.Start()
	if err != nil {
		log.Fatal(err)
	}

	s := scheme.Scheme
	if err := v1alpha1.AddToScheme(s); err != nil {
		log.Fatal(err)
	}
	if err := apiextensionsv1.AddToScheme(s); err != nil {
		log.Fatal(err)
	}

	code := m.Run()

	err = testEnv.Stop()
	if err != nil {
		log.Printf("Error: Failed to stop test env: %v", err)
	}

	os.Exit(code)
}

// StartTestManager adds recFn
func StartTestManager(t *testing.T, mgr manager.Manager) {
	go func() {
		err := mgr.Start(context.Background())
		assert.NoError(t, err)
	}()
}

func NewReconciler(mgr manager.Manager) (*Reconciler, error) {
	resmap := resourcemap.NewResourceMap()
	watches, err := watch.NewManager(mgr.GetConfig(), resmap, nil, nil)
	if err != nil {
		return nil, err
	}
	r := &Reconciler{
		Client:  mgr.GetClient(),
		cfg:     mgr.GetConfig(),
		log:     ctrl.Log.WithName("controllers").WithName("Root"),
		resMap:  resmap,
		watches: watches,
	}
	obj := &v1alpha1.ResourceGroup{}
	_, err = ctrl.NewControllerManagedBy(mgr).
		For(obj).
		Build(r)
	if err != nil {
		return nil, err
	}
	return r, nil
}
