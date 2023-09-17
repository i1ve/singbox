package provider

import (
	"context"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/sagernet/sing-box/adapter"
	C "github.com/sagernet/sing-box/constant"
	"github.com/sagernet/sing-box/log"
	"github.com/sagernet/sing-box/option"
	O "github.com/sagernet/sing-box/outbound"
	E "github.com/sagernet/sing/common/exceptions"
)

type SubscriptionInfo struct {
	upload   uint64
	download uint64
	total    uint64
	expire   uint64
}

type myProviderAdapter struct {
	subscriptionInfo SubscriptionInfo
	logger           log.ContextLogger
	tag              string
	path             string
	healthCheckUrl   string
	providerType     string
	updateTime       time.Time
	outbounds        []adapter.Outbound
	outboundByTag    map[string]adapter.Outbound
}

func (a *myProviderAdapter) Tag() string {
	return a.tag
}

func (a *myProviderAdapter) Path() string {
	return a.path
}

func (a *myProviderAdapter) Type() string {
	return a.providerType
}

func (a *myProviderAdapter) HealthCheckUrl() string {
	return a.healthCheckUrl
}

func (a *myProviderAdapter) UpdateTime() time.Time {
	return a.updateTime
}

func (a *myProviderAdapter) Outbound(tag string) (adapter.Outbound, bool) {
	outbound, loaded := a.outboundByTag[tag]
	return outbound, loaded
}

func (a *myProviderAdapter) Outbounds() []adapter.Outbound {
	return a.outbounds
}

func GetFirstLine(content []byte) string {
	firstLine := strings.Split(string(content), "\n")[0]
	firstLine = strings.Trim(firstLine, " ")
	firstLine = strings.Trim(firstLine, "\t")
	return firstLine
}

func (a *myProviderAdapter) SubscriptionInfo() map[string]uint64 {
	info := make(map[string]uint64)
	info["Upload"] = a.subscriptionInfo.upload
	info["Download"] = a.subscriptionInfo.download
	info["Total"] = a.subscriptionInfo.total
	info["Expire"] = a.subscriptionInfo.expire
	return info
}

func (a *myProviderAdapter) ParseSubInfo(infoString string) {
	reg := regexp.MustCompile("^#[ \t]*upload=(\\d+);[ \t]*download=(\\d+);[ \t]*total=(\\d+);[ \t]*expire=(\\d*);$")
	result := reg.FindStringSubmatch(infoString)
	if len(result) > 0 {
		upload, _ := strconv.Atoi(result[1:][0])
		download, _ := strconv.Atoi(result[1:][1])
		total, _ := strconv.Atoi(result[1:][2])
		expire, _ := strconv.Atoi(result[1:][3])
		a.subscriptionInfo.upload = uint64(upload)
		a.subscriptionInfo.download = uint64(download)
		a.subscriptionInfo.total = uint64(total)
		a.subscriptionInfo.expire = uint64(expire)
	}
}

func (a *myProviderAdapter) CreateOutboundFromContent(ctx context.Context, router adapter.Router, outbounds []option.Outbound) error {
	for _, outbound := range outbounds {
		otype := outbound.Type
		tag := outbound.Tag
		switch otype {
		case C.TypeDirect, C.TypeBlock, C.TypeDNS, C.TypeSelector, C.TypeURLTest:
			continue
		default:
			out, err := O.New(ctx, router, a.logger, tag, outbound)
			if err != nil {
				E.New("invalid outbound")
				continue
			}
			a.outboundByTag[tag] = out
			a.outbounds = append(a.outbounds, out)
		}
	}
	return nil
}

func (p *myProviderAdapter) GetContentFromFile(router adapter.Router) []byte {
	content := []byte{}
	updateTime := time.Unix(int64(0), int64(0))
	path := p.path
	fileInfo, err := os.Stat(path)
	if !os.IsNotExist(err) {
		updateTime = fileInfo.ModTime()
		content, _ = os.ReadFile(path)
		p.ParseSubInfo(GetFirstLine(content))
	}
	p.updateTime = updateTime
	return content
}

func (p *myProviderAdapter) BackupProvider() ([]adapter.Outbound, map[string]adapter.Outbound, SubscriptionInfo) {
	outboundsBackup := p.outbounds
	outboundByTagBackup := p.outboundByTag
	subscriptionInfoBackup := p.subscriptionInfo
	p.outbounds = []adapter.Outbound{}
	p.outboundByTag = make(map[string]adapter.Outbound)
	p.subscriptionInfo = SubscriptionInfo{
		upload:   uint64(0),
		download: uint64(0),
		total:    uint64(0),
		expire:   uint64(0),
	}
	return outboundsBackup, outboundByTagBackup, subscriptionInfoBackup
}

func (p *myProviderAdapter) RevertProvider(outboundsBackup []adapter.Outbound, outboundByTagBackup map[string]adapter.Outbound, subscriptionInfoBackup SubscriptionInfo) {
	p.outbounds = outboundsBackup
	p.outboundByTag = outboundByTagBackup
	p.subscriptionInfo = subscriptionInfoBackup
}

func (p *myProviderAdapter) UpdateGroups(router adapter.Router) error {
	for _, outbound := range router.Outbounds() {
		if group, ok := outbound.(adapter.OutboundGroup); ok {
			err := group.UpdateOutbounds(p.tag)
			if err != nil {
				return E.Cause(err, "update provider ", p.tag, " failed")
			}
		}
	}
	return nil
}
