package provider

import (
	"context"
	"time"

	"github.com/sagernet/sing-box/adapter"
	C "github.com/sagernet/sing-box/constant"
	"github.com/sagernet/sing-box/log"
	"github.com/sagernet/sing-box/option"
	E "github.com/sagernet/sing/common/exceptions"
)

var _ adapter.OutboundProvider = (*FileProvider)(nil)

type FileProvider struct {
	myProviderAdapter
}

func (p *FileProvider) ParseProvider(ctx context.Context, router adapter.Router) error {
	content := p.GetContentFromFile(router)
	var options option.Options
	err := options.UnmarshalJSON(content)
	if err != nil {
		return E.Cause(err, "decode config at ", p.path)
	}
	err = p.CreateOutboundFromContent(ctx, router, options.Outbounds)
	if err != nil {
		return err
	}
	return nil
}

func (p *FileProvider) UpdateProvider(ctx context.Context, router adapter.Router) error {
	outboundsBackup, outboundByTagBackup, subscriptionInfoBackup := p.BackupProvider()
	err := p.ParseProvider(ctx, router)
	if err != nil {
		p.RevertProvider(outboundsBackup, outboundByTagBackup, subscriptionInfoBackup)
		return err
	}
	err = p.UpdateGroups(router)
	if err != nil {
		p.RevertProvider(outboundsBackup, outboundByTagBackup, subscriptionInfoBackup)
		return err
	}
	return nil
}

func NewFileProvider(ctx context.Context, router adapter.Router, logger log.ContextLogger, options option.OutboundProvider) (*FileProvider, error) {
	provider := &FileProvider{
		myProviderAdapter: myProviderAdapter{
			logger:         logger,
			tag:            options.Tag,
			path:           options.Path,
			healthCheckUrl: options.HealthCheckUrl,
			providerType:   C.TypeFileProvider,
			updateTime:     time.Unix(int64(0), int64(0)),
			subscriptionInfo: SubscriptionInfo{
				upload:   0,
				download: 0,
				total:    0,
				expire:   0,
			},
			outbounds:     []adapter.Outbound{},
			outboundByTag: make(map[string]adapter.Outbound),
		},
	}
	err := provider.ParseProvider(ctx, router)
	if err != nil {
		return nil, err
	}
	return provider, nil
}
