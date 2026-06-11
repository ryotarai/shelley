package browse

import (
	"reflect"
	"testing"

	"golang.org/x/net/http/httpproxy"
)

func TestProxyFlagsFromEnv(t *testing.T) {
	tests := []struct {
		name string
		cfg  *httpproxy.Config
		want []proxyFlag
	}{
		{
			name: "no proxy configured",
			cfg:  &httpproxy.Config{},
			want: nil,
		},
		{
			name: "https proxy only",
			cfg:  &httpproxy.Config{HTTPSProxy: "http://proxy:8080"},
			want: []proxyFlag{{Name: "proxy-server", Value: "http://proxy:8080"}},
		},
		{
			name: "http proxy only falls back to http",
			cfg:  &httpproxy.Config{HTTPProxy: "http://httponly:3128"},
			want: []proxyFlag{{Name: "proxy-server", Value: "http://httponly:3128"}},
		},
		{
			name: "https takes precedence over http",
			cfg: &httpproxy.Config{
				HTTPProxy:  "http://httpproxy:3128",
				HTTPSProxy: "http://httpsproxy:8080",
			},
			want: []proxyFlag{{Name: "proxy-server", Value: "http://httpsproxy:8080"}},
		},
		{
			name: "server and bypass list",
			cfg: &httpproxy.Config{
				HTTPSProxy: "http://proxy:8080",
				NoProxy:    "localhost,127.0.0.1,.internal.example.com",
			},
			want: []proxyFlag{
				{Name: "proxy-server", Value: "http://proxy:8080"},
				{Name: "proxy-bypass-list", Value: "localhost,127.0.0.1,.internal.example.com"},
			},
		},
		{
			name: "bypass list is normalized: trim spaces and drop empties",
			cfg: &httpproxy.Config{
				HTTPSProxy: "http://proxy:8080",
				NoProxy:    " localhost , , 127.0.0.1 ,",
			},
			want: []proxyFlag{
				{Name: "proxy-server", Value: "http://proxy:8080"},
				{Name: "proxy-bypass-list", Value: "localhost,127.0.0.1"},
			},
		},
		{
			name: "no_proxy without a proxy server yields no flags",
			cfg:  &httpproxy.Config{NoProxy: "localhost,127.0.0.1"},
			want: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := proxyFlagsFromEnv(tt.cfg)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("proxyFlagsFromEnv() = %#v, want %#v", got, tt.want)
			}
		})
	}
}

func TestRedactProxyServerForLog(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  string
	}{
		{
			name:  "no credentials is unchanged",
			value: "http://proxy.example.com:8080",
			want:  "http://proxy.example.com:8080",
		},
		{
			name:  "user and password are redacted",
			value: "http://user:secret@proxy.example.com:8080",
			want:  "http://proxy.example.com:8080",
		},
		{
			name:  "username only is redacted",
			value: "http://user@proxy.example.com:8080",
			want:  "http://proxy.example.com:8080",
		},
		{
			name:  "value without scheme is unchanged",
			value: "proxy.example.com:8080",
			want:  "proxy.example.com:8080",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := redactProxyServerForLog(tt.value); got != tt.want {
				t.Errorf("redactProxyServerForLog(%q) = %q, want %q", tt.value, got, tt.want)
			}
		})
	}
}
