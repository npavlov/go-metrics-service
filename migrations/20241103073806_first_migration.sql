-- +goose Up
-- create enum type "metric_type"
CREATE TYPE "metric_type" AS ENUM ('gauge', 'counter');
-- create "mtr_metrics" table
CREATE TABLE "mtr_metrics" (
  "id" character varying(255) NOT NULL,
  "type" "metric_type" NOT NULL,
  "delta" bigint NULL,
  "value" double precision NULL,
  PRIMARY KEY ("id", "type")
);
-- create index "mtr_metrics_id_idx" to table: "mtr_metrics"
CREATE INDEX "mtr_metrics_id_idx" ON "mtr_metrics" ("id");

-- +goose Down
-- reverse: create index "mtr_metrics_id_idx" to table: "mtr_metrics"
DROP INDEX "mtr_metrics_id_idx";
-- reverse: create "mtr_metrics" table
DROP TABLE "mtr_metrics";
-- reverse: create enum type "metric_type"
DROP TYPE "metric_type";
