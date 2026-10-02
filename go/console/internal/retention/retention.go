// Package retention deletes Outbound and Inbound rows older than the
// configured number of days, at boot and then every 24 hours.
package retention

import (
	"context"
	"time"

	"github.com/actinc/cyberbiz-sdk/go/console/internal/inbound"
	"github.com/actinc/cyberbiz-sdk/go/console/internal/outbound"
	"github.com/rs/zerolog/log"
)

// Job prunes the two log tables.
type Job struct {
	outbound *outbound.Repository
	inbound  *inbound.Repository
	days     int
	now      func() time.Time
}

// New returns a Job keeping rows for the given number of days; zero
// disables pruning.
func New(o *outbound.Repository, i *inbound.Repository, days int) *Job {
	return &Job{outbound: o, inbound: i, days: days, now: time.Now}
}

// Run prunes once and returns the number of rows deleted.
func (j *Job) Run(ctx context.Context) (int64, error) {
	if j.days <= 0 {
		return 0, nil
	}
	cutoff := j.now().Add(-time.Duration(j.days) * 24 * time.Hour)
	nOut, err := j.outbound.DeleteBefore(ctx, cutoff)
	if err != nil {
		return 0, err
	}
	nIn, err := j.inbound.DeleteBefore(ctx, cutoff)
	if err != nil {
		return nOut, err
	}
	log.Info().Int64("outbound", nOut).Int64("inbound", nIn).Time("cutoff", cutoff).Msg("retention pruned")
	return nOut + nIn, nil
}

// Start runs the job now and then every interval until ctx is done.
func (j *Job) Start(ctx context.Context, interval time.Duration) {
	go func() {
		if _, err := j.Run(ctx); err != nil {
			log.Error().Err(err).Msg("retention run failed")
		}
		t := time.NewTicker(interval)
		defer t.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-t.C:
				if _, err := j.Run(ctx); err != nil {
					log.Error().Err(err).Msg("retention run failed")
				}
			}
		}
	}()
}
