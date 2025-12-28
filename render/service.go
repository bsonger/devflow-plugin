package render

import (
	"gopkg.in/yaml.v3"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	"github.com/bsonger/devflow-plugin/model"
)

// RenderService 根据 Release 渲染 Service YAML
func RenderService(c *model.Release) (string, error) {
	if len(c.Service) == 0 {
		return "", nil // 没有 service 配置，直接返回空
	}

	var ports []corev1.ServicePort

	for _, p := range c.Service {
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
			Name:      c.App.Name,
			Namespace: c.App.Namespace,
			Labels: map[string]string{
				"app": c.App.Name,
			},
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				"app": c.App.Name,
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

func RenderServiceBlueGreen(r *model.Release, role string) (string, error) {

	name := r.App.Name + "-" + role

	var ports []corev1.ServicePort
	for _, p := range r.Service {
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
			Name:      name,
			Namespace: r.App.Namespace,
			Labels: map[string]string{
				"app":  r.App.Name,
				"role": role,
			},
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				"app":  r.App.Name,
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
