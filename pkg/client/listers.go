package client

import (
	sharev1alpha1 "github.com/openshift/csi-driver-projected-resource/pkg/generated/listers/projectedresource/v1alpha1"
	corev1 "k8s.io/client-go/listers/core/v1"
)

type Listers struct {
	Secrets    corev1.SecretLister
	ConfigMaps corev1.ConfigMapLister
	Shares     sharev1alpha1.ShareLister
}
