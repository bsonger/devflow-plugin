package render

import (
	"fmt"
	"github.com/bsonger/devflow-common/model"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/yaml"
)

func Service(m *model.Manifest) (string, error) {
	if len(m.Service.Ports) == 0 {
		return "", nil
	}

	out := ""

	metricsPort := findMetricsPort(m.Service.Ports)

	// 普通 / Canary / Rolling Update
	yml, err := buildServiceYAML(m.ApplicationName, m.ApplicationName, nil, m.Service.Ports, metricsPort)
	if err != nil {
		return "", err
	}
	out += yml + "\n---\n"

	// BlueGreen：preview Service（不加 metrics）
	if m.Type == model.BlueGreen {
		previewName := m.ApplicationName + "-preview"
		yml, err := buildServiceYAML(previewName, m.ApplicationName, nil, m.Service.Ports, nil)
		if err != nil {
			return "", err
		}
		out += yml + "\n---\n"
	}

	return out, nil
}

func buildServiceYAML(name string, selectorApp string, extraSelector map[string]string, ports []model.Port, metricsPort *model.Port) (string, error) {

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

	labels := map[string]string{}
	for k, v := range selector {
		labels[k] = v
	}

	annotations := map[string]string{}

	// ✅ metrics annotation 只在 Service 上加
	if metricsPort != nil {
		labels["monitoring"] = "enabled"
	}

	svc := &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Labels:      labels,
			Annotations: annotations,
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
func findMetricsPort(ports []model.Port) *model.Port {
	for _, p := range ports {
		if p.Name == "metrics" {
			return &p
		}
	}
	return nil
}
