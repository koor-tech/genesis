package kubeone

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"strings"

	"github.com/hashicorp/terraform-exec/tfexec"
	"github.com/koor-tech/genesis/pkg/models"
	"github.com/koor-tech/genesis/pkg/providers/hetzner"
	"github.com/koor-tech/genesis/pkg/utils"
	types "github.com/koor-tech/genesis/pkg/utils/types"
	"gopkg.in/yaml.v3"
)

type Builder struct {
	logger     *slog.Logger
	config     *models.TerraformConfig
	dst        string
	CloudToken string
}

const terraformExec = "/usr/local/bin/terraform"

func New(logger *slog.Logger, cluster *models.Cluster, dst string, cloudProvider *hetzner.Provider) (*Builder, error) {
	if err := cloudProvider.ConfigureCredentials(); err != nil {
		return nil, err
	}

	return &Builder{
		logger:     logger,
		CloudToken: cloudProvider.Token,
		dst:        dst,
		config:     models.NewTerraformConfig(cluster, dst),
	}, nil
}

func (b *Builder) WriteTFVars() error {
	tfConfigMap, err := types.ExtractData(*b.config)
	if err != nil {
		return err
	}

	var tfConfigFile []string

	for key, value := range tfConfigMap {
		configLine := fmt.Sprintf("%s=\"%v\"", key, value)
		tfConfigFile = append(tfConfigFile, configLine)
	}

	tfVars := fmt.Sprintf("%s/%s", b.dst, "terraform.tfvars")
	err = utils.SaveInFile(tfVars, strings.Join(tfConfigFile, "\n"), 0600)
	if err != nil {
		return err
	}
	return nil

}

func (b *Builder) WriteConfigFile(clusterName string) error {
	clusterConfig := models.NewKubeOneConfig(clusterName)

	configData, err := yaml.Marshal(clusterConfig)
	if err != nil {
		return err
	}
	err = utils.SaveInFile(b.dst+"/kubeone.yaml", string(configData), 0600)
	if err != nil {
		return err
	}
	return nil
}

func (b *Builder) RunTerraform(ctx context.Context) error {
	tf, err := tfexec.NewTerraform(b.dst, terraformExec)
	if err != nil {
		return fmt.Errorf("error running NewTerraform. %w", err)
	}

	err = tf.Init(ctx, tfexec.Upgrade(true))
	if err != nil {
		return fmt.Errorf("error running init. %w", err)
	}
	fmt.Println("==========  terraform init  done ==========")
	fmt.Println("==========  running... terraform plan  ==========")
	_, err = tf.Plan(ctx)
	if err != nil {
		return fmt.Errorf("error running Plan: %w", err)
	}
	fmt.Println("========== terraform plan done ==========")
	fmt.Println("==========  running... terraform apply ==========")
	err = tf.Apply(ctx)

	if err != nil {
		b.logger.Error("error running apply", "err", err)
		return err
	}

	fmt.Println("========== terraform apply done ==========")

	fmt.Println("========== dump terraform file: terraform output -json -no-color > tf.json ==========")
	cmd := exec.Command("terraform", "output", "-json", "-no-color")
	cmd.Dir = b.dst

	outfile, err := os.Create(b.dst + "/tf.json")
	if err != nil {
		return err
	}
	defer outfile.Close()

	// Direct the output to our file
	cmd.Stdout = outfile

	err = cmd.Run()
	if err != nil {
		b.logger.Error("failed to run terraform", "err", err)
		return err
	}

	fmt.Println("========== terraform output done ==========")
	return nil
}

func (b *Builder) RunKubeOne() (string, error) {
	cmd := exec.Command("kubeone", "apply", "-m", b.dst+"/kubeone.yaml", "-t", "tf.json", "-y")
	cmd.Dir = b.dst

	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		fmt.Println("Executing: " + cmd.String())
		return "", err
	}

	return out.String(), nil
}
