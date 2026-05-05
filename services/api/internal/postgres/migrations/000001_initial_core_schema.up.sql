CREATE TABLE athlete_profiles (
    id uuid PRIMARY KEY,
    display_name text,
    experience_level text,
    current_weekly_distance_meters numeric NOT NULL DEFAULT 0,
    preferred_run_days smallint[] NOT NULL DEFAULT '{}',
    constraints text[] NOT NULL DEFAULT '{}',
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT athlete_profiles_current_weekly_distance_non_negative CHECK (current_weekly_distance_meters >= 0)
);

CREATE TABLE training_goals (
    id uuid PRIMARY KEY,
    athlete_id uuid NOT NULL REFERENCES athlete_profiles(id),
    type text NOT NULL,
    target_date date,
    target_distance_meters numeric,
    target_duration_seconds integer,
    notes text,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT training_goals_target_distance_non_negative CHECK (target_distance_meters IS NULL OR target_distance_meters >= 0),
    CONSTRAINT training_goals_target_duration_non_negative CHECK (target_duration_seconds IS NULL OR target_duration_seconds >= 0)
);

CREATE TABLE plan_weeks (
    id uuid PRIMARY KEY,
    athlete_id uuid NOT NULL REFERENCES athlete_profiles(id),
    goal_id uuid REFERENCES training_goals(id),
    plan_id uuid NOT NULL,
    week_index integer NOT NULL,
    starts_on date NOT NULL,
    focus text,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT plan_weeks_week_index_positive CHECK (week_index > 0),
    CONSTRAINT plan_weeks_plan_week_unique UNIQUE (plan_id, week_index)
);

CREATE TABLE planned_workouts (
    id uuid PRIMARY KEY,
    plan_week_id uuid NOT NULL REFERENCES plan_weeks(id) ON DELETE CASCADE,
    plan_id uuid NOT NULL,
    scheduled_for date NOT NULL,
    type text NOT NULL,
    status text NOT NULL,
    target_distance_meters numeric,
    target_duration_seconds integer,
    intensity_kind text,
    intensity_description text,
    notes text,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT planned_workouts_target_distance_non_negative CHECK (target_distance_meters IS NULL OR target_distance_meters >= 0),
    CONSTRAINT planned_workouts_target_duration_non_negative CHECK (target_duration_seconds IS NULL OR target_duration_seconds >= 0)
);

CREATE TABLE imported_activities (
    id uuid PRIMARY KEY,
    athlete_id uuid NOT NULL REFERENCES athlete_profiles(id),
    type text NOT NULL,
    started_at timestamptz NOT NULL,
    duration_seconds integer NOT NULL,
    distance_meters numeric,
    average_pace_seconds_per_kilometer integer,
    average_heart_bpm integer,
    created_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT imported_activities_duration_positive CHECK (duration_seconds > 0),
    CONSTRAINT imported_activities_distance_non_negative CHECK (distance_meters IS NULL OR distance_meters >= 0),
    CONSTRAINT imported_activities_average_pace_non_negative CHECK (average_pace_seconds_per_kilometer IS NULL OR average_pace_seconds_per_kilometer >= 0),
    CONSTRAINT imported_activities_average_heart_bpm_non_negative CHECK (average_heart_bpm IS NULL OR average_heart_bpm >= 0)
);

CREATE TABLE workout_matches (
    id uuid PRIMARY KEY,
    planned_workout_id uuid NOT NULL REFERENCES planned_workouts(id),
    imported_activity_id uuid NOT NULL REFERENCES imported_activities(id),
    status text NOT NULL,
    confidence text NOT NULL,
    matched_by text NOT NULL,
    matched_at timestamptz NOT NULL,
    notes text,
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE workout_results (
    id uuid PRIMARY KEY,
    planned_workout_id uuid NOT NULL REFERENCES planned_workouts(id),
    imported_activity_id uuid REFERENCES imported_activities(id),
    outcome text NOT NULL,
    completed_at timestamptz,
    distance_meters numeric,
    duration_seconds integer,
    notes text,
    created_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT workout_results_distance_non_negative CHECK (distance_meters IS NULL OR distance_meters >= 0),
    CONSTRAINT workout_results_duration_non_negative CHECK (duration_seconds IS NULL OR duration_seconds >= 0)
);

CREATE TABLE adaptation_events (
    id uuid PRIMARY KEY,
    plan_id uuid NOT NULL,
    athlete_id uuid NOT NULL REFERENCES athlete_profiles(id),
    type text NOT NULL,
    reason text NOT NULL,
    summary text NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE adaptation_event_changes (
    id uuid PRIMARY KEY,
    adaptation_event_id uuid NOT NULL REFERENCES adaptation_events(id) ON DELETE CASCADE,
    planned_workout_id uuid REFERENCES planned_workouts(id),
    type text NOT NULL,
    description text NOT NULL,
    position integer NOT NULL,
    CONSTRAINT adaptation_event_changes_position_non_negative CHECK (position >= 0)
);

CREATE INDEX training_goals_athlete_id_idx ON training_goals (athlete_id);
CREATE INDEX plan_weeks_athlete_id_starts_on_idx ON plan_weeks (athlete_id, starts_on);
CREATE INDEX plan_weeks_plan_id_week_index_idx ON plan_weeks (plan_id, week_index);
CREATE INDEX planned_workouts_plan_week_id_scheduled_for_idx ON planned_workouts (plan_week_id, scheduled_for);
CREATE INDEX imported_activities_athlete_id_started_at_idx ON imported_activities (athlete_id, started_at);
CREATE INDEX workout_matches_planned_workout_id_idx ON workout_matches (planned_workout_id);
CREATE INDEX workout_matches_imported_activity_id_idx ON workout_matches (imported_activity_id);
CREATE INDEX workout_results_planned_workout_id_idx ON workout_results (planned_workout_id);
CREATE INDEX workout_results_imported_activity_id_idx ON workout_results (imported_activity_id);
CREATE INDEX adaptation_events_athlete_id_created_at_idx ON adaptation_events (athlete_id, created_at);
CREATE INDEX adaptation_events_plan_id_created_at_idx ON adaptation_events (plan_id, created_at);
CREATE INDEX adaptation_event_changes_event_id_position_idx ON adaptation_event_changes (adaptation_event_id, position);

