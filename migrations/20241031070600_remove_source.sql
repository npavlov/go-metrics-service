-- +goose Up
-- modify "metrics" table
ALTER TABLE "metrics" DROP COLUMN "m_source";

-- +goose Down
-- reverse: modify "metrics" table
ALTER TABLE "metrics" ADD COLUMN "m_source" text NULL;
