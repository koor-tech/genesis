package hetzner

import (
	"context"
	"fmt"
	"github.com/hetznercloud/hcloud-go/v2/hcloud"
	"github.com/koor-tech/genesis/pkg/utils"
	"log"
	"strings"
)

type Provider struct {
	Client *hcloud.Client
}

func NewProvider() *Provider {
	cfg := NewConfig()
	return &Provider{
		Client: hcloud.NewClient(hcloud.WithToken(cfg.Token)),
	}
}

func (p *Provider) buildLabels(ctx context.Context, customerName string) string {
	clusterNameLabel := fmt.Sprintf("koor-client-%s", customerName)
	clusterWorkerLabelKey := fmt.Sprintf("koor-client-%s-workers", customerName)

	labels := map[string]string{
		"kubeone_cluster_name": clusterNameLabel,
		clusterWorkerLabelKey:  "pool1",
	}

	var hetznerLabels []string
	for labelKey, labelValue := range labels {
		hetznerLabels = append(hetznerLabels, fmt.Sprintf("%s=%s", labelKey, labelValue))
	}
	return strings.Join(hetznerLabels, ",")
}

func (p *Provider) GetServerByLabels(ctx context.Context, customerName string) ([]*hcloud.Server, error) {
	hetznerLabels := p.buildLabels(ctx, customerName)
	filterOpts := hcloud.ServerListOpts{
		ListOpts: hcloud.ListOpts{
			LabelSelector: hetznerLabels,
		},
	}
	servers, err := p.Client.Server.AllWithOpts(ctx, filterOpts)

	if err != nil {
		log.Printf("unable to get server, error: %q\n", err)
		return nil, err
	}

	return servers, nil
}

func (p *Provider) AttacheVolumesToServers(ctx context.Context, customerName string, servers []*hcloud.Server) error {
	volLabels := map[string]string{"koor-client": customerName, "pool": "pool1"}
	for _, server := range servers {
		serverName := server.Name
		name := fmt.Sprintf("data-%s-pool1-%s", customerName, serverName[len(serverName)-5:])
		err := p.AttachVolumeToServer(ctx, server, volLabels, 40, name, false)
		if err != nil {
			log.Print("unable to create volume:", "err", err)
			return err
		}
	}
	return nil
}

func (p *Provider) AttachVolumeToServer(ctx context.Context, server *hcloud.Server, labels map[string]string, sizeInGB int, volumeName string, autoMount bool) error {
	opts := hcloud.VolumeCreateOpts{
		Name:      volumeName,
		Size:      sizeInGB,
		Server:    server,
		Labels:    labels,
		Automount: utils.ToPointer(autoMount),
	}
	_, _, err := p.Client.Volume.Create(ctx, opts)
	if err != nil {
		log.Printf("unable to attach volumes due to: %q\n", err)
		return err
	}
	return nil
}
