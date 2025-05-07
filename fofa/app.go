package fofa

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"

	"github.com/yoshino-s/go-framework/application"
	"github.com/yoshino-s/go-framework/configuration"
)

var _ application.Application = &FofaApp{}

type FofaApp struct {
	*application.EmptyApplication
	config config
}

func New() *FofaApp {
	return &FofaApp{
		EmptyApplication: application.NewEmptyApplication("Fofa"),
	}
}

func (f *FofaApp) Configuration() configuration.Configuration {
	return &f.config
}

func (f *FofaApp) Setup(context.Context) {
	if err := f.check(); err != nil {
		panic(err)
	}
}

func (a *FofaApp) Query(query string, page int, size int) ([]Asset, error) {
	url, err := url.Parse(a.config.Endpoint + "/search/all")
	if err != nil {
		return nil, err
	}

	qBase64 := base64.StdEncoding.EncodeToString([]byte(query))

	queryParams := url.Query()
	queryParams.Set("email", a.config.Email)
	queryParams.Set("fields", "host,ip,port,protocol")
	queryParams.Set("full", "false")
	queryParams.Set("key", a.config.Key)
	queryParams.Set("page", fmt.Sprintf("%d", page))
	queryParams.Set("qbase64", qBase64)
	queryParams.Set("size", fmt.Sprintf("%d", size))
	url.RawQuery = queryParams.Encode()

	r, err := http.Get(url.String())
	if err != nil {
		return nil, err
	}

	var resp struct {
		ErrMsg  string     `json:"errmsg"`
		Error   bool       `json:"error"`
		Results [][]string `json:"results"`
	}

	if err := json.NewDecoder(r.Body).Decode(&resp); err != nil {
		return nil, err
	}

	if resp.Error && resp.ErrMsg != "" {
		return nil, fmt.Errorf("fofa error response: %s", resp.ErrMsg)
	}

	res := make([]Asset, 0, len(resp.Results))

	for _, item := range resp.Results {
		host := item[0]
		ip := item[1]
		port := item[2]
		protocol := item[3]

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
			IP:  net.ParseIP(ip),
			URL: url,
		})
	}

	return res, nil
}

func (a *FofaApp) check() error {
	url, err := url.Parse(a.config.Endpoint + "/info/my")
	if err != nil {
		return err
	}
	query := url.Query()
	query.Set("email", a.config.Email)
	query.Set("key", a.config.Key)
	url.RawQuery = query.Encode()

	r, err := http.Get(url.String())
	if err != nil {
		return err
	}

	var resp struct {
		ErrMsg string `json:"errmsg"`
		Error  bool   `json:"error"`
	}

	if err := json.NewDecoder(r.Body).Decode(&resp); err != nil {
		return err
	}

	if resp.Error && resp.ErrMsg != "" {
		return fmt.Errorf("fofa error response: %s", resp.ErrMsg)
	}

	return nil
}
