package app_test

import (
	"context"
	"testing"
	"time"

	"github.com/actinc/cyberbiz-sdk/go/console/internal/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRetentionDeletesOldRows(t *testing.T) {
	env := newTestEnv(t)
	old := time.Now().Add(-40 * 24 * time.Hour)
	fresh := time.Now()
	require.NoError(t, env.app.DB.Create(&[]db.OutboundLog{
		{Method: "GET", Path: "/old", CreatedAt: old},
		{Method: "GET", Path: "/new", CreatedAt: fresh},
	}).Error)
	require.NoError(t, env.app.DB.Create(&[]db.InboundLog{
		{Event: "orders/paid", Status: db.InboundValid, CreatedAt: old},
		{Event: "orders/paid", Status: db.InboundValid, CreatedAt: fresh},
	}).Error)

	n, err := env.app.Retention.Run(context.Background())
	require.NoError(t, err)
	assert.Equal(t, int64(2), n)

	var out []db.OutboundLog
	require.NoError(t, env.app.DB.Find(&out).Error)
	require.Len(t, out, 1)
	assert.Equal(t, "/new", out[0].Path)
	var in int64
	require.NoError(t, env.app.DB.Model(&db.InboundLog{}).Count(&in).Error)
	assert.Equal(t, int64(1), in)
}
