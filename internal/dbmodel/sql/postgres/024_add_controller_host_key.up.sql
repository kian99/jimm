-- Adds a host key column to the controller table.
-- The column can be NULL as not all controller's 
-- have a host key.

ALTER TABLE controllers ADD COLUMN ssh_host_key BYTEA;
