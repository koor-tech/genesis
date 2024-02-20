package kubeone

import (
	"bytes"
	"context"
	"fmt"
	"github.com/hashicorp/terraform-exec/tfexec"
	"github.com/koor-tech/genesis/pkg/files"
	"github.com/koor-tech/genesis/pkg/models"
	"github.com/koor-tech/genesis/pkg/types"
	"gopkg.in/yaml.v3"
	"log"
	"os"
	"os/exec"
	"strings"
)

type Builder struct {
	config *models.TerraformConfig
	dst    string
}

func New(cluster *models.Cluster, dst string) *Builder {
	return &Builder{
		dst:    dst,
		config: models.NewTerraformConfig(cluster, dst),
	}
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
	err = files.SaveInFile(tfVars, strings.Join(tfConfigFile, "\n"), 0600)
	if err != nil {
		return err
	}
	return nil

}

func (b *Builder) WriteConfigFile() error {
	clusterConfig := models.NewKubeOneConfig()

	configData, err := yaml.Marshal(clusterConfig)
	if err != nil {
		return err
	}
	err = files.SaveInFile(b.dst+"/kubeone.yaml", string(configData), 0600)
	if err != nil {
		return err
	}
	return nil
}

func (b *Builder) RunTerraform() error {
	tf, err := tfexec.NewTerraform(b.dst, "/usr/bin/terraform")
	if err != nil {
		log.Printf("error running NewTerraform: %s", err)
		return err
	}

	err = tf.Init(context.Background(), tfexec.Upgrade(true))
	if err != nil {
		log.Fatalf("error running Init: %s", err)
	}
	fmt.Println("==========  terraform init  done ==========")
	fmt.Println("==========  running... terraform plan  ==========")
	_, err = tf.Plan(context.Background())
	if err != nil {
		log.Printf("error running Plan: %s", err)
		return err
	}
	fmt.Println("========== terraform plan done ==========")
	fmt.Println("==========  running... terraform init apply ==========")
	err = tf.Apply(context.Background())
	if err != nil {
		log.Fatalf("error running Apply: %s", err)
	}
	fmt.Println("========== terraform apply done ==========")

	fmt.Println("========== dump terraform file: terraform output -json -no-color > tf.json ==========")
	cmd := exec.Command("terraform", "output", "-json", "-no-color")
	cmd.Dir = b.dst

	outfile, err := os.Create(b.dst + "/tf.json")
	if err != nil {
		panic(err)
	}
	defer outfile.Close()

	// Direct the output to our file
	cmd.Stdout = outfile

	err = cmd.Run()
	if err != nil {
		log.Println(err)
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
