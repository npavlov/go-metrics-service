-- name: GetAllMetrics :many
SELECT id, type, delta, value FROM mtr_metrics;

-- name: GetMetric :one
SELECT id, type, delta, value FROM mtr_metrics
WHERE id = $1;

-- name: GetManyMetrics :many
SELECT id, type, delta, value FROM mtr_metrics
WHERE id = ANY($1::text[]);

-- name: InsertMetric :exec
INSERT INTO mtr_metrics (id, type, delta, value)
VALUES ($1, $2, $3, $4)
ON CONFLICT (id, type) DO NOTHING;

-- name: UpdateMetric :exec
UPDATE mtr_metrics
SET delta = $3, value = $4
WHERE id = $1 AND type = $2;

-- name: UpsertMetric :exec
INSERT INTO mtr_metrics (id, type, delta, value)
VALUES ($1, $2, $3, $4)
ON CONFLICT (id, type) DO UPDATE
    SET delta = EXCLUDED.delta,
        value = EXCLUDED.value;