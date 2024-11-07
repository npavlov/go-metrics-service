-- Define metric type enum if not already defined
CREATE TYPE "metric_type" AS ENUM ('gauge', 'counter');

-- Main metrics table to store only metadata (normalized structure)
CREATE TABLE IF NOT EXISTS mtr_metrics (
                                           "id" VARCHAR(255) NOT NULL,
                                           "type" "metric_type" NOT NULL,
                                           PRIMARY KEY ("id", "type"),
                                           UNIQUE ("id")  -- Unique constraint added to allow foreign key referencing
);

-- Table for counter metrics, storing the delta value only
CREATE TABLE IF NOT EXISTS counter_metrics (
                                               "metric_id" VARCHAR(255) NOT NULL,
                                               "delta" BIGINT NOT NULL,
                                               PRIMARY KEY ("metric_id"),
                                               CONSTRAINT fk_metric FOREIGN KEY ("metric_id")
                                                   REFERENCES mtr_metrics ("id") ON DELETE CASCADE
);

-- Table for gauge metrics, storing the value only
CREATE TABLE IF NOT EXISTS gauge_metrics (
                                             "metric_id" VARCHAR(255) NOT NULL,
                                             "value" DOUBLE PRECISION NOT NULL,
                                             PRIMARY KEY ("metric_id"),
                                             CONSTRAINT fk_metric FOREIGN KEY ("metric_id")
                                                 REFERENCES mtr_metrics ("id") ON DELETE CASCADE
);

-- Optional: Add index on metric ID for quick lookups by ID
CREATE INDEX IF NOT EXISTS mtr_metrics_id_idx ON mtr_metrics ("id");

-- Optional: Add a unique constraint to enforce one entry per metric in either counter or gauge tables
ALTER TABLE counter_metrics ADD CONSTRAINT unique_counter_id UNIQUE ("metric_id");
ALTER TABLE gauge_metrics ADD CONSTRAINT unique_gauge_id UNIQUE ("metric_id");