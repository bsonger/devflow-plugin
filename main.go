package main

import (
	"flag"
	"fmt"
	"github.com/bsonger/devflow-plugin/model"
	"github.com/bsonger/devflow-plugin/render"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"k8s.io/utils/ptr"
	"log"
	"path/filepath"
	"strconv"
)

func main() {

	releaseType := flag.String("type", "normal", "release type: normal/canary/bluegreen")
	replicaStr := flag.String("replica", "1", "replica count")
	internet := flag.String("internet", "internal", "internet type: internal/external")
	env := flag.String("env", "devflow-plugin", "env type: devflow-plugin")
	repoURL := flag.String("repoUrl", "", "release.yaml github raw URL")
	path := flag.String("path", ".", "subpath in repo where release.yaml exists")

	flag.Parse()

	replica, err := strconv.Atoi(*replicaStr)
	if err != nil {
		log.Fatalf("invalid replica: %v", err)
	}

	// git clone / pull 仓库
	baseDir := filepath.Join("/tmp", render.RepoDirName(*repoURL))
	if err := render.GitCloneOrPull(*repoURL, "main", baseDir); err != nil {
		log.Fatalf("git clone/pull failed: %v", err)
	}

	// 拼接 release.yaml 的完整路径
	appPath := filepath.Join(baseDir, *path)
	releaseFile := filepath.Join(appPath, "release.yaml")

	// 读取 release.yaml 文件
	data, err := ioutil.ReadFile(releaseFile)
	if err != nil {
		log.Fatalf("failed to read release.yaml: %v", err)
	}

	var release model.Release
	if err := yaml.Unmarshal(data, &release); err != nil {
		log.Fatalf("failed to unmarshal release.yaml: %v", err)
	}

	// 合并 CLI 参数覆盖 YAML 中的字段
	release.Type = model.ReleaseType(*releaseType)
	release.Replica = ptr.To[int32](int32(replica))
	release.Internet = model.Internet(*internet)
	release.Env = *env
	ConfigYAML, err := render.ConfigMap(&release, appPath)
	if err != nil {
		log.Fatalf("RenderConfigmap failed: %v", err)
	}
	fmt.Println("---")
	fmt.Println(ConfigYAML)

	switch release.Type {
	case model.Normal:
		// 普通部署和灰度都生成一个 Service
		svcYAML, err := render.Service(&release)
		if err != nil {
			log.Fatalf("Service failed: %v", err)
		}
		fmt.Println("---")
		fmt.Println(svcYAML)

		deployYAML, err := render.Deployment(&release)
		if err != nil {
			log.Fatalf("RenderDeploy failed: %v", err)
		}
		fmt.Println("---")
		fmt.Println(deployYAML)

	case model.Canary:
		svcYAML, err := render.Service(&release)
		if err != nil {
			log.Fatalf("Service failed: %v", err)
		}
		fmt.Println("---")
		fmt.Println(svcYAML)

		rolloutYAML, err := render.Rollout(&release)
		if err != nil {
			log.Fatalf("Rollout failed: %v", err)
		}
		fmt.Println("---")
		fmt.Println(rolloutYAML)

	case model.BlueGreen:
		// 蓝绿发布生成两个 Service：active 和 preview
		activeSvcYAML, err := render.BlueGreen(&release, "active")
		if err != nil {
			log.Fatalf("Service active failed: %v", err)
		}
		previewSvcYAML, err := render.BlueGreen(&release, "preview")
		if err != nil {
			log.Fatalf("Service preview failed: %v", err)
		}
		fmt.Println("---")
		fmt.Println(activeSvcYAML)
		fmt.Println("---")
		fmt.Println(previewSvcYAML)

		rolloutYAML, err := render.Rollout(&release)
		if err != nil {
			log.Fatalf("Rollout failed: %v", err)
		}
		fmt.Println("---")
		fmt.Println(rolloutYAML)
	}

	// -----------------------------
	// 3️⃣ 外网生成 VirtualService（仅 Canary / BlueGreen）
	// -----------------------------
	if release.Internet == model.External {
		vsYAML, err := render.VirtualService(&release)
		if err != nil {
			log.Fatalf("VirtualService failed: %v", err)
		}
		if vsYAML != "" {
			fmt.Println("---")
			fmt.Println(vsYAML)
		}
	}
}
