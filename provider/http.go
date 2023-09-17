package provider

import (
	"context"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/sagernet/sing-box/adapter"
	C "github.com/sagernet/sing-box/constant"
	"github.com/sagernet/sing-box/log"
	"github.com/sagernet/sing-box/option"
	E "github.com/sagernet/sing/common/exceptions"
	M "github.com/sagernet/sing/common/metadata"
)

var _ adapter.OutboundProvider = (*HTTPProvider)(nil)

type HTTPProvider struct {
	myProviderAdapter
	url    string
	ua     string
	detour string
	start  bool
}

func (p *HTTPProvider) Start() bool {
	return p.start
}

func (p *HTTPProvider) SetStart() {
	p.start = true
}

func (p *HTTPProvider) FetchHTTP(httpClient *http.Client, parsedURL *url.URL) ([]byte, string, error) {
	request, err := http.NewRequest("GET", parsedURL.String(), nil)
	if err != nil {
		return nil, "", err
	}
	ua := p.ua
	if ua == "" {
		ua = "singbox"
	}
	request.Header.Add("User-Agent", ua)
	response, err := httpClient.Do(request)
	if err != nil {
		return nil, "", err
	}
	defer response.Body.Close()
	content, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, "", err
	}
	subInfo := response.Header.Get("subscription-userinfo")
	if subInfo != "" {
		subInfo = "# " + subInfo + ";"
		content = append([]byte(subInfo+"\n"), content...)
	}
	return content, subInfo, nil
}

func (p *HTTPProvider) FetchContent(router adapter.Router) ([]byte, string, error) {
	detour := router.DefaultOutboundForConnection()
	if p.detour != "" {
		if outbound, ok := router.Outbound(p.detour); ok {
			detour = outbound
		}
	}
	httpClient := &http.Client{
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				return detour.DialContext(ctx, network, M.ParseSocksaddr(addr))
			},
			ForceAttemptHTTP2: true,
		},
	}
	defer httpClient.CloseIdleConnections()
	parsedURL, err := url.Parse(p.url)
	if err != nil {
		return nil, "", err
	}
	switch parsedURL.Scheme {
	case "":
		parsedURL.Scheme = "http"
		fallthrough
	case "http", "https":
		content, subInfo, err := p.FetchHTTP(httpClient, parsedURL)
		if err != nil {
			return nil, "", err
		}
		return content, subInfo, nil
	default:
		return nil, "", E.New("invalid url scheme")
	}
}

func (p *HTTPProvider) ContentFromHTTP(router adapter.Router) []byte {
	content, subInfoRaw, err := p.FetchContent(router)
	if err != nil {
		E.Cause(err, "fetch provider ", p.tag, " failed")
		return nil
	}
	path := p.path
	p.ParseSubInfo(subInfoRaw)
	p.updateTime = time.Now()
	file, _ := os.OpenFile(path, os.O_CREATE, 0777)
	defer file.Close()
	file.Write(content)
	return content
}

func (p *HTTPProvider) GetContent(router adapter.Router) []byte {
	if !p.start {
		p.start = true
		return p.GetContentFromFile(router)
	}
	return p.ContentFromHTTP(router)
}

func (p *HTTPProvider) ParseProvider(ctx context.Context, router adapter.Router) error {
	content := p.GetContent(router)
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

func (p *HTTPProvider) UpdateProvider(ctx context.Context, router adapter.Router) error {
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

func NewHTTPProvider(ctx context.Context, router adapter.Router, logger log.ContextLogger, options option.OutboundProvider) (*HTTPProvider, error) {
	httpOptions := options.HTTPOptions
	url := httpOptions.Url
	if url == "" {
		return nil, E.New("provider download url missing")
	}
	provider := &HTTPProvider{
		myProviderAdapter: myProviderAdapter{
			logger:         logger,
			tag:            options.Tag,
			path:           options.Path,
			healthCheckUrl: options.HealthCheckUrl,
			providerType:   C.TypeHTTPProvider,
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
		url:    httpOptions.Url,
		ua:     httpOptions.UserAgent,
		detour: httpOptions.Detour,
		start:  false,
	}
	err := provider.ParseProvider(ctx, router)
	if err != nil {
		return nil, err
	}
	return provider, nil
}
