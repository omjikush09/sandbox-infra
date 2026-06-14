package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"sync"
	"syscall"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/omjikush09/sandboxing-infra/packages/vm/pool"
	"github.com/omjikush09/sandboxing-infra/packages/vm/router"
)

const firecrackerLabPath = "/home/ubuntu/firecracker-lab"

const (
	rootfsBucket = "firecracker-rootfs-bucket"
	rootfsKey    = "firecracker/rootfs/node-agent-rootfs.ext4.zst"
	rootfsRegion = "us-east-1"
)

func main() {
	ctx := context.Background()
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer cancel()
	err := initSystem()

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	go pool.InitPoolManager(ctx, 8)

	app := fiber.New()
	app.Use(cors.New())

	router.Start(app)
	wg := sync.WaitGroup{}

	wg.Go(func() {
		<-ctx.Done()
		app.Shutdown()
	})

	PORT := "8000"
	if err := app.Listen(":" + PORT); err != nil {
		fmt.Println(err.Error())
	}

	wg.Wait()

}

func initSystem() error {
	if err := os.MkdirAll(firecrackerLabPath, 0755); err != nil {
		return err
	}

	rootfsPath := firecrackerLabPath + "/rootfs.ext4"
	if _, err := os.Stat(rootfsPath); err != nil {
		if !os.IsNotExist(err) {
			return err
		}
		if err := downloadRootfs(firecrackerLabPath); err != nil {
			return err
		}
	}

	kernelPath := firecrackerLabPath + "/vmlinux.bin"
	if _, err := os.Stat(kernelPath); err != nil {
		if !os.IsNotExist(err) {
			return err
		}
		if err := downloadFile(
			"https://s3.amazonaws.com/spec.ccfc.min/img/quickstart_guide/x86_64/kernels/vmlinux.bin",
			kernelPath,
		); err != nil {
			return err
		}
	}

	return nil
}

func downloadRootfs(folderPath string) error {
	compressedRootfsPath := folderPath + "/rootfs.ext4.zst"

	cfg, err := config.LoadDefaultConfig(context.Background(), config.WithRegion(rootfsRegion))
	if err != nil {
		return err
	}

	s3Client := s3.NewFromConfig(cfg)

	resp, err := s3Client.GetObject(context.Background(), &s3.GetObjectInput{
		Bucket: aws.String(rootfsBucket),
		Key:    aws.String(rootfsKey),
	})
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	file, err := os.Create(compressedRootfsPath)
	if err != nil {
		return err
	}

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		file.Close()
		return err
	}
	if err := file.Close(); err != nil {
		return err
	}

	if err := exec.Command("zstd", "-d", compressedRootfsPath).Run(); err != nil {
		return err
	}

	return nil
}

func downloadFile(url, path string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("download failed for %s with status %s", url, resp.Status)
	}

	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(file, resp.Body)
	return err
}
