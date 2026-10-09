package middlewares

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRealIP(t *testing.T) {
	trusted := ParseTrustedProxies([]string{"127.0.0.0/8", "::1/128", "172.16.0.0/12", "10.0.0.0/8"})
	cases := []struct {
		name   string
		remote string
		xff    []string
		xreal  string
		want   string
	}{
		{"internet client forging XFF is ignored", "203.0.113.9:5555", []string{"1.2.3.4"}, "", "203.0.113.9:5555"},
		{"internet client forging X-Real-IP is ignored", "203.0.113.9:5555", nil, "1.2.3.4", "203.0.113.9:5555"},
		{"proxy on loopback: client from XFF", "127.0.0.1:4000", []string{"198.51.100.7"}, "", "198.51.100.7"},
		{"docker proxy: rightmost untrusted wins", "172.18.0.5:4000", []string{"6.6.6.6, 198.51.100.7"}, "", "198.51.100.7"},
		{"forged left entry does not win", "172.18.0.5:4000", []string{"6.6.6.6", "198.51.100.7, 10.0.0.2"}, "", "198.51.100.7"},
		{"LAN client behind proxy: leftmost private", "172.18.0.5:4000", []string{"192.168.1.20"}, "", "192.168.1.20"},
		{"proxy without XFF: X-Real-IP", "127.0.0.1:4000", nil, "198.51.100.7", "198.51.100.7"},
		{"proxy without headers keeps peer", "127.0.0.1:4000", nil, "", "127.0.0.1:4000"},
		{"malformed hop stops the walk", "127.0.0.1:4000", []string{"6.6.6.6, lixo, 10.0.0.3"}, "", "10.0.0.3"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var got string
			h := RealIP(trusted)(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) { got = r.RemoteAddr }))
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.RemoteAddr = tc.remote
			for _, v := range tc.xff {
				req.Header.Add("X-Forwarded-For", v)
			}
			if tc.xreal != "" {
				req.Header.Set("X-Real-IP", tc.xreal)
			}
			h.ServeHTTP(httptest.NewRecorder(), req)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestParseTrustedProxies(t *testing.T) {
	got := ParseTrustedProxies([]string{"10.1.2.3", " 192.168.0.0/16 ", "lixo", ""})
	assert.Len(t, got, 2)
	assert.Equal(t, "10.1.2.3/32", got[0].String())
}

func TestRealIPDefaultTrustsLoopbackOnly(t *testing.T) {
	trusted := ParseTrustedProxies(DefaultTrustedProxies)
	for remote, want := range map[string]string{
		"127.0.0.1:4000":    "198.51.100.7", // proxy on the same host
		"192.168.1.50:4000": "192.168.1.50:4000",
		"172.18.0.1:4000":   "172.18.0.1:4000", // docker gateway: needs TRUSTED_PROXIES
	} {
		var got string
		h := RealIP(trusted)(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) { got = r.RemoteAddr }))
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.RemoteAddr = remote
		req.Header.Set("X-Forwarded-For", "198.51.100.7")
		h.ServeHTTP(httptest.NewRecorder(), req)
		assert.Equal(t, want, got, remote)
	}
}
