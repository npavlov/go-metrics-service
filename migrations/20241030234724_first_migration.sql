-- +goose Up
-- create "metrics" table
CREATE TABLE "metrics" (
  "id" character varying(255) NOT NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  "type" text NOT NULL,
  "m_source" text NULL,
  "delta" bigint NULL,
  "value" double precision NULL,
  PRIMARY KEY ("id")
);
-- create index "idx_metrics_deleted_at" to table: "metrics"
CREATE INDEX "idx_metrics_deleted_at" ON "metrics" ("deleted_at");

-- +goose Down
-- reverse: create index "idx_metrics_deleted_at" to table: "metrics"
DROP INDEX "idx_metrics_deleted_at";
-- reverse: create "metrics" table
DROP TABLE "metrics";
