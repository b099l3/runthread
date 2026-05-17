# Athlete Onboarding Profile Implementation Plan

This feature uses onboarding to understand who the athlete is, then adjusts plan language, coaching detail, and data prompts to fit their experience level.

The goal is to make beginners feel guided without jargon, while giving moderate and advanced runners enough detail to trust the training.

## Product Shape

During onboarding, the app asks for basic athlete context:

- Running experience level.
- Current weekly running volume.
- Recent training consistency.
- Main goal.
- Preferred training days.
- Access to heart-rate data or a heart-rate monitor.
- Confidence with training terms.

Experience level should affect how the app explains training.

Example beginner language:

```text
Run at a pace where you can speak in full sentences. This should feel comfortable.
```

Example moderate or advanced language:

```text
Keep this run mostly in Zone 2. A heart-rate monitor will make this easier to track accurately.
```

## Experience Levels

Initial levels:

- Beginner.
- Moderate.
- Advanced.

The app should not rely only on self-selection. It can also use training volume, consistency, race history, and connected activity data later to refine the profile.

## Adaptive Coaching Detail

Beginner experience:

- Avoid unnecessary jargon.
- Explain effort using feel and simple cues.
- Keep metrics secondary.
- Avoid overwhelming dashboards.
- Introduce concepts gradually.

Moderate experience:

- Use common training terms with short explanations.
- Show pace, effort, and heart-rate guidance where available.
- Suggest using a heart-rate monitor for better intensity control.
- Surface more detail in workout rationale.

Advanced experience:

- Show richer detail by default.
- Use zones, thresholds, progression logic, and training load language.
- Make heart-rate and intensity distribution more prominent.
- Provide more specific execution targets and report insights.

## Heart-Rate Monitor Prompting

The app should ask whether the athlete has access to reliable heart-rate data.

If the athlete is moderate or advanced and does not have reliable heart-rate data, the app can suggest using a heart-rate monitor because it improves:

- Easy-run discipline.
- Zone-based training.
- Readiness and fatigue interpretation.
- Intensity distribution reports.

This should be framed as helpful, not required.

## Backend Implementation Phases

### Phase 1: Domain Model

Extend the athlete profile with:

- Experience level.
- Current weekly volume.
- Training consistency.
- Goal context.
- Training terminology comfort.
- Heart-rate data availability.
- Heart-rate monitor usage.

These fields should remain provider-neutral.

### Phase 2: Onboarding Service

Add service methods to:

- Start onboarding.
- Save onboarding answers.
- Build or update an athlete profile.
- Mark onboarding complete.
- Return onboarding state to mobile.

Validation should be permissive enough to support partial onboarding and later edits.

### Phase 3: Coaching Detail Levels

Add a deterministic coaching detail level derived from onboarding:

- Simple.
- Standard.
- Detailed.

This level can drive workout rationale, readiness prompts, report wording, and metric visibility.

### Phase 4: Mobile Onboarding Flow

Build an onboarding flow that collects the required profile fields before the first plan is created.

The flow should:

- Use plain language.
- Keep the number of questions low.
- Allow edits later.
- Avoid blocking the user on provider connection.

### Phase 5: App-Wide Personalisation

Use the coaching detail level across:

- Workout descriptions.
- Workout rationale.
- Readiness prompts.
- Activity feedback prompts.
- Monthly reports.
- Heart-rate monitor suggestions.

## Acceptance Criteria

This feature is ready for a first MVP version when:

- New users can complete onboarding before plan creation.
- Athlete experience level is stored on the profile.
- The app derives a coaching detail level from onboarding answers.
- Beginner users see simpler workout language.
- Moderate and advanced users can see heart-rate and zone-based guidance where available.
- The app can suggest a heart-rate monitor when appropriate.
- Onboarding answers can be edited after setup.
- Tests cover profile validation, detail-level derivation, partial onboarding, and experience-specific wording selection.

## Assumptions

- The first version uses self-reported experience level.
- Provider activity history can refine the profile later but is not required for MVP onboarding.
- Heart-rate monitor suggestions should be optional and never block plan creation.
- Medical, nutrition, and coach marketplace onboarding are out of scope.
