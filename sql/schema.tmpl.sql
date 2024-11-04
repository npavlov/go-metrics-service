CREATE TYPE "metric_type" AS ENUM ('gauge', 'counter');
-- create "metrics" table
CREATE TABLE IF NOT EXISTS mtr_metrics (
                           "id" character varying(255) NOT NULL,
                           "type" "metric_type" NOT NULL,
                           "delta" bigint NULL,
                           "value" double precision NULL,
                           CONSTRAINT mtr_metrics_pk PRIMARY KEY (id, type)
);
CREATE INDEX IF NOT EXISTS  mtr_metrics_id_idx ON mtr_metrics (id);