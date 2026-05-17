# Release Plan

Runthread uses GitHub Actions for CI and Render for private-beta API deploys.

## Release Tracks

### API

The API release path is:

1. Open a pull request.
2. Wait for Backend CI to pass.
3. Merge to `main`.
4. Let Render auto-deploy the API from `main`.
5. Confirm `/healthz` returns `{"status":"ok"}`.
6. Confirm a basic ConnectRPC read path still works.
7. Check Render logs for startup, storage, and request errors.

Backend CI currently runs:

- `go test ./...`
- `go vet ./...`
- `sqlc generate`
- `buf generate`
- `git diff --exit-code -- internal/postgres/db internal/rpc/runthread`

### Database

Migrations are not automatic yet. For private beta, treat database changes as an explicit release step.

Before merging a schema change:

1. Confirm the migration is additive or otherwise safe for the currently deployed API.
2. Confirm generated sqlc code is committed.
3. Decide whether the migration must run before or after the API deploy.
4. Take a managed Postgres backup or verify a recent backup exists.
5. Run the migration against the beta database.
6. Confirm the API starts and can read/write expected data.

Avoid destructive migrations during private beta unless there is a tested rollback plan.

### Mobile

The mobile release path is:

1. Open a pull request.
2. Wait for Mobile CI to pass.
3. Merge to `main`.
4. Build a beta artifact with the hosted API URL.
5. Upload to TestFlight or the Play Console internal testing track.
6. Smoke test the installed app against the hosted API.

Mobile CI currently runs:

- `flutter pub get`
- `flutter analyze`
- `flutter test`

Use version numbers like:

```text
0.1.0+1
0.1.1+2
0.2.0+3
```

Increment the build number for every uploaded mobile artifact.

## Private Beta Checklist

Before inviting a tester:

- Backend CI is required on `main`.
- Mobile CI is required on `main`.
- Render API deploy from `main` is enabled.
- Render environment variables are configured.
- Managed Postgres is attached.
- Migrations have been applied.
- `/healthz` passes on the hosted API.
- The mobile app is built with the hosted API base URL.
- Logs have been checked after deploy.
- Known demo limitations are documented for testers.

Current private-beta limitations:

- Production auth is not implemented.
- Some request shapes still use demo athlete and goal identifiers.
- Provider OAuth and import flows are still being hardened.
- Migrations are manual.
- Rollback is currently redeploying a previous Render deploy or reverting the PR.

## Rollback

For API regressions:

1. Use Render to redeploy the previous successful deploy, or revert the offending PR.
2. Check `/healthz`.
3. Check logs for startup errors.
4. Verify the mobile app can read the current plan screen.

For mobile regressions:

1. Stop rollout in TestFlight or the Play Console.
2. Promote or reinstall the previous known-good build.
3. Keep the API backward-compatible with the previous mobile build during beta.

For database regressions:

1. Prefer forward fixes when data has already been written.
2. Restore from backup only if the data state is known and acceptable.
3. Record the incident and add a migration test or checklist item before the next release.
