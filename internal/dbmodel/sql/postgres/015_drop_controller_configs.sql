-- 1_15.sql is a migration to delete controller configs
-- +goose Up

DROP TABLE controller_configs;
