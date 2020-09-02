/*
Copyright 2020 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package hostpath

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/openshift/csi-driver-projected-resource/pkg/cache"
)

func testHostPathDriver() (*hostPath, string, error) {
	tmpDir, err := ioutil.TempDir(os.TempDir(), "ut")
	if err != nil {
		return nil, "", err
	}
	hp, err := NewHostPathDriver(tmpDir, "ut-driver", "nodeID1", "endpoint1", 0, "version1")
	return hp, tmpDir, err
}

func TestCreateHostPathVolumeBadAccessType(t *testing.T) {
	hp, dir, err := testHostPathDriver()
	if err != nil {
		t.Fatalf("%s", err.Error())
	}
	defer os.RemoveAll(dir)
	volPath := hp.getVolumePath("volID", "podNamespace", "podName", "podUID", "podSA")
	_, err = hp.createHostpathVolume("volID", volPath, "", 0, mountAccess+1)
	if err == nil {
		t.Fatalf("err nil unexpectedly")
	}
	if !strings.Contains(err.Error(), "unsupported access type") {
		t.Fatalf("unexpected err %s", err.Error())
	}
}

func TestCreateConfigMapHostPathVolume(t *testing.T) {
	hp, dir, err := testHostPathDriver()
	if err != nil {
		t.Fatalf("%s", err.Error())
	}
	defer os.RemoveAll(dir)
	targetPath, err := ioutil.TempDir(os.TempDir(), "ut")
	if err != nil {
		t.Fatalf("err on targetPath %s", err.Error())
	}
	defer os.RemoveAll(targetPath)
	cm := primeConfigMapVolume(hp, targetPath, t)
	_, foundConfigMap := findSharedItems(targetPath, t)
	if !foundConfigMap {
		t.Fatalf("did not find configmap in mount path")
	}
	cache.DelConfigMap(cm)
	_, foundConfigMap = findSharedItems(targetPath, t)
	if foundConfigMap {
		t.Fatalf("configmap not deleted")
	}
}

func TestCreateSecretHostPathVolume(t *testing.T) {
	hp, dir, err := testHostPathDriver()
	if err != nil {
		t.Fatalf("%s", err.Error())
	}
	defer os.RemoveAll(dir)
	targetPath, err := ioutil.TempDir(os.TempDir(), "ut")
	if err != nil {
		t.Fatalf("err on targetPath %s", err.Error())
	}
	defer os.RemoveAll(targetPath)
	secret := primeSecretVolume(hp, targetPath, t)
	foundSecret, _ := findSharedItems(targetPath, t)
	if !foundSecret {
		t.Fatalf("did not find secret in mount path")
	}
	cache.DelSecret(secret)
	foundSecret, _ = findSharedItems(targetPath, t)
	if foundSecret {
		t.Fatalf("secret not deleted")
	}
}

func TestDeleteSecretVolume(t *testing.T) {
	hp, dir, err := testHostPathDriver()
	if err != nil {
		t.Fatalf("%s", err.Error())
	}
	defer os.RemoveAll(dir)
	targetPath, err := ioutil.TempDir(os.TempDir(), "ut")
	if err != nil {
		t.Fatalf("err on targetPath %s", err.Error())
	}
	defer os.RemoveAll(targetPath)
	primeSecretVolume(hp, targetPath, t)
	err = hp.deleteHostpathVolume("volID")
	if err != nil {
		t.Fatalf("unexpeted error on delete volume: %s", err.Error())
	}
	foundSecret, _ := findSharedItems(dir, t)

	if foundSecret {
		t.Fatalf("secret not deleted")
	}
	if empty, err := isDirEmpty(dir); !empty || err != nil {
		t.Fatalf("volume directory not cleaned out empty %v err %s", empty, err.Error())
	}

}

func TestDeleteConfigMapVolume(t *testing.T) {
	hp, dir, err := testHostPathDriver()
	if err != nil {
		t.Fatalf("%s", err.Error())
	}
	defer os.RemoveAll(dir)
	targetPath, err := ioutil.TempDir(os.TempDir(), "ut")
	if err != nil {
		t.Fatalf("err on targetPath %s", err.Error())
	}
	defer os.RemoveAll(targetPath)
	primeConfigMapVolume(hp, targetPath, t)
	err = hp.deleteHostpathVolume("volID")
	if err != nil {
		t.Fatalf("unexpeted error on delete volume: %s", err.Error())
	}
	_, foundConfigMap := findSharedItems(dir, t)

	if foundConfigMap {
		t.Fatalf("configmap not deleted")
	}
	if empty, err := isDirEmpty(dir); !empty || err != nil {
		t.Fatalf("volume directory not cleaned out empty %v err %s", empty, err.Error())
	}

}

func primeSecretVolume(hp *hostPath, targetPath string, t *testing.T) *corev1.Secret {
	volPath := hp.getVolumePath("volID", "podNamespace", "podName", "podUID", "podSA")
	hpv, err := hp.createHostpathVolume("volID", volPath, targetPath, 0, mountAccess)
	if err != nil {
		t.Fatalf("unexpected err %s", err.Error())
	}
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "secret1",
			Namespace: "namespace",
		},
	}

	hpv.SharedDataKind = "Secret"
	hpv.SharedDataKey = cache.BuildKey(secret.Namespace, secret.Name)

	cache.UpsertSecret(secret)
	err = hp.mapVolumeToPod(hpv)
	if err != nil {
		t.Fatalf("unexpected err %s", err.Error())
	}
	return secret
}

func primeConfigMapVolume(hp *hostPath, targetPath string, t *testing.T) *corev1.ConfigMap {
	volPath := hp.getVolumePath("volID", "podNamespace", "podName", "podUID", "podSA")
	hpv, err := hp.createHostpathVolume("volID", volPath, targetPath, 0, mountAccess)
	if err != nil {
		t.Fatalf("unexpected err %s", err.Error())
	}
	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "configmap1",
			Namespace: "namespace",
		},
	}

	hpv.SharedDataKind = "ConfigMap"
	hpv.SharedDataKey = cache.BuildKey(cm.Namespace, cm.Name)

	cache.UpsertConfigMap(cm)
	err = hp.mapVolumeToPod(hpv)
	if err != nil {
		t.Fatalf("unexpected err %s", err.Error())
	}
	return cm
}

func findSharedItems(dir string, t *testing.T) (bool, bool) {
	foundSecret := false
	foundConfigMap := false
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		t.Logf("found file %s dir flag %v", info.Name(), info.IsDir())
		if err == nil && strings.Contains(info.Name(), "secret1") {
			foundSecret = true
		}
		if err == nil && strings.Contains(info.Name(), "configmap1") {
			foundConfigMap = true
		}
		return nil
	})
	if err != nil {
		t.Fatalf("unexpected walk error: %s", err.Error())
	}
	return foundSecret, foundConfigMap
}
