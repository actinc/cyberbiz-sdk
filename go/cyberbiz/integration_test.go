//go:build integration

package cyberbiz_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/actinc/cyberbiz-sdk/go/cyberbiz"
)

// Integration tests hit the live CYBERBIZ API with the token in
// CYBERBIZ_API_TOKEN. They are read-only and skipped without a token.
//
//	go test -tags integration ./cyberbiz
func liveClient(t *testing.T) *cyberbiz.Client {
	t.Helper()
	token := os.Getenv("CYBERBIZ_API_TOKEN")
	if token == "" {
		t.Skip("CYBERBIZ_API_TOKEN not set")
	}
	c, err := cyberbiz.New(token)
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func liveContext(t *testing.T) context.Context {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	t.Cleanup(cancel)
	return ctx
}

func TestLiveShopInfo(t *testing.T) {
	c := liveClient(t)
	info, resp, err := c.Shop.Info(liveContext(t))
	if err != nil {
		t.Fatal(err)
	}
	if info == nil || info.ID == 0 || info.PrimaryDomain == "" {
		t.Errorf("unexpected shop info: %+v", info)
	}
	if resp.RequestID == "" {
		t.Error("no X-Request-Id")
	}
}

func TestLiveAppSettings(t *testing.T) {
	c := liveClient(t)
	settings, _, err := c.Shop.Settings(liveContext(t))
	if err != nil {
		t.Fatal(err)
	}
	if settings == nil || settings.ID == 0 || settings.AddOnVersion == nil {
		t.Errorf("unexpected settings: %+v", settings)
	}
}
