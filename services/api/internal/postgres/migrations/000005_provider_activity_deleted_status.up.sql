ALTER TABLE provider_activities
    DROP CONSTRAINT provider_activities_status_valid;

ALTER TABLE provider_activities
    ADD CONSTRAINT provider_activities_status_valid CHECK (status IN ('received', 'normalised', 'ignored', 'failed', 'deleted'));
