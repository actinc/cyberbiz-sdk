// Package db opens the SQLite database and owns the GORM models.
package db

import (
	"fmt"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Open opens (creating when needed) the SQLite file at path and migrates
// every model.
func Open(path string) (*gorm.DB, error) {
	dsn := path + "?_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)&_pragma=foreign_keys(0)"
	g, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{Logger: logger.Discard})
	if err != nil {
		return nil, fmt.Errorf("db: open %s: %w", path, err)
	}
	if err := g.AutoMigrate(&User{}, &Session{}, &Shop{}, &OutboundLog{}, &InboundLog{}); err != nil {
		return nil, fmt.Errorf("db: migrate: %w", err)
	}
	return g, nil
}
