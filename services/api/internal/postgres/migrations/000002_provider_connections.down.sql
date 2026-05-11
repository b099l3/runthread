DROP INDEX IF EXISTS provider_import_events_provider_delivery_unique_idx;
DROP INDEX IF EXISTS provider_import_events_status_idx;
DROP INDEX IF EXISTS provider_import_events_activity_id_idx;
DROP INDEX IF EXISTS provider_import_events_connection_received_at_idx;

DROP INDEX IF EXISTS provider_activity_payloads_activity_id_received_at_idx;

DROP INDEX IF EXISTS provider_activities_provider_activity_id_idx;
DROP INDEX IF EXISTS provider_activities_status_idx;
DROP INDEX IF EXISTS provider_activities_imported_activity_id_idx;
DROP INDEX IF EXISTS provider_activities_athlete_started_at_idx;

DROP INDEX IF EXISTS provider_connections_status_idx;
DROP INDEX IF EXISTS provider_connections_provider_user_id_idx;
DROP INDEX IF EXISTS provider_connections_athlete_provider_idx;

DROP TABLE IF EXISTS provider_import_events;
DROP TABLE IF EXISTS provider_activity_payloads;
DROP TABLE IF EXISTS provider_activities;
DROP TABLE IF EXISTS provider_connections;
