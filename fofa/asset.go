package fofa

import (
	"net"
	"net/url"
)

type Asset struct {
	IP  net.IP   `json:"ip"`
	URL *url.URL `json:"url"`
}
