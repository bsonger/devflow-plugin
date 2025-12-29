package render

import (
	"github.com/bsonger/devflow-plugin/model"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/yaml"

	rolloutv1alpha1 "github.com/argoproj/argo-rollouts/pkg/apis/rollouts/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Rollout(c *model.Release) (string, error) {
	rollout := &rolloutv1alpha1.Rollout{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "argoproj.io/v1alpha1",
			Kind:       "Rollout",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      c.App.Name,
			Namespace: c.App.Namespace,
			Labels: map[string]string{
				"app": c.App.Name,
			},
		},
		Spec: rolloutv1alpha1.RolloutSpec{
			Replicas:             c.Replica,
			RevisionHistoryLimit: ptr.To(int32(10)),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": c.App.Name,
				},
			},
			Template: buildPodTemplate(c),
		},
	}

	// 🔥 关键：根据发布类型选择 Strategy
	switch c.Type {
	case model.Canary:
		rollout.Spec.Strategy = buildCanaryStrategy(c)
	case model.BlueGreen:
		rollout.Spec.Strategy = buildBlueGreenStrategy(c) // 你下一步要写的
	default:
		// normal 可以直接 Deployment 或 RollingUpdate
	}

	out, err := yaml.Marshal(rollout)
	if err != nil {
		return "", err
	}
	return string(out), nil
}

func buildCanaryStrategy(c *model.Release) rolloutv1alpha1.RolloutStrategy {
	return rolloutv1alpha1.RolloutStrategy{
		Canary: &rolloutv1alpha1.CanaryStrategy{
			Steps: []rolloutv1alpha1.CanaryStep{
				{SetWeight: ptr.To(int32(20))},
				{
					Pause: &rolloutv1alpha1.RolloutPause{
						Duration: ptr.To(intstr.FromInt(30)),
					},
				},
				{SetWeight: ptr.To(int32(50))},
				{
					Pause: &rolloutv1alpha1.RolloutPause{
						Duration: ptr.To(intstr.FromInt(60)),
					},
				},
				{SetWeight: ptr.To(int32(100))},
			},
			TrafficRouting: &rolloutv1alpha1.RolloutTrafficRouting{
				Istio: &rolloutv1alpha1.IstioTrafficRouting{
					VirtualService: &rolloutv1alpha1.IstioVirtualService{
						Name:   c.App.Name,
						Routes: []string{"primary"},
					},
					DestinationRule: &rolloutv1alpha1.IstioDestinationRule{
						Name:             c.App.Name,
						CanarySubsetName: "canary",
						StableSubsetName: "stable",
					},
				},
			},
		},
	}
}

func buildBlueGreenStrategy(c *model.Release) rolloutv1alpha1.RolloutStrategy {
	activeSvc := c.App.Name + "-active"
	previewSvc := c.App.Name + "-preview"

	autoPromote := false
	if c.Internet == model.Internal {
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
