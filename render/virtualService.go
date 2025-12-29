package render

import (
	"github.com/bsonger/devflow-plugin/model"
	v1alpha3 "istio.io/api/networking/v1alpha3"
	networkingv1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"
)

// VirtualService 根据 ReleaseContent 生成 VirtualService YAML
func VirtualService(c *model.Release) (string, error) {
	// 内部流量不需要 VirtualService
	if c.Internet == model.Internal {
		return "", nil
	}

	vs := &networkingv1beta1.VirtualService{
		TypeMeta: metav1.TypeMeta{
			Kind:       "VirtualService",
			APIVersion: "networking.istio.io/v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      c.App.Name,
			Namespace: c.App.Namespace,
			Labels: map[string]string{
				"app": c.App.Name,
			},
		},
		Spec: v1alpha3.VirtualService{
			Hosts: []string{c.App.Name},
			Http: []*v1alpha3.HTTPRoute{
				{
					Route: buildHTTPRouteDestinations(c),
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
func buildHTTPRouteDestinations(c *model.Release) []*v1alpha3.HTTPRouteDestination {
	switch c.Type {
	case model.Canary:
		// Canary 按百分比流量
		return []*v1alpha3.HTTPRouteDestination{
			{
				Destination: &v1alpha3.Destination{
					Host:   c.App.Name,
					Subset: "stable",
				},
				Weight: 80,
			},
			{
				Destination: &v1alpha3.Destination{
					Host:   c.App.Name,
					Subset: "canary",
				},
				Weight: 20,
			},
		}
	case model.BlueGreen:
		// BlueGreen 刚开始全部流量打到 active
		return []*v1alpha3.HTTPRouteDestination{
			{
				Destination: &v1alpha3.Destination{
					Host: c.App.Name + "-active",
				},
				Weight: 100,
			},
		}
	default:
		// 普通部署直接指向服务
		return []*v1alpha3.HTTPRouteDestination{
			{
				Destination: &v1alpha3.Destination{
					Host: c.App.Name,
				},
				Weight: 100,
			},
		}
	}
}
