-- 1_5.sql is a migration that adds a column necessary for model migrations
-- +goose Up

ALTER TABLE models ADD COLUMN migration_controller_id INTEGER REFERENCES controllers (id);
