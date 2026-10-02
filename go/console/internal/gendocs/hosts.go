package gendocs

import (
	"regexp"
	"strings"
	"sync"

	"github.com/actinc/cyberbiz-sdk/go/internal/redact"
)

// synthShopHost is the host every merchant hostname becomes in a Sample.
const synthShopHost = "example.cyberbiz.co"

// platformSubdomains are CYBERBIZ-owned hostnames that identify no merchant.
var platformSubdomains = map[string]bool{
	"api": true, "app-store-api": true, "app": true, "www": true, "api-doc": true, "example": true,
}

// platformHost reports whether a *.cyberbiz.co / *.cyberbiz.io host is
// CYBERBIZ's own rather than a Shop's.
func platformHost(host string) bool {
	return platformSubdomains[strings.ToLower(strings.SplitN(host, ".", 2)[0])]
}

// denylist is redact.Denylist, compiled once per process.
var denylist = sync.OnceValues(redact.Denylist)

// denylistHost matches any host containing a denylisted name; nil when the
// list is unset. An invalid list yields nil here and is reported by
// scanOutput, which runs on every generated file.
var denylistHost = sync.OnceValue(func() *regexp.Regexp {
	names, err := denylist()
	if err != nil || names == nil {
		return nil
	}
	return regexp.MustCompile(`[a-z0-9.-]*(?:` + names.String() + `)[a-z0-9.-]*\.(?:io|co|com)`)
})

// synthHosts rewrites every Shop subdomain, and every host that contains a
// denylisted integrator or merchant name, to the example Shop.
func synthHosts(v string) string {
	v = reShopHost.ReplaceAllStringFunc(v, func(m string) string {
		if platformHost(m) {
			return m
		}
		return synthShopHost
	})
	if re := denylistHost(); re != nil {
		v = re.ReplaceAllString(v, synthShopHost)
	}
	return v
}
