# Runthread Web

Astro landing page for `runthread.app`.

## Local Development

```sh
cd apps/web
npm install
npm run dev
```

For phone testing on the same Wi-Fi network:

```sh
npm run dev:lan
```

Then open this from your phone:

```text
http://<your-mac-lan-ip>:4321/
```

## Environment

Copy the example file and add the Loops form ID from the Loops form builder:

```sh
cp .env.example .env.local
```

```text
PUBLIC_SITE_URL=https://runthread.app
PUBLIC_LOOPS_FORM_ID=<loops-form-id>
```

The site posts directly to the Loops form endpoint:

```text
https://app.loops.so/api/newsletter-form/<loops-form-id>
```

## Deployment

Deploy this app as a static site on Cloudflare Pages.

Cloudflare Pages settings:

- Production branch: `main`
- Root directory: `apps/web`
- Framework preset: `Astro`
- Install command: `npm ci`
- Build command: `npm run build`
- Build output directory: `dist`
- Node version environment variable: `NODE_VERSION=22`

Cloudflare Pages environment variables:

```text
SKIP_DEPENDENCY_INSTALL=1
NODE_VERSION=22
PUBLIC_SITE_URL=https://runthread.app
PUBLIC_LOOPS_FORM_ID=<loops-form-id>
```

`SKIP_DEPENDENCY_INSTALL=1` is required because this monorepo has a root
`.tool-versions` file for Go, Flutter, sqlc, and buf. Cloudflare Pages detects
that file before entering `apps/web` and otherwise tries to install tools that
the static web build does not need.

Custom domains:

- `runthread.app`
- `www.runthread.app`

Use the DNS records Cloudflare Pages provides in Porkbun. Keep the existing
Loops records for `hey.runthread.app`, and reserve `api.runthread.app` for the
future Go API.

After deployment, test the Cloudflare preview URL first, then the production
domain after DNS verifies.
