package render

import (
	"github.com/bsonger/devflow-common/model"
	v1alpha3 "istio.io/api/networking/v1alpha3"
	networkingv1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"
)

// VirtualService 根据 Manifest 生成 VirtualService YAML
func VirtualService(m *model.Manifest) (string, error) {
	// 内部流量不需要 VirtualService
	if m.Internet == model.Internal {
		return "", nil
	}

	vs := &networkingv1beta1.VirtualService{
		TypeMeta: metav1.TypeMeta{
			Kind:       "VirtualService",
			APIVersion: "networking.istio.io/v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: m.ApplicationName,
			Labels: map[string]string{
				"app": m.ApplicationName,
				"env": string(m.Internet),
			},
		},
		Spec: v1alpha3.VirtualService{
			Hosts: []string{m.ApplicationName},
			Http: []*v1alpha3.HTTPRoute{
				{
					Route: buildHTTPRouteDestinations(m),
				},
			},
		},
	}

	yml, err := yaml.Marshal(vs)
	if err != nil {
		return "", err
	}
	return string(yml), nil
}

// buildHTTPRouteDestinations 构造流量路由
func buildHTTPRouteDestinations(m *model.Manifest) []*v1alpha3.HTTPRouteDestination {
	switch m.Type {
	case model.Canary:
		stableWeight, canaryWeight := 80, 20
		// 可扩展：后续从 m.CanaryConfig 获取权重
		return []*v1alpha3.HTTPRouteDestination{
			{
				Destination: &v1alpha3.Destination{
					Host:   m.ApplicationName,
					Subset: "stable",
				},
				Weight: int32(stableWeight),
			},
			{
				Destination: &v1alpha3.Destination{
					Host:   m.ApplicationName,
					Subset: "canary",
				},
				Weight: int32(canaryWeight),
			},
		}

	case model.BlueGreen:
		activeSvc := m.ApplicationName + "-active"
		return []*v1alpha3.HTTPRouteDestination{
			{
				Destination: &v1alpha3.Destination{
					Host: activeSvc,
				},
				Weight: 100,
			},
		}

	default: // normal / rolling update
		return []*v1alpha3.HTTPRouteDestination{
			{
				Destination: &v1alpha3.Destination{
					Host: m.ApplicationName,
				},
				Weight: 100,
			},
		}
	}
}
