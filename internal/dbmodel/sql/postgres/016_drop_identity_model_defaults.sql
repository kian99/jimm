-- 1_16.sql is a migration to delete identitymodel defaults
-- +goose Up

DROP TABLE identity_model_defaults;
