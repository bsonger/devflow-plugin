package main

import (
	"encoding/json"
	"fmt"
	"github.com/bsonger/devflow-common/model"
	"github.com/bsonger/devflow-plugin/render"
	"log"
	"net/http"
	"strings"
)

func main() {
	//manifestID := os.Getenv("PARAM_MANIFEST_ID")
	//env := os.Getenv("PARAM_ENV")
	//devflowAddress := os.Getenv("PARAM_DEVFLOW_ADDRESS") // 修改名称
	//consulAddress := os.Getenv("PARAM_CONSUL_ADDRESS")   // 新增 Consul 地址
	manifestID := "697452200cb9977fda9c31ec"
	env := "prod"
	devflowAddress := "https://devflow.bei.com"
	consulAddress := "https://consul.bei.com"

	log.Printf("manifestID=%s, env=%s, devflowAddress=%s, consulAddress=%s\n",
		manifestID, env, devflowAddress, consulAddress)

	if manifestID == "" {
		log.Fatalf("manifest-id is required")
		return
	}

	manifest, err := FetchRelease(devflowAddress, manifestID)
	if err != nil {
		log.Fatal(err)
		return
	}
	//manifest.Type = model.BlueGreen
	//manifest.Internet = model.Internal

	// -----------------------------
	// 1️⃣ ConfigMap
	// -----------------------------
	ConfigYAML, err := render.ConfigMap(manifest, env, consulAddress)
	if err != nil {
		log.Fatalf("RenderConfigmap failed: %v", err)
	}
	fmt.Println("---")
	fmt.Println(ConfigYAML)

	// -----------------------------
	// 2️⃣ Service & Rollout/Deployment
	// -----------------------------
	svcYAML, err := render.Service(manifest)
	if err != nil {
		log.Fatalf("Service failed: %v", err)
	}
	fmt.Println("---")
	fmt.Println(svcYAML)

	switch manifest.Type {
	case model.Normal:
		deployYAML, err := render.Deployment(manifest, env)
		if err != nil {
			log.Fatalf("RenderDeploy failed: %v", err)
		}
		fmt.Println("---")
		fmt.Println(deployYAML)

	case model.Canary, model.BlueGreen:
		rolloutYAML, err := render.Rollout(manifest, env)
		if err != nil {
			log.Fatalf("Rollout failed: %v", err)
		}
		fmt.Println("---")
		fmt.Println(rolloutYAML)
	}

	// -----------------------------
	// 3️⃣ VirtualService & DestinationRule (仅 Canary / BlueGreen)
	// -----------------------------
	vsYAML, err := render.VirtualService(manifest, env)
	if err != nil {
		log.Fatalf("VirtualService failed: %v", err)
	}
	vsYAML = strings.ReplaceAll(vsYAML, "status: {}\n", "")
	if vsYAML != "" {
		fmt.Println("---")
		fmt.Println(vsYAML)
	}

	drYAML, err := render.DestinationRule(manifest, env)
	if err != nil {
		log.Fatalf("DestinationRule failed: %v", err)
	}
	drYAML = strings.ReplaceAll(drYAML, "status: {}\n", "")
	if drYAML != "" {
		fmt.Println("---")
		fmt.Println(drYAML)
	}
}

// FetchRelease 调用 DevFlow API 获取 Manifest
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
