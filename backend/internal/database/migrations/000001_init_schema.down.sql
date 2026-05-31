-- Full teardown of the initial schema. Drops everything created by the up
-- migration. Destructive — used only to roll the baseline all the way back.
DROP SCHEMA public CASCADE;
CREATE SCHEMA public;
