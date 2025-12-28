package render

import (
	"github.com/bsonger/devflow-plugin/model"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
)

func buildPodTemplate(c *model.Release) corev1.PodTemplateSpec {
	volumes, mounts := buildVolumes(c.Configs, c.Env)

	return corev1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Labels: map[string]string{
				"app": c.App.Name,
			},
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:            c.App.Name,
					Image:           c.Image.Repository + ":" + c.Image.Tag,
					ImagePullPolicy: corev1.PullAlways,
					Env: []corev1.EnvVar{
						{
							Name:  "env",
							Value: string(c.Internet), // 或 prod/test
						},
					},
					VolumeMounts: mounts,
				},
			},
			Volumes:                       volumes,
			RestartPolicy:                 corev1.RestartPolicyAlways,
			DNSPolicy:                     corev1.DNSClusterFirst,
			SchedulerName:                 corev1.DefaultSchedulerName,
			TerminationGracePeriodSeconds: ptr.To(int64(30)),
		},
	}
}

func buildVolumes(cfgs []model.Config, env string) ([]corev1.Volume, []corev1.VolumeMount) {
	var volumes []corev1.Volume
	var mounts []corev1.VolumeMount

	for _, cfg := range cfgs {
		volumes = append(volumes, corev1.Volume{
			Name: cfg.Name,
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: cfg.Name,
					},
					DefaultMode: ptr.To(int32(420)),
				},
			},
		})

		mounts = append(mounts, corev1.VolumeMount{
			Name:      cfg.Name,
			MountPath: cfg.MountPath,
			ReadOnly:  true,
		})
	}

	return volumes, mounts
}
