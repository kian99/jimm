-- 1_7.sql is a migration that adds a UUID column to the groups table.
-- +goose Up

ALTER TABLE groups ADD COLUMN uuid TEXT NOT NULL UNIQUE;
