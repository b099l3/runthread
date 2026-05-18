UPDATE provider_activities
SET status = 'ignored',
    last_error = COALESCE(NULLIF(last_error, ''), 'provider activity deleted')
WHERE status = 'deleted';

ALTER TABLE provider_activities
    DROP CONSTRAINT provider_activities_status_valid;

ALTER TABLE provider_activities
    ADD CONSTRAINT provider_activities_status_valid CHECK (status IN ('received', 'normalised', 'ignored', 'failed'));
