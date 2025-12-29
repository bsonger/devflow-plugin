package render

import (
	"github.com/bsonger/devflow-plugin/model"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"os"
	"path/filepath"
	"sigs.k8s.io/yaml"
)

func ConfigMap(c *model.Release, baseDir string) (string, error) {
	// 1️⃣ Argo CD clone 后的 repo 根目录

	data := make(map[string]string)

	// 2️⃣ 遍历 release.config
	for _, cfg := range c.Configs {
		dir, ok := cfg.FilesPath[c.Env]
		if !ok {
			continue
		}

		absDir := filepath.Join(baseDir, dir)

		err := filepath.Walk(absDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				return nil
			}

			content, err := os.ReadFile(path)
			if err != nil {
				return err
			}

			// key = 文件名
			data[info.Name()] = string(content)
			return nil
		})

		if err != nil {
			return "", err
		}
	}

	if len(data) == 0 {
		return "", nil
	}

	// 3️⃣ 构造 ConfigMap
	cm := &corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "ConfigMap",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      c.App.Name,
			Namespace: c.App.Namespace,
			Labels: map[string]string{
				"app": c.App.Name,
			},
		},
		Data: data,
	}

	yml, err := yaml.Marshal(cm)
	if err != nil {
		return "", err
	}

	return string(yml), nil
}
