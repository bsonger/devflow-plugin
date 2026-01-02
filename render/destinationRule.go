package render

import (
	"github.com/bsonger/devflow-common/model"
	v1alpha3 "istio.io/api/networking/v1alpha3"
	networkingv1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"
)

// DestinationRule 根据 Manifest 生成 DestinationRule YAML
func DestinationRule(m *model.Manifest, env string) (string, error) {
	if m.Type != model.Canary {
		return "", nil
	}
	dr := &networkingv1beta1.DestinationRule{
		TypeMeta: metav1.TypeMeta{
			Kind:       "DestinationRule",
			APIVersion: "networking.istio.io/v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: m.ApplicationName,
			Labels: map[string]string{
				"app": m.ApplicationName,
				"env": env,
			},
		},
	}

	host := m.ApplicationName // ✅ 单 Service 名

	// 根据类型生成 Subset
	var subsets []*v1alpha3.Subset
	switch m.Type {
	case model.Canary:
		subsets = []*v1alpha3.Subset{
			{
				Name:   "stable",
				Labels: map[string]string{}, // rollouts-pod-template-hash Argo Rollouts 自动 patch
			},
			{
				Name:   "canary",
				Labels: map[string]string{},
			},
		}
	case model.BlueGreen:
		// 使用同一个 Service，subset 用 active / preview（或 blue / green）
		subsets = []*v1alpha3.Subset{
			{
				Name:   "active",
				Labels: map[string]string{}, // rollouts-pod-template-hash 自动 patch
			},
			{
				Name:   "preview",
				Labels: map[string]string{},
			},
		}
	default: // normal / rolling
		subsets = []*v1alpha3.Subset{
			{
				Name:   "stable",
				Labels: map[string]string{},
			},
		}
	}

	dr.Spec = v1alpha3.DestinationRule{
		Host:    host,
		Subsets: subsets,
	}

	yml, err := yaml.Marshal(dr)
	if err != nil {
		return "", err
	}
	return string(yml), nil
}
