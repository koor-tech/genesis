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

func NewProvider(token string) *Provider {
	return &Provider{
		Client: hcloud.NewClient(hcloud.WithToken(token)),
	}
}

func (p *Provider) buildLabels(labels map[string]string) string {
	var hetznerLabels []string
	for labelKey, labelValue := range labels {
		hetznerLabels = append(hetznerLabels, fmt.Sprintf("%s=%s", labelKey, labelValue))
	}
	return strings.Join(hetznerLabels, ",")
}

func (p *Provider) GetServerByLabels(ctx context.Context, labels map[string]string) ([]*hcloud.Server, error) {
	hetznerLabels := p.buildLabels(labels)

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
