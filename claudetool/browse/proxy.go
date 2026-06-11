package browse

import (
	"net/url"
	"strings"

	"golang.org/x/net/http/httpproxy"
)

// proxyFlag is a single headless-shell command-line flag (without the leading
// "--") and its value, e.g. {"proxy-server", "http://proxy:8080"}.
type proxyFlag struct {
	Name  string
	Value string
}

// proxyFlagsFromEnv translates standard proxy environment variables into the
// headless-shell flags --proxy-server and --proxy-bypass-list.
//
//   - proxy-server: HTTPS_PROXY if set, otherwise HTTP_PROXY. A single proxy is
//     used for all schemes (the common corporate-proxy case).
//   - proxy-bypass-list: NO_PROXY, normalized (entries trimmed, empties dropped).
//     Chromium accepts a comma- or semicolon-separated list and only honors it
//     when a proxy server is set, so it is omitted when no server is configured.
//
// Returns nil when no proxy server is configured.
func proxyFlagsFromEnv(cfg *httpproxy.Config) []proxyFlag {
	server := cfg.HTTPSProxy
	if server == "" {
		server = cfg.HTTPProxy
	}
	if server == "" {
		return nil
	}

	flags := []proxyFlag{{Name: "proxy-server", Value: server}}
	if bypass := normalizeBypassList(cfg.NoProxy); bypass != "" {
		flags = append(flags, proxyFlag{Name: "proxy-bypass-list", Value: bypass})
	}
	return flags
}

// redactProxyServerForLog strips any userinfo (user:password) from a proxy
// server URL so credentials embedded in HTTPS_PROXY/HTTP_PROXY are not leaked
// to logs. Values that do not parse or carry no userinfo are returned as-is.
func redactProxyServerForLog(value string) string {
	u, err := url.Parse(value)
	if err != nil || u.User == nil {
		return value
	}
	u.User = nil
	return u.String()
}

// normalizeBypassList parses a NO_PROXY-style comma-separated list, trims each
// entry, drops empties, and rejoins with commas.
func normalizeBypassList(noProxy string) string {
	parts := strings.Split(noProxy, ",")
	out := parts[:0]
	for _, p := range parts {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return strings.Join(out, ",")
}
