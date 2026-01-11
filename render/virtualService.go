package render

import (
	"fmt"
	"github.com/bsonger/devflow-common/model"
	v1alpha3 "istio.io/api/networking/v1alpha3"
	networkingv1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"
)

// VirtualService 根据 manifest 生成 VS，只在有 HTTP port 时生成
func VirtualService(m *model.Manifest, env string) (string, error) {
	// 内部服务且非 Canary，不生成 VS
	if m.Type != model.Canary && m.Internet == model.Internal {
		return "", nil
	}

	// 找 HTTP port
	var httpPort int32
	for _, p := range m.Service.Ports {
		if p.Name == "http" {
			httpPort = int32(p.Port)
			break
		}
	}

	if httpPort == 0 {
		// 没有 HTTP port，不生成 VS
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
				Route: buildHTTPRouteDestinations(m, httpPort),
			},
		},
	}

	// 外部服务设置 Host 和 Gateway
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

// buildHTTPRouteDestinations 带 port number
func buildHTTPRouteDestinations(m *model.Manifest, port int32) []*v1alpha3.HTTPRouteDestination {
	switch m.Type {
	case model.Canary:
		return []*v1alpha3.HTTPRouteDestination{
			{
				Destination: &v1alpha3.Destination{
					Host:   m.ApplicationName,
					Subset: "stable",
					Port: &v1alpha3.PortSelector{
						Number: uint32(port),
					},
				},
				Weight: 100,
			},
			{
				Destination: &v1alpha3.Destination{
					Host:   m.ApplicationName,
					Subset: "canary",
					Port: &v1alpha3.PortSelector{
						Number: uint32(port),
					},
				},
				Weight: 0,
			},
		}
	default: // normal / rolling
		return []*v1alpha3.HTTPRouteDestination{
			{
				Destination: &v1alpha3.Destination{
					Host: m.ApplicationName,
					Port: &v1alpha3.PortSelector{
						Number: uint32(port),
					},
				},
				Weight: 100,
			},
		}
	}
}
