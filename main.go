package main

import (
	"encoding/json"
	"fmt"
	"github.com/bsonger/devflow-common/model"
	"github.com/bsonger/devflow-plugin/render"
	"log"
	"net/http"
	"os"
)

func main() {

	manifestID := os.Getenv("PARAM_MANIFEST_ID")
	env := os.Getenv("PARAM_ENV")
	devflowAddress := os.Getenv("PARAM_DEVFLOW_ADDRESS") // 修改名称
	consulAddress := os.Getenv("PARAM_CONSUL_ADDRESS")   // 新增 Consul 地址

	log.Printf("manifestID=%s, env=%s, devflowAPI=%s\n", manifestID, env, devflowAddress)

	if manifestID == "" {
		log.Fatalf("manifest-id is required")
		return
	}

	manifest, err := FetchRelease(devflowAddress, manifestID)
	if err != nil {
		log.Fatal(err)
		return
	}

	ConfigYAML, err := render.ConfigMap(manifest, env, consulAddress)
	if err != nil {
		log.Fatalf("RenderConfigmap failed: %v", err)
	}
	fmt.Println("---")
	fmt.Println(ConfigYAML)

	switch manifest.Type {
	case model.Normal:
		// 普通部署和灰度都生成一个 Service
		svcYAML, err := render.Service(manifest)
		if err != nil {
			log.Fatalf("Service failed: %v", err)
		}
		fmt.Println("---")
		fmt.Println(svcYAML)

		deployYAML, err := render.Deployment(manifest, env)
		if err != nil {
			log.Fatalf("RenderDeploy failed: %v", err)
		}
		fmt.Println("---")
		fmt.Println(deployYAML)

	case model.Canary:
		svcYAML, err := render.Service(manifest)
		if err != nil {
			log.Fatalf("Service failed: %v", err)
		}
		fmt.Println("---")
		fmt.Println(svcYAML)

		rolloutYAML, err := render.Rollout(manifest, env)
		if err != nil {
			log.Fatalf("Rollout failed: %v", err)
		}
		fmt.Println("---")
		fmt.Println(rolloutYAML)

	case model.BlueGreen:
		// 蓝绿发布生成两个 Service：active 和 preview
		activeSvcYAML, err := render.BlueGreen(manifest, "active")
		if err != nil {
			log.Fatalf("Service active failed: %v", err)
		}
		previewSvcYAML, err := render.BlueGreen(manifest, "preview")
		if err != nil {
			log.Fatalf("Service preview failed: %v", err)
		}
		fmt.Println("---")
		fmt.Println(activeSvcYAML)
		fmt.Println("---")
		fmt.Println(previewSvcYAML)

		rolloutYAML, err := render.Rollout(manifest, env)
		if err != nil {
			log.Fatalf("Rollout failed: %v", err)
		}
		fmt.Println("---")
		fmt.Println(rolloutYAML)
	}

	// -----------------------------
	// 3️⃣ 外网生成 VirtualService（仅 Canary / BlueGreen）
	// -----------------------------
	if manifest.Internet == model.External {
		vsYAML, err := render.VirtualService(manifest)
		if err != nil {
			log.Fatalf("VirtualService failed: %v", err)
		}
		if vsYAML != "" {
			fmt.Println("---")
			fmt.Println(vsYAML)
		}
	}
}

func FetchRelease(api, manifestID string) (*model.Manifest, error) {
	url := fmt.Sprintf("%s/api/v1/manifests/%s", api, manifestID)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("devflow api error: %s", resp.Status)
	}

	var manifest model.Manifest
	if err := json.NewDecoder(resp.Body).Decode(&manifest); err != nil {
		return nil, err
	}

	return &manifest, nil
}
