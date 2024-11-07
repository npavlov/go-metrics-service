-- +goose Up
-- create enum type "metric_type"
CREATE TYPE "metric_type" AS ENUM ('gauge', 'counter');
-- create "mtr_metrics" table
CREATE TABLE "mtr_metrics" (
  "id" character varying(255) NOT NULL,
  "type" "metric_type" NOT NULL,
  PRIMARY KEY ("id", "type"),
  CONSTRAINT "mtr_metrics_id_key" UNIQUE ("id")
);
-- create index "mtr_metrics_id_idx" to table: "mtr_metrics"
CREATE INDEX "mtr_metrics_id_idx" ON "mtr_metrics" ("id");
-- create "counter_metrics" table
CREATE TABLE "counter_metrics" (
  "metric_id" character varying(255) NOT NULL,
  "delta" bigint NOT NULL,
  PRIMARY KEY ("metric_id"),
  CONSTRAINT "unique_counter_id" UNIQUE ("metric_id"),
  CONSTRAINT "fk_metric" FOREIGN KEY ("metric_id") REFERENCES "mtr_metrics" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- create "gauge_metrics" table
CREATE TABLE "gauge_metrics" (
  "metric_id" character varying(255) NOT NULL,
  "value" double precision NOT NULL,
  PRIMARY KEY ("metric_id"),
  CONSTRAINT "unique_gauge_id" UNIQUE ("metric_id"),
  CONSTRAINT "fk_metric" FOREIGN KEY ("metric_id") REFERENCES "mtr_metrics" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);

-- +goose Down
-- reverse: create "gauge_metrics" table
DROP TABLE "gauge_metrics";
-- reverse: create "counter_metrics" table
DROP TABLE "counter_metrics";
-- reverse: create index "mtr_metrics_id_idx" to table: "mtr_metrics"
DROP INDEX "mtr_metrics_id_idx";
-- reverse: create "mtr_metrics" table
DROP TABLE "mtr_metrics";
-- reverse: create enum type "metric_type"
DROP TYPE "metric_type";
