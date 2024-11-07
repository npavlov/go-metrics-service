-- name: GetAllMetrics :many
SELECT m.id,
       m.type,
       COALESCE(c.delta, 0) AS delta,
       COALESCE(g.value, 0.0) AS value
FROM mtr_metrics AS m
         LEFT JOIN counter_metrics AS c ON m.id = c.metric_id
         LEFT JOIN gauge_metrics AS g ON m.id = g.metric_id;

-- name: GetUnifiedMetric :one
SELECT m.id,
       m.type,
       COALESCE(c.delta, 0) AS delta,
       COALESCE(g.value, 0.0) AS value
FROM mtr_metrics AS m
         LEFT JOIN counter_metrics AS c ON m.id = c.metric_id
         LEFT JOIN gauge_metrics AS g ON m.id = g.metric_id
WHERE m.id = $1;

-- name: GetManyMetrics :many
SELECT m.id,
       m.type,
       COALESCE(c.delta, 0) AS delta,
       COALESCE(g.value, 0.0) AS value
FROM mtr_metrics AS m
         LEFT JOIN counter_metrics AS c ON m.id = c.metric_id
         LEFT JOIN gauge_metrics AS g ON m.id = g.metric_id
WHERE m.id = ANY($1::text[]);

-- name: InsertMtrMetric :exec
INSERT INTO mtr_metrics (id, type)
VALUES ($1, $2)
ON CONFLICT (id, type) DO NOTHING;

-- name: InsertCounterMetric :exec
INSERT INTO counter_metrics (metric_id, delta)
VALUES ($1, $2)
ON CONFLICT (metric_id) DO NOTHING;

-- name: InsertGaugeMetric :exec
INSERT INTO gauge_metrics (metric_id, value)
VALUES ($1, $2)
ON CONFLICT (metric_id) DO NOTHING;

-- name: UpdateCounterMetric :exec
UPDATE counter_metrics
SET delta = $2
WHERE metric_id = $1;

-- name: UpdateGaugeMetric :exec
UPDATE gauge_metrics
SET value = $2
WHERE metric_id = $1;

-- name: UpsertMtrMetric :exec
-- Insert into mtr_metrics or update if conflict on (id, type)
INSERT INTO mtr_metrics (id, type)
VALUES ($1, $2)
ON CONFLICT (id, type) DO UPDATE
    SET id = EXCLUDED.id, type = EXCLUDED.type; 

-- name: UpsertCounterMetric :exec
-- Insert into counter_metrics or update if conflict on (metric_id)
INSERT INTO counter_metrics (metric_id, delta)
VALUES ($1, $2)
ON CONFLICT (metric_id) DO UPDATE
    SET delta = EXCLUDED.delta;

-- name: UpsertGaugeMetric :exec
-- Insert into gauge_metrics or update if conflict on (metric_id)
INSERT INTO gauge_metrics (metric_id, value)
VALUES ($1, $2)
ON CONFLICT (metric_id) DO UPDATE
    SET value = EXCLUDED.value;