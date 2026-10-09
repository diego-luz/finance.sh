package middlewares

import (
	"log/slog"
	"net"
	"net/http"
	"net/netip"
	"strings"
	"sync"
)

// DefaultTrustedProxies are the peers whose X-Forwarded-For / X-Real-IP is
// believed when TRUSTED_PROXIES is not set: loopback only. Trusting the private
// ranges by default let any machine on the LAN forge its address, and behind
// rootless Docker every client arrives from a 172.x gateway. A proxy elsewhere
// (another container, another host) has to be listed in TRUSTED_PROXIES.
var DefaultTrustedProxies = []string{"127.0.0.0/8", "::1/128"}

// avisoProxy logs, once, a forwarded header from a private peer that is not
// trusted: almost always a reverse proxy that still needs TRUSTED_PROXIES.
var avisoProxy sync.Once

// ParseTrustedProxies turns a list of IPs or CIDRs into prefixes, skipping (and
// logging) anything it cannot parse.
func ParseTrustedProxies(itens []string) []netip.Prefix {
	var out []netip.Prefix
	for _, item := range itens {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		if p, err := netip.ParsePrefix(item); err == nil {
			out = append(out, p.Masked())
			continue
		}
		if a, err := netip.ParseAddr(item); err == nil {
			out = append(out, netip.PrefixFrom(a.Unmap(), a.Unmap().BitLen()))
			continue
		}
		slog.Warn("TRUSTED_PROXIES: ignoring invalid entry", "entry", item)
	}
	return out
}

// RealIP replaces chi's middleware.RealIP, which believes True-Client-IP,
// X-Real-IP and X-Forwarded-For from anyone: a client could send a new address
// on every request and escape the per-IP rate limit, and forge the IP written
// to the audit log and the session list.
//
// Here the headers only count when the TCP peer is a trusted proxy. The client
// is then the rightmost X-Forwarded-For address that is not itself a trusted
// proxy (entries to its left were written by the client and prove nothing),
// falling back to X-Real-IP. Otherwise RemoteAddr is left alone.
func RealIP(trusted []netip.Prefix) func(http.Handler) http.Handler {
	confia := func(a netip.Addr) bool {
		a = a.Unmap()
		for _, p := range trusted {
			if p.Contains(a) {
				return true
			}
		}
		return false
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if ip, ok := clientIP(r, confia); ok {
				r.RemoteAddr = ip
			}
			next.ServeHTTP(w, r)
		})
	}
}

func clientIP(r *http.Request, confia func(netip.Addr) bool) (string, bool) {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		host = r.RemoteAddr
	}
	par, err := netip.ParseAddr(host)
	if err != nil {
		return "", false
	}
	if !confia(par) {
		if par.IsPrivate() && (r.Header.Get("X-Forwarded-For") != "" || r.Header.Get("X-Real-IP") != "") {
			avisoProxy.Do(func() {
				slog.Warn("X-Forwarded-For ignored from an untrusted private address; if a reverse proxy runs there, "+
					"add it to TRUSTED_PROXIES (e.g. TRUSTED_PROXIES=172.16.0.0/12 for one on the Docker network), "+
					"otherwise every client shares the proxy's rate limit", "peer", par.String())
			})
		}
		return "", false
	}
	var cadeia []string
	for _, h := range r.Header.Values("X-Forwarded-For") {
		cadeia = append(cadeia, strings.Split(h, ",")...)
	}
	// every hop trusted (a LAN client behind the proxy): the leftmost one
	ultimo := ""
	for i := len(cadeia) - 1; i >= 0; i-- {
		a, err := netip.ParseAddr(strings.TrimSpace(cadeia[i]))
		if err != nil {
			break // malformed hop: nothing to its left can be trusted
		}
		if !confia(a) {
			return a.Unmap().String(), true
		}
		ultimo = a.Unmap().String()
	}
	if ultimo != "" {
		return ultimo, true
	}
	if a, err := netip.ParseAddr(strings.TrimSpace(r.Header.Get("X-Real-IP"))); err == nil {
		return a.Unmap().String(), true
	}
	return "", false
}
