package render_test

import (
	"github.com/bsonger/devflow-plugin/render"
	"k8s.io/utils/ptr"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"

	"github.com/bsonger/devflow-plugin/model"
)

func TestRenderRollout_Canary(t *testing.T) {
	release := &model.Release{
		Type:    model.Canary,
		Replica: ptr.To(int32(3)),
		App: model.App{
			Name:      "devflow",
			Namespace: "app",
		},
		Image: model.Image{
			Repository: "repo/devflow",
			Tag:        "v1",
		},
		Configs: []model.Config{
			model.Config{
				Name:      "devflow",
				MountPath: "/etc/config/devflow",
				FilesPath: map[string]string{
					"config.yaml": "config",
				},
			},
		},
	}

	yml, err := render.RenderRollout(release)
	assert.NoError(t, err)
	assert.NotEmpty(t, yml)
	t.Log(yml)

	var obj map[string]any
	err = yaml.Unmarshal([]byte(yml), &obj)
	assert.NoError(t, err)

	spec := obj["spec"].(map[string]any)
	strategy := spec["strategy"].(map[string]any)

	assert.Contains(t, strategy, "canary")

}
