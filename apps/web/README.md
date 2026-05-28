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

Deploy this app as a static site on Vercel with:

- Root directory: `apps/web`
- Build command: `npm run build`
- Output directory: `dist`

Keep the Go API on Render and use a separate API hostname later, such as
`api.runthread.app`.
