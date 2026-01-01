package render

import (
	"github.com/bsonger/devflow-common/model"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/yaml"

	rolloutv1alpha1 "github.com/argoproj/argo-rollouts/pkg/apis/rollouts/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Rollout 根据 Manifest 和环境生成 Argo Rollout YAML
func Rollout(m *model.Manifest, env string) (string, error) {
	rollout := &rolloutv1alpha1.Rollout{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "argoproj.io/v1alpha1",
			Kind:       "Rollout",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: m.ApplicationName + "-" + env,
			Labels: map[string]string{
				"app":  m.ApplicationName,
				"env":  env,
				"type": string(m.Type),
			},
		},
		Spec: rolloutv1alpha1.RolloutSpec{
			Replicas:             m.Replica,
			RevisionHistoryLimit: ptr.To(int32(10)),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": m.ApplicationName,
					"env": env,
				},
			},
			Template: buildPodTemplate(m, env),
		},
	}

	// 根据 ReleaseType 选择 Rollout Strategy
	switch m.Type {
	case model.Canary:
		rollout.Spec.Strategy = buildCanaryStrategy(m, env)
	case model.BlueGreen:
		rollout.Spec.Strategy = buildBlueGreenStrategy(m, env)
	default:
		// normal: 不设置 Strategy，默认 RollingUpdate
	}

	out, err := yaml.Marshal(rollout)
	if err != nil {
		return "", err
	}
	return string(out), nil
}

func buildCanaryStrategy(m *model.Manifest, env string) rolloutv1alpha1.RolloutStrategy {
	// 固定 Canary 流程示例
	steps := []rolloutv1alpha1.CanaryStep{
		{SetWeight: ptr.To(int32(20))}, // 先发 20%
		{
			Pause: &rolloutv1alpha1.RolloutPause{
				Duration: ptr.To(intstr.FromInt(30)), // pause 30s
			},
		},
		{SetWeight: ptr.To(int32(50))}, // 升级到 50%
		{
			Pause: &rolloutv1alpha1.RolloutPause{
				Duration: ptr.To(intstr.FromInt(60)), // pause 60s
			},
		},
		{SetWeight: ptr.To(int32(100))}, // 完全发布
	}

	return rolloutv1alpha1.RolloutStrategy{
		Canary: &rolloutv1alpha1.CanaryStrategy{
			Steps: steps,
			TrafficRouting: &rolloutv1alpha1.RolloutTrafficRouting{
				Istio: &rolloutv1alpha1.IstioTrafficRouting{
					VirtualService: &rolloutv1alpha1.IstioVirtualService{
						Name:   m.ApplicationName,
						Routes: []string{"primary"},
					},
					DestinationRule: &rolloutv1alpha1.IstioDestinationRule{
						Name:             m.ApplicationName,
						CanarySubsetName: "canary",
						StableSubsetName: "stable",
					},
				},
			},
		},
	}
}

func buildBlueGreenStrategy(m *model.Manifest, env string) rolloutv1alpha1.RolloutStrategy {
	activeSvc := m.ApplicationName + "-active"
	previewSvc := m.ApplicationName + "-preview"

	autoPromote := false
	if m.Internet == model.Internal {
		autoPromote = true
	}

	return rolloutv1alpha1.RolloutStrategy{
		BlueGreen: &rolloutv1alpha1.BlueGreenStrategy{
			ActiveService:         activeSvc,
			PreviewService:        previewSvc,
			AutoPromotionEnabled:  ptr.To(autoPromote),
			ScaleDownDelaySeconds: ptr.To(int32(30)),
		},
	}
}
