# Runthread Mobile

Flutter MVP app for Runthread.

The main screen is the weekly plan view. It calls the local Go backend's `GetCurrentPlanWeek` ConnectRPC endpoint and renders the returned seven-day plan with workout details, loading state, and error state.

The MVP also includes:

- Workout detail view.
- Recent activity and adaptation history tab.
- Demo fallback when the local backend is unavailable or unseeded.
- Disabled completion affordance that explains real completion will come from imported provider activity later.

## Local Development

From the repository root:

```sh
asdf install
```

Run the backend separately:

```sh
cd services/api
asdf exec go run ./cmd/server
```

Run the mobile app:

```sh
cd apps/mobile
asdf exec flutter run
```

The default entrypoint uses local defaults, including `http://localhost:8080`.
For emulator and physical-device development, create mobile-only env files from
the tracked examples:

```sh
cp env/mobile.emulator.env.example env/mobile.emulator.env
cp env/mobile.physical.env.example env/mobile.physical.env
```

Then run the target-specific entrypoint:

```sh
asdf exec flutter run -t lib/main_emulator.dart
asdf exec flutter run -t lib/main_physical.dart
```

The mobile env files are bundled into debug app builds. Keep them limited to
public client configuration such as API URLs, demo athlete IDs, demo goal IDs,
and OAuth redirect URIs. Do not put backend secrets such as `DATABASE_URL`,
Strava client secret, webhook tokens, or encryption keys in `apps/mobile/env`.

## Demo Bootstrap

The weekly plan screen first tries the local backend. If the backend is unavailable or does not have the current demo athlete and goal records, the app falls back to local demo plan data and shows a visible `Demo data` notice on the plan and history tabs.

This keeps the Flutter MVP usable without hidden setup while auth, onboarding, and seeded beta data remain deferred.

## Current API Boundary

The app has a small API boundary in `lib/src/api/runthread_api.dart`.

Dart protobuf/ConnectRPC generation is not wired yet, so the current client uses Connect's JSON protocol directly. Replace this with generated Dart RPC code once the mobile proto generation workflow is chosen.

The current backend read path is demo-shaped. The mobile app sends demo `athlete-1` and `goal-1` identifiers to `GetCurrentPlanWeek`. `lib/src/demo` owns the local fallback data and should stay separate from the API client.

The mobile app does not call `CompleteImportedActivity` yet. That RPC is still demo-shaped because it accepts full athlete, goal, and imported activity payloads. Production completion should come from provider-imported activities matched by the backend.

Before beta use, the app still needs auth-backed athlete lookup, seeded or persisted current plan data, generated Dart protobuf/ConnectRPC bindings, real imported activity data, and production completion/adaptation refresh flows.

## Verification

```sh
asdf exec flutter analyze
asdf exec flutter test
```
