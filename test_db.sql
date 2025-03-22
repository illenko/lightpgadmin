CREATE TABLE event (
                       id UUID PRIMARY KEY DEFAULT gen_random_uuid() NOT NULL,
                       tenant_id INT NOT NULL,
                       entity_id UUID NOT NULL,
                       created_at TIMESTAMP NOT NULL,
                       updated_at TIMESTAMP NOT NULL,
                       occurred_at TIMESTAMP NOT NULL,
                       scheduled_at TIMESTAMP,
                       processed_at TIMESTAMP,
                       status VARCHAR(255) NOT NULL,
                       attempt INT NOT NULL,
                       mock_time_millis BIGINT NOT NULL,
                       mock_error BOOLEAN NOT NULL,
                       payload JSONB
);

INSERT INTO event (tenant_id, entity_id, created_at, updated_at, occurred_at, scheduled_at, processed_at, status, attempt, mock_time_millis, mock_error, payload)
VALUES
    (1, gen_random_uuid(), NOW(), NOW(), NOW(), NOW() + INTERVAL '1 day', NOW() + INTERVAL '2 days', 'pending', 1, 1000, false, '{"key": "value"}'::jsonb),
    (2, gen_random_uuid(), NOW(), NOW(), NOW(), NOW() + INTERVAL '2 days', NOW() + INTERVAL '3 days', 'completed', 2, 2000, true, '{"key": "value2"}'::jsonb),
    (3, gen_random_uuid(), NOW(), NOW(), NOW(), NOW() + INTERVAL '3 days', NOW() + INTERVAL '4 days', 'failed', 3, 3000, false, '{"key": "value3"}'::jsonb);