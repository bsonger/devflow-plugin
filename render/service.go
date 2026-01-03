package render

import (
	"fmt"
	"github.com/bsonger/devflow-common/model"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/yaml"
)

// Service 根据 Release 渲染 Service YAML
// 蓝绿部署会生成 active + preview 两个 Service，并合并成单一 YAML 输出
func Service(m *model.Manifest) (string, error) {
	if len(m.Service.Ports) == 0 {
		return "", nil
	}

	out := ""

	// 普通 / Canary / Rolling Update 使用同一个 Service
	yml, err := buildServiceYAML(m.ApplicationName, m.ApplicationName, nil, m.Service.Ports)
	if err != nil {
		return "", err
	}
	out += yml + "\n---\n"

	// BlueGreen 额外生成 preview Service
	if m.Type == model.BlueGreen {
		previewName := m.ApplicationName + "-preview"
		yml, err := buildServiceYAML(previewName, m.ApplicationName, nil, m.Service.Ports)
		if err != nil {
			return "", err
		}
		out += yml + "\n---\n"
	}

	return out, nil
}

// buildServiceYAML 构建单个 Service 的 YAML
func buildServiceYAML(name string, selectorApp string, extraSelector map[string]string, ports []model.Port) (string, error) {
	var servicePorts []corev1.ServicePort
	for _, p := range ports {
		servicePorts = append(servicePorts, corev1.ServicePort{
			Name:       p.Name,
			Port:       int32(p.Port),
			TargetPort: intstr.IntOrString{Type: intstr.Int, IntVal: int32(p.TargetPort)},
			Protocol:   corev1.ProtocolTCP,
		})
	}

	selector := map[string]string{"app": selectorApp}
	for k, v := range extraSelector {
		selector[k] = v
	}

	svc := &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:   name,
			Labels: selector,
		},
		Spec: corev1.ServiceSpec{
			Selector: selector,
			Ports:    servicePorts,
			Type:     corev1.ServiceTypeClusterIP,
		},
	}

	yml, err := yaml.Marshal(svc)
	if err != nil {
		return "", fmt.Errorf("marshal service %s failed: %w", name, err)
	}
	return string(yml), nil
}
