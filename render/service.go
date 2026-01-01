package render

import (
	"github.com/bsonger/devflow-common/model"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/yaml"
)

// Service 根据 Release 渲染 Service YAML
func Service(m *model.Manifest) (string, error) {
	if len(m.Service.Ports) == 0 {
		return "", nil // 没有 service 配置，直接返回空
	}

	var ports []corev1.ServicePort

	for _, p := range m.Service.Ports {
		ports = append(ports, corev1.ServicePort{
			Name:       p.Name,
			Port:       int32(p.Port),
			TargetPort: intstrFromInt(p.TargetPort),
			Protocol:   corev1.ProtocolTCP,
		})
	}

	svc := &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: m.ApplicationName,
			Labels: map[string]string{
				"app": m.ApplicationName,
			},
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				"app": m.ApplicationName,
			},
			Ports: ports,
			Type:  corev1.ServiceTypeClusterIP,
		},
	}

	yml, err := yaml.Marshal(svc)
	if err != nil {
		return "", err
	}
	return string(yml), nil
}

// intstrFromInt 辅助函数
func intstrFromInt(i int) intstr.IntOrString {
	return intstr.IntOrString{
		Type:   intstr.Int,
		IntVal: int32(i),
	}
}

func BlueGreen(m *model.Manifest, role string) (string, error) {

	name := m.ApplicationName + "-" + role

	var ports []corev1.ServicePort
	for _, p := range m.Service.Ports {
		ports = append(ports, corev1.ServicePort{
			Name:       p.Name,
			Port:       int32(p.Port),
			TargetPort: intstrFromInt(p.TargetPort),
		})
	}

	svc := &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Service",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			//Namespace: m.App.Namespace,
			Labels: map[string]string{
				"app":  m.ApplicationName,
				"role": role,
			},
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				"app":  m.ApplicationName,
				"role": role,
			},
			Ports: ports,
			Type:  corev1.ServiceTypeClusterIP,
		},
	}

	out, err := yaml.Marshal(svc)
	if err != nil {
		return "", err
	}
	return string(out), nil
}
