-- 1_8.sql is a migration that adds a tls_hostname column to the controller table.
-- +goose Up

ALTER TABLE controllers ADD COLUMN tls_hostname TEXT;
