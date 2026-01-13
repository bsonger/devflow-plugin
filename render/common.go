package render

import (
	"fmt"
	"github.com/bsonger/devflow-common/model"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
)

const (
	Repository = "registry.cn-hangzhou.aliyuncs.com/devflow/"
)

func buildPodTemplate(m *model.Manifest, env string) corev1.PodTemplateSpec {
	envVars := buildEnvVars(m.Envs, env, m.ApplicationName)
	volumes, mounts := buildVolumes(m.ApplicationName, m.ConfigMaps, env)

	metricsPort := findMetricsPort(m.Service.Ports)
	annotations := map[string]string{}
	if metricsPort != nil {
		annotations["prometheus.io/scrape"] = "true"
		annotations["prometheus.io/path"] = "/metrics"
		annotations["prometheus.io/port"] = fmt.Sprintf("%d", metricsPort.Port)
		annotations["prometheus.io/scheme"] = "http"
		annotations["prometheus.io/job"] = m.ApplicationName
	}

	return corev1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Labels: map[string]string{
				"app":  m.ApplicationName,
				"type": string(m.Type),
				"env":  env,
			},
			Annotations: annotations,
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:            m.ApplicationName,
					Image:           Repository + m.ApplicationName + ":" + m.Name,
					ImagePullPolicy: corev1.PullIfNotPresent,
					Env:             envVars,
					VolumeMounts:    mounts,
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

func buildVolumes(appName string, cfgs []*model.ConfigMap, env string) ([]corev1.Volume, []corev1.VolumeMount) {
	var volumes []corev1.Volume
	var mounts []corev1.VolumeMount

	for _, cfg := range cfgs {
		// 可以根据 env 选择不同 config
		cfgName := fmt.Sprintf("%s-%s-%s", appName, cfg.Name, env)
		volumes = append(volumes, corev1.Volume{
			Name: cfgName,
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: cfgName,
					},
					DefaultMode: ptr.To(int32(420)),
				},
			},
		})

		mounts = append(mounts, corev1.VolumeMount{
			Name:      cfgName,
			MountPath: cfg.MountPath,
			ReadOnly:  true,
		})
	}

	return volumes, mounts
}

func buildEnvVars(envMap map[string][]model.EnvVar, env string, serviceName string) []corev1.EnvVar {
	var envVars []corev1.EnvVar

	// ========== 1️⃣ 固定业务环境变量 ==========
	envVars = append(envVars,
		corev1.EnvVar{
			Name:  "ENV",
			Value: env,
		},
		corev1.EnvVar{
			Name:  "SERVICE_NAME",
			Value: serviceName,
		},
	)

	// ========== 2️⃣ Kubernetes Downward API ==========
	envVars = append(envVars,
		corev1.EnvVar{
			Name: "POD_NAME",
			ValueFrom: &corev1.EnvVarSource{
				FieldRef: &corev1.ObjectFieldSelector{
					FieldPath: "metadata.name",
				},
			},
		},
		corev1.EnvVar{
			Name: "POD_NAMESPACE",
			ValueFrom: &corev1.EnvVarSource{
				FieldRef: &corev1.ObjectFieldSelector{
					FieldPath: "metadata.namespace",
				},
			},
		},
		corev1.EnvVar{
			Name: "POD_UID",
			ValueFrom: &corev1.EnvVarSource{
				FieldRef: &corev1.ObjectFieldSelector{
					FieldPath: "metadata.uid",
				},
			},
		},
		corev1.EnvVar{
			Name: "NODE_NAME",
			ValueFrom: &corev1.EnvVarSource{
				FieldRef: &corev1.ObjectFieldSelector{
					FieldPath: "spec.nodeName",
				},
			},
		},
		corev1.EnvVar{
			Name: "POD_IP",
			ValueFrom: &corev1.EnvVarSource{
				FieldRef: &corev1.ObjectFieldSelector{
					FieldPath: "status.podIP",
				},
			},
		},
	)

	// ========== 3️⃣ 按环境注入自定义变量 ==========
	if envList, ok := envMap[env]; ok {
		for _, e := range envList {
			envVars = append(envVars, corev1.EnvVar{
				Name:  e.Name,
				Value: e.Value,
			})
		}
	}

	return envVars
}
