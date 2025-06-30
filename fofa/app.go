package fofa

import (
	"context"
	"encoding/base64"
	"fmt"
	"net"
	"net/url"
	"strings"

	"github.com/yoshino-s/go-framework/application"
	"github.com/yoshino-s/go-framework/configuration"
	"resty.dev/v3"
)

var _ application.Application = &FofaApp{}

type FofaApp struct {
	*application.EmptyApplication
	config config
	client *resty.Client
}

func New() *FofaApp {
	return &FofaApp{
		EmptyApplication: application.NewEmptyApplication("Fofa"),
		client:           resty.New(),
	}
}

func (f *FofaApp) Configuration() configuration.Configuration {
	return &f.config
}

func (f *FofaApp) Initialize(ctx context.Context) {
	if err := f.Check(ctx); err != nil {
		panic(err)
	}
}

func (a *FofaApp) Query(ctx context.Context, query string, page int, size int, options ...WithQueryOption) ([]Asset, error) {
	queryOptions := &queryOptions{
		fields: []string{
			"host", "ip", "port", "protocol",
		},
	}
	for _, opt := range options {
		opt(queryOptions)
	}

	var resp struct {
		ErrMsg  string     `json:"errmsg"`
		Error   bool       `json:"error"`
		Results [][]string `json:"results"`
	}

	_, err := a.client.R().
		SetContext(ctx).
		SetQueryParams(map[string]string{
			"email":   a.config.Email,
			"key":     a.config.Key,
			"fields":  strings.Join(queryOptions.fields, ","),
			"full":    "false",
			"page":    fmt.Sprintf("%d", page),
			"size":    fmt.Sprintf("%d", size),
			"qbase64": base64.StdEncoding.EncodeToString([]byte(query)),
		}).
		SetResult(&resp).
		Get(a.config.Endpoint + "/search/all")

	if err != nil {
		return nil, err
	}

	if resp.Error && resp.ErrMsg != "" {
		return nil, fmt.Errorf("fofa error response: %s", resp.ErrMsg)
	}

	res := make([]Asset, 0, len(resp.Results))

	for _, item := range resp.Results {
		rawMap := make(map[string]string)
		for i, field := range queryOptions.fields {
			rawMap[field] = item[i]
		}

		host := rawMap["host"]
		ip := net.ParseIP(rawMap["ip"])
		port := rawMap["port"]
		protocol := rawMap["protocol"]

		if protocol == "" {
			if port == "443" {
				protocol = "https"
			} else {
				protocol = "http"
			}
		}

		if host == "" {
			host = fmt.Sprintf("%s:%s", ip, port)
		}

		if !strings.HasPrefix(host, protocol+"://") {
			host = fmt.Sprintf("%s://%s", protocol, host)
		}

		url, err := url.Parse(host)
		if err != nil {
			return nil, err
		}

		res = append(res, Asset{
			IP:  ip,
			URL: url,
			Raw: rawMap,
		})
	}

	return res, nil
}

func (a *FofaApp) Check(ctx context.Context) error {
	var resp struct {
		ErrMsg string `json:"errmsg"`
		Error  bool   `json:"error"`
	}

	_, err := a.client.R().
		SetContext(ctx).
		SetQueryParams(map[string]string{
			"email": a.config.Email,
			"key":   a.config.Key,
		}).
		SetResult(&resp).
		Get(a.config.Endpoint + "/info/my")

	if err != nil {
		return err
	}

	if resp.Error && resp.ErrMsg != "" {
		return fmt.Errorf("fofa error response: %s", resp.ErrMsg)
	}

	return nil
}
