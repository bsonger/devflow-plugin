package render

import (
	"github.com/bsonger/devflow-plugin/model"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"
)

// Deployment 生成 Deployment YAML，用于普通部署（RollingUpdate）
func Deployment(c *model.Release) (string, error) {

	deploy := &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apps/v1",
			Kind:       "Deployment",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      c.App.Name,
			Namespace: c.App.Namespace,
			Labels: map[string]string{
				"app": c.App.Name,
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: c.Replica,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": c.App.Name,
				},
			},
			Template: buildPodTemplate(c),
		},
	}

	yml, err := yaml.Marshal(deploy)
	if err != nil {
		return "", err
	}
	return string(yml), nil
}
