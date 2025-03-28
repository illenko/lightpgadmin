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


CREATE TABLE event_log (
                       id UUID PRIMARY KEY DEFAULT gen_random_uuid() NOT NULL,
                       event_id UUID NOT NULL,
                       log_message TEXT NOT NULL,
                       created_at TIMESTAMP NOT NULL
);

INSERT INTO event_log (event_id, log_message, created_at)
VALUES
    (gen_random_uuid(), 'Event created', NOW()),
    (gen_random_uuid(), 'Event processed', NOW() + INTERVAL '1 hour'),
    (gen_random_uuid(), 'Event failed', NOW() + INTERVAL '2 hours');

CREATE TABLE event_error (
                       id UUID PRIMARY KEY DEFAULT gen_random_uuid() NOT NULL,
                       event_id UUID NOT NULL,
                       error_message TEXT NOT NULL,
                       created_at TIMESTAMP NOT NULL,
                       resolved_at TIMESTAMP);

INSERT INTO event_error (event_id, error_message, created_at, resolved_at)
VALUES
    (gen_random_uuid(), 'Error processing event', NOW(), NULL),
    (gen_random_uuid(), 'Network error', NOW() + INTERVAL '1 hour', NOW() + INTERVAL '2 hours'),
    (gen_random_uuid(), 'Timeout error', NOW() + INTERVAL '2 hours', NULL);