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

const forms = document.querySelectorAll<HTMLFormElement>("[data-waitlist-form]");
const threadStages = [
  {
    label: "Plan",
    title: "32.4 km week built around your rhythm",
    body: "Easy days, threshold work, and the long run land where they can actually happen.",
    metric: "4 runs",
  },
  {
    label: "Run",
    title: "Your real activity becomes the source of truth",
    body: "Keep recording with your normal tracker. Runthread waits for the completed work.",
    metric: "7.2 km",
  },
  {
    label: "Import",
    title: "Strava pulls the session into the loop",
    body: "Provider data stays behind the scenes while Runthread normalises the useful training signal.",
    metric: "Synced",
  },
  {
    label: "Match",
    title: "The run is matched to today, not guessed",
    body: "Date, distance, duration, and workout type combine into a confidence score you can review.",
    metric: "94%",
  },
  {
    label: "Adapt",
    title: "Thursday gets lighter after a partial finish",
    body: "The plan changes conservatively and explains the reason in plain language.",
    metric: "-1.5 km",
  },
];

const threadDemo = document.querySelector<HTMLElement>("[data-thread-demo]");

if (threadDemo) {
  const label = threadDemo.querySelector<HTMLElement>("[data-thread-label]");
  const title = threadDemo.querySelector<HTMLElement>("[data-thread-title]");
  const body = threadDemo.querySelector<HTMLElement>("[data-thread-body]");
  const metric = threadDemo.querySelector<HTMLElement>("[data-thread-metric]");
  const progress = threadDemo.querySelector<SVGPathElement>("[data-route-progress]");
  const controls = threadDemo.querySelectorAll<HTMLButtonElement>(
    "[data-thread-step], [data-thread-tab]",
  );

  controls.forEach((control) => {
    control.addEventListener("click", () => {
      const index = Number(control.dataset.threadStep ?? control.dataset.threadTab ?? "0");
      setThreadStage(index);
    });
  });

  setThreadStage(0);

  function setThreadStage(index: number) {
    const boundedIndex = Math.max(0, Math.min(threadStages.length - 1, index));
    const stage = threadStages[boundedIndex];
    const progressValue = boundedIndex / (threadStages.length - 1);

    if (label) label.textContent = stage.label;
    if (title) title.textContent = stage.title;
    if (body) body.textContent = stage.body;
    if (metric) metric.textContent = stage.metric;
    if (progress) progress.style.strokeDashoffset = String(1 - progressValue);

    controls.forEach((control) => {
      const controlIndex = Number(control.dataset.threadStep ?? control.dataset.threadTab ?? "0");
      const isActive = controlIndex === boundedIndex;
      control.classList.toggle("is-active", isActive);
      control.setAttribute("aria-pressed", String(isActive));
    });
  }
}

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
