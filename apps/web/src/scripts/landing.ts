const revealElements = document.querySelectorAll<HTMLElement>("[data-reveal]");
const reducedMotion = window.matchMedia("(prefers-reduced-motion: reduce)").matches;

if (reducedMotion) {
  revealElements.forEach((element) => element.classList.add("is-visible"));
} else {
  const observer = new IntersectionObserver(
    (entries) => {
      entries.forEach((entry) => {
        if (entry.isIntersecting) {
          entry.target.classList.add("is-visible");
          observer.unobserve(entry.target);
        }
      });
    },
    { rootMargin: "0px 0px -12% 0px", threshold: 0.12 },
  );

  revealElements.forEach((element) => observer.observe(element));
}

const demoStates = [
  {
    heading: "Adaptive week",
    status: "Plan ready",
    label: "Plan",
    title: "Build a week you can actually run.",
    body: "Runthread starts with available days, useful training balance, and the next race or goal.",
    meter: "34%",
  },
  {
    heading: "Run completed",
    status: "Awaiting import",
    label: "Run",
    title: "Complete the session where you already track it.",
    body: "The beta assumes your run happens outside Runthread first, then flows back into the plan.",
    meter: "48%",
  },
  {
    heading: "Activity imported",
    status: "Strava synced",
    label: "Import",
    title: "Bring the activity details back into context.",
    body: "Date, distance, duration, and activity type are enough to begin checking the planned workout.",
    meter: "62%",
  },
  {
    heading: "Workout matched",
    status: "92% confidence",
    label: "Match",
    title: "Link the completed run to the likely workout.",
    body: "Runthread shows the matching confidence before the plan adapts around that completed activity.",
    meter: "80%",
  },
  {
    heading: "Week adapted",
    status: "Change explained",
    label: "Adapt",
    title: "Adjust the next useful thing, not the whole plan.",
    body: "A small long-run change is explained while the rest of the week remains stable.",
    meter: "100%",
  },
];

const demo = document.querySelector<HTMLElement>("[data-product-demo]");

if (demo) {
  const heading = demo.querySelector<HTMLElement>("[data-demo-heading]");
  const status = demo.querySelector<HTMLElement>("[data-demo-status]");
  const label = demo.querySelector<HTMLElement>("[data-demo-label]");
  const title = demo.querySelector<HTMLElement>("[data-demo-title]");
  const body = demo.querySelector<HTMLElement>("[data-demo-body]");
  const meter = demo.querySelector<HTMLElement>("[data-demo-meter]");
  const controls = demo.querySelectorAll<HTMLButtonElement>("[data-demo-step]");
  let activeIndex = 0;
  let timer: number | undefined;

  const setDemoState = (index: number) => {
    const boundedIndex = Math.max(0, Math.min(demoStates.length - 1, index));
    const state = demoStates[boundedIndex];
    activeIndex = boundedIndex;

    if (heading) heading.textContent = state.heading;
    if (status) status.textContent = state.status;
    if (label) label.textContent = state.label;
    if (title) title.textContent = state.title;
    if (body) body.textContent = state.body;
    if (meter) meter.style.width = state.meter;

    controls.forEach((control) => {
      const controlIndex = Number(control.dataset.demoStep ?? "0");
      const isActive = controlIndex === boundedIndex;
      control.classList.toggle("is-active", isActive);
      control.setAttribute("aria-pressed", String(isActive));
    });
  };

  controls.forEach((control) => {
    control.addEventListener("click", () => {
      window.clearInterval(timer);
      setDemoState(Number(control.dataset.demoStep ?? "0"));
    });
  });

  setDemoState(0);

  if (!reducedMotion) {
    timer = window.setInterval(() => {
      setDemoState((activeIndex + 1) % demoStates.length);
    }, 3400);
  }
}

const forms = document.querySelectorAll<HTMLFormElement>("[data-waitlist-form]");

forms.forEach((form) => {
  const status = form.querySelector<HTMLElement>("[data-form-status]");
  const submitButton = form.querySelector<HTMLButtonElement>("button[type='submit']");

  form.addEventListener("submit", async (event) => {
    event.preventDefault();

    const emailInput = form.querySelector<HTMLInputElement>("input[name='email']");
    if (!emailInput?.checkValidity()) {
      emailInput?.reportValidity();
      return;
    }

    const formAction = form.getAttribute("action") ?? "";

    if (!formAction) {
      setStatus(
        status,
        "Beta signup is not configured yet. Add PUBLIC_LOOPS_FORM_ID to enable submissions.",
        "error",
      );
      return;
    }

    const previousButtonText = submitButton?.textContent ?? "";
    submitButton?.setAttribute("disabled", "true");
    if (submitButton) submitButton.textContent = "Joining...";
    setStatus(status, "", "neutral");

    const body = new URLSearchParams();
    new FormData(form).forEach((value, key) => {
      if (typeof value === "string") {
        body.append(key, value);
      }
    });
    body.set(
      "notes",
      `${body.get("notes") ?? "Runthread beta waitlist signup"} at ${new Date().toISOString()}`,
    );

    try {
      const response = await fetch(formAction, {
        method: "POST",
        headers: { "Content-Type": "application/x-www-form-urlencoded" },
        body,
      });

      if (!response.ok) {
        throw new Error(`Loops responded with ${response.status}`);
      }

      form.reset();
      setStatus(status, "You are on the list. We will email beta invites as builds open up.", "success");
    } catch {
      setStatus(status, "Signup did not go through. Please try again in a moment.", "error");
    } finally {
      submitButton?.removeAttribute("disabled");
      if (submitButton) submitButton.textContent = previousButtonText;
    }
  });
});

function setStatus(
  status: HTMLElement | null,
  message: string,
  tone: "success" | "error" | "neutral",
) {
  if (!status) return;

  status.textContent = message;
  if (tone === "error") {
    status.dataset.tone = "error";
  } else {
    delete status.dataset.tone;
  }
}
