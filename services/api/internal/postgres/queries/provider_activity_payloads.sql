-- name: CreateProviderActivityPayload :one
INSERT INTO provider_activity_payloads (
    id,
    provider_activity_id,
    payload,
    payload_kind,
    received_at
) VALUES (
    sqlc.arg(id),
    sqlc.arg(provider_activity_id),
    sqlc.arg(payload),
    sqlc.arg(payload_kind),
    sqlc.arg(received_at)
)
RETURNING *;

-- name: GetProviderActivityPayload :one
SELECT *
FROM provider_activity_payloads
WHERE id = sqlc.arg(id);

-- name: UpdateProviderActivityPayload :one
UPDATE provider_activity_payloads
SET
    provider_activity_id = sqlc.arg(provider_activity_id),
    payload = sqlc.arg(payload),
    payload_kind = sqlc.arg(payload_kind),
    received_at = sqlc.arg(received_at)
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: ListProviderActivityPayloads :many
SELECT *
FROM provider_activity_payloads
WHERE provider_activity_id = sqlc.arg(provider_activity_id)
ORDER BY received_at DESC, id;
