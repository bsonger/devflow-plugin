package main

import (
	"flag"
	"fmt"
	"k8s.io/utils/ptr"
	"log"
	"os"
	"os/exec"

	"github.com/bsonger/devflow-plugin/model"
	"github.com/bsonger/devflow-plugin/render"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Usage: ./plugin <release.yaml>")
	}

	releaseType := flag.String("type", "normal", "release type: normal/canary/bluegreen")
	replica := flag.Int("replica", 1, "replica count")
	internet := flag.String("internet", "internal", "internet type: internal/external")
	env := flag.String("env", "devflow-plugin", "env type: devflow-plugin")
	flag.Parse()

	repoURL := flag.String("repoURL", "", "release.yaml github raw URL")

	// clone 仓库到本地 path，如果已存在则 pull
	appPath := os.Getenv("ARGOCD_APP_SOURCE_PATH")
	if appPath == "" {
		fmt.Errorf("ARGOCD_APP_SOURCE_PATH not set")
	}
	if _, err := os.Stat(appPath); os.IsNotExist(err) {
		cmd := exec.Command("git", "clone", *repoURL, appPath)
		if err := cmd.Run(); err != nil {
			log.Fatalf("failed to git clone: %v", err)
		}
	} else {
		cmd := exec.Command("git", "-C", appPath, "pull")
		if err := cmd.Run(); err != nil {
			log.Fatalf("failed to git pull: %v", err)
		}
	}

	release := model.Release{
		Type:     model.ReleaseType(*releaseType),
		Replica:  ptr.To(int32(*replica)),
		Internet: model.Internet(*internet),
		Env:      *env,
	}

	ConfigYAML, err := render.RenderConfigMap(&release)
	if err != nil {
		log.Fatalf("RenderConfigmap failed: %v", err)
	}
	fmt.Println("---")
	fmt.Println(ConfigYAML)

	switch release.Type {
	case model.Normal:
		// 普通部署和灰度都生成一个 Service
		svcYAML, err := render.RenderService(&release)
		if err != nil {
			log.Fatalf("RenderService failed: %v", err)
		}
		fmt.Println("---")
		fmt.Println(svcYAML)

		deployYAML, err := render.RenderDeployment(&release)
		if err != nil {
			log.Fatalf("RenderDeploy failed: %v", err)
		}
		fmt.Println("---")
		fmt.Println(deployYAML)

	case model.Canary:
		svcYAML, err := render.RenderService(&release)
		if err != nil {
			log.Fatalf("RenderService failed: %v", err)
		}
		fmt.Println("---")
		fmt.Println(svcYAML)

		rolloutYAML, err := render.RenderRollout(&release)
		if err != nil {
			log.Fatalf("RenderRollout failed: %v", err)
		}
		fmt.Println("---")
		fmt.Println(rolloutYAML)

	case model.BlueGreen:
		// 蓝绿发布生成两个 Service：active 和 preview
		activeSvcYAML, err := render.RenderServiceBlueGreen(&release, "active")
		if err != nil {
			log.Fatalf("RenderService active failed: %v", err)
		}
		previewSvcYAML, err := render.RenderServiceBlueGreen(&release, "preview")
		if err != nil {
			log.Fatalf("RenderService preview failed: %v", err)
		}
		fmt.Println("---")
		fmt.Println(activeSvcYAML)
		fmt.Println("---")
		fmt.Println(previewSvcYAML)

		rolloutYAML, err := render.RenderRollout(&release)
		if err != nil {
			log.Fatalf("RenderRollout failed: %v", err)
		}
		fmt.Println("---")
		fmt.Println(rolloutYAML)
	}

	// -----------------------------
	// 3️⃣ 外网生成 VirtualService（仅 Canary / BlueGreen）
	// -----------------------------
	if release.Internet == model.External {
		vsYAML, err := render.RenderVirtualService(&release)
		if err != nil {
			log.Fatalf("RenderVirtualService failed: %v", err)
		}
		if vsYAML != "" {
			fmt.Println("---")
			fmt.Println(vsYAML)
		}
	}
}
