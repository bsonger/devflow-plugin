package render

import (
	"fmt"
	"github.com/bsonger/devflow-common/model"
	v1alpha3 "istio.io/api/networking/v1alpha3"
	networkingv1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"
)

func VirtualService(m *model.Manifest, env string) (string, error) {
	if m.Type == model.Normal && m.Internet == model.Internal {
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
				"env": env,
			},
		},
	}

	spec := v1alpha3.VirtualService{
		Http: []*v1alpha3.HTTPRoute{
			{
				Name:  m.ApplicationName,
				Route: buildHTTPRouteDestinations(m),
			},
		},
	}

	if m.Internet == model.External {
		spec.Hosts = []string{
			fmt.Sprintf("%s.bei.com", m.ApplicationName),
			m.ApplicationName,
		}
		spec.Gateways = []string{
			"istio-system/devflow-gateway",
		}
	} else {
		spec.Hosts = []string{
			m.ApplicationName,
		}
	}

	vs.Spec = spec

	yml, err := yaml.Marshal(vs)
	if err != nil {
		return "", err
	}
	return string(yml), nil
}

func buildHTTPRouteDestinations(m *model.Manifest) []*v1alpha3.HTTPRouteDestination {
	switch m.Type {

	case model.Canary:
		// Canary 初始全部流量给 stable
		return []*v1alpha3.HTTPRouteDestination{
			{
				Destination: &v1alpha3.Destination{
					Host:   m.ApplicationName,
					Subset: "stable",
				},
				Weight: 100,
			},
			{
				Destination: &v1alpha3.Destination{
					Host:   m.ApplicationName,
					Subset: "canary",
				},
				Weight: 0,
			},
		}

	case model.BlueGreen:
		// BlueGreen 使用同一个 Service，通过 subset 区分 blue/green
		return []*v1alpha3.HTTPRouteDestination{
			{
				Destination: &v1alpha3.Destination{
					Host:   m.ApplicationName,
					Subset: "blue",
				},
				Weight: 100, // 初始全部流量给 active 蓝色
			},
			{
				Destination: &v1alpha3.Destination{
					Host:   m.ApplicationName,
					Subset: "green",
				},
				Weight: 0,
			},
		}

	default: // normal / rolling
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
