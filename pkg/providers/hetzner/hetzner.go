package hetzner

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/hetznercloud/hcloud-go/v2/hcloud"
	"github.com/koor-tech/genesis/pkg/config"
	"github.com/koor-tech/genesis/pkg/utils"
	"go.uber.org/fx"
)

const (
	defaultOSDSizeInGB = 100
)

type Provider struct {
	Client *hcloud.Client
	Token  string
	Logger *slog.Logger
}

type Params struct {
	fx.In

	Config *config.Config
	Logger *slog.Logger
}

func New(p Params) (*Provider, error) {
	prov := &Provider{
		Token:  p.Config.CloudProvider.Hetzner.Token,
		Client: hcloud.NewClient(hcloud.WithToken(p.Config.CloudProvider.Hetzner.Token)),
	}

	if err := prov.ConfigureCredentials(); err != nil {
		return nil, err
	}

	return prov, nil
}

func (p *Provider) ConfigureCredentials() error {
	err := os.Setenv("HCLOUD_TOKEN", p.Token)
	if err != nil {
		return fmt.Errorf("unable to set HCLOUD_TOKEN. %w", err)
	}

	return nil
}

func (p *Provider) buildLabels(customerName string) string {
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
	hetznerLabels := p.buildLabels(customerName)
	filterOpts := hcloud.ServerListOpts{
		ListOpts: hcloud.ListOpts{
			LabelSelector: hetznerLabels,
		},
	}
	servers, err := p.Client.Server.AllWithOpts(ctx, filterOpts)

	if err != nil {
		p.Logger.Error("unable to get server", "err", err)
		return nil, err
	}

	return servers, nil
}

func (p *Provider) AttacheVolumesToServers(ctx context.Context, customerName string, servers []*hcloud.Server) error {
	volLabels := map[string]string{"koor-client": customerName, "pool": "pool1"}
	for _, server := range servers {
		serverName := server.Name
		name := fmt.Sprintf("data-%s-pool1-%s", customerName, serverName[len(serverName)-5:])
		err := p.AttachVolumeToServer(ctx, server, volLabels, defaultOSDSizeInGB, name, false)
		if err != nil {
			p.Logger.Error("unable to create volume", "err", err)
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
		p.Logger.Error("unable to attach volumes due to", "err", err)
		return err
	}
	return nil
}
