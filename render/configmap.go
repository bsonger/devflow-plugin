package render

import (
	"encoding/base64"
	"fmt"
	"github.com/bsonger/devflow-common/model"
	"github.com/hashicorp/consul/api"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"path/filepath"
	"sigs.k8s.io/yaml"
	"strings"
)

func ConfigMap(m *model.Manifest, env string) (string, error) {
	cfg := api.DefaultConfig()
	cfg.Address = "https://consul.bei.com:30000"

	client, err := api.NewClient(cfg)
	if err != nil {
		return "", err
	}

	var allCMs []*corev1.ConfigMap

	for _, cmCfg := range m.ConfigMaps {

		// 🔥 正确的 Consul KV 路径（不是 UI）
		kvPath := fmt.Sprintf(
			"app/%s/%s/%s",
			m.ApplicationName,
			cmCfg.Name,
			env,
		)

		data, err := fetchConsulKVFromPath(client, kvPath)
		if err != nil {
			return "", err
		}

		if len(data) == 0 {
			continue
		}

		cm := &corev1.ConfigMap{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "v1",
				Kind:       "ConfigMap",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: fmt.Sprintf("%s-%s-%s", m.ApplicationName, cmCfg.Name, env),
				Labels: map[string]string{
					"app": m.ApplicationName,
					"env": env,
				},
			},
			Data: data,
		}

		allCMs = append(allCMs, cm)
	}

	if len(allCMs) == 0 {
		return "", nil
	}

	var out strings.Builder
	for _, cm := range allCMs {
		yml, err := yaml.Marshal(cm)
		if err != nil {
			return "", err
		}
		out.WriteString(string(yml))
		out.WriteString("\n---\n")
	}

	return out.String(), nil
}

// fetchConsulKVFromPath 使用 Consul 官方 client 获取 KV
func fetchConsulKVFromPath(client *api.Client, prefix string) (map[string]string, error) {
	data := make(map[string]string)

	kv := client.KV()

	pairs, _, err := kv.List(prefix, nil)
	if err != nil {
		return nil, err
	}

	for _, pair := range pairs {
		if len(pair.Value) == 0 {
			continue
		}

		// key 只保留文件名
		keyName := filepath.Base(pair.Key)

		decoded, err := base64.StdEncoding.DecodeString(string(pair.Value))
		if err != nil {
			// 有些 KV 可能本来就不是 base64（可选兜底）
			data[keyName] = string(pair.Value)
			continue
		}

		data[keyName] = string(decoded)
	}

	return data, nil
}
