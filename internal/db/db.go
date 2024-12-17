// Copyright 2024 Canonical.

// Package db contains routines to store and retrieve data from a database.
package db

import (
	"context"
	"database/sql"
	"fmt"
	"path"
	"sync/atomic"
	"time"

	"github.com/pressly/goose/v3"
	"gorm.io/gorm"

	"github.com/canonical/jimm/v3/internal/dbmodel"
	"github.com/canonical/jimm/v3/internal/errors"
)

// A Database provides access to the database model. A Database instance
// is safe to use from multiple goroutines.
type Database struct {
	// DB contains the gorm database storing the data.
	DB *gorm.DB

	// migrated holds whether the database has been successfully migrated
	// to the current database version. The value of migrated should always
	// be read using atomic.LoadUint32 and will contain a 0 if the
	// migration is yet to be run, or 1 if it has been run successfully.
	migrated uint32
}

// Transaction starts a new transaction using the database. This allows
// a set of changes to be performed as a single atomic unit. All of the
// transaction steps should be performed in the given function, if this
// function returns an error then all changes in the transaction will be
// aborted and the error returned. Transactions may be nested.
//
// Attempting to start a transaction on an unmigrated database will result
// in an error with a code of errors.CodeUpgradeInProgress.
func (d *Database) Transaction(f func(*Database) error) error {
	if err := d.ready(); err != nil {
		return err
	}
	return d.DB.Transaction(func(tx *gorm.DB) error {
		d := *d
		d.DB = tx
		return f(&d)
	})
}

// Migrate migrates the configured database to have the structure required
// by the current data model. Unless forced the migration will only be
// performed if there is either no current database, or the major version
// of the current database is the same as the target database. If the
// current database is incompatible then an error with a code of
// errors.CodeServerConfiguration will be returned. If force is requested
// then the migration will be performed no matter what the current version
// is. The force parameter should only be set when the migration is
// initiated by a user request.
func (d *Database) Migrate(ctx context.Context, force bool) error {
	const op = errors.Op("db.Migrate")
	if d == nil || d.DB == nil {
		return errors.E(op, errors.CodeServerConfiguration, "database not configured")
	}
	db := d.DB.WithContext(ctx)

	goose.SetBaseFS(dbmodel.SQL)

	if err := goose.SetDialect("postgres"); err != nil {
		return errors.E(op, fmt.Errorf("failed to migration dialect: %w", err))
	}
	sqlDB, err := db.DB()
	if err != nil {
		return errors.E(op, "unable to obtain raw DB")
	}

	if err := goose.Up(sqlDB, path.Join("sql", db.Name())); err != nil {
		return errors.E(op, fmt.Errorf("failed to migrate db: %w", err))
	}
	atomic.StoreUint32(&d.migrated, 1)
	return nil
}

// ready checks that the database is ready to accept requests. An error is
// returned if the database is not yet initialised.
func (d *Database) ready() error {
	if d == nil || d.DB == nil {
		return errors.E(errors.CodeServerConfiguration, "database not configured")
	}
	if atomic.LoadUint32(&d.migrated) == 0 {
		return errors.E(errors.CodeUpgradeInProgress)
	}
	return nil
}

// Close closes open connections to the underlying database backend.
func (d *Database) Close() error {
	sqlDB, err := d.DB.DB()
	if err != nil {
		return errors.E(err, "failed to get the internal DB object")
	}
	if err := sqlDB.Close(); err != nil {
		return errors.E(err, "failed to close database connection")
	}
	return nil
}

// Now returns the current time as a valid sql.NullTime. The time that is
// returned is in UTC and is truncated to milliseconds which is the
// resolution supported on all databases.
func Now() sql.NullTime {
	return sql.NullTime{
		Time:  time.Now().UTC().Truncate(time.Millisecond),
		Valid: true,
	}
}
