ALTER TABLE athlete_profiles
ALTER COLUMN preferred_run_days TYPE bigint[]
USING preferred_run_days::bigint[];
