ALTER TABLE athlete_profiles
ALTER COLUMN preferred_run_days TYPE smallint[]
USING preferred_run_days::smallint[];
