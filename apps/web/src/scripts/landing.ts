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
const previewCards = [
  {
    label: "Today",
    title: "Easy run",
    detail: "7.2 km / relaxed",
    note: "Scheduled for recovery after Tuesday's workout.",
    status: "Ready",
  },
  {
    label: "Imported",
    title: "Morning run",
    detail: "6.8 km / 42 min",
    note: "Matched to today's easy run with high confidence.",
    status: "Matched",
  },
  {
    label: "Updated",
    title: "Long run adjusted",
    detail: "15.0 km to 13.5 km",
    note: "Reduced slightly after a partial completion earlier in the week.",
    status: "Adapted",
  },
];

const preview = document.querySelector<HTMLElement>("[data-preview]");

if (preview) {
  const label = preview.querySelector<HTMLElement>("[data-preview-label]");
  const title = preview.querySelector<HTMLElement>("[data-preview-title]");
  const detail = preview.querySelector<HTMLElement>("[data-preview-detail]");
  const note = preview.querySelector<HTMLElement>("[data-preview-note]");
  const status = preview.querySelector<HTMLElement>("[data-preview-status]");
  const controls = preview.querySelectorAll<HTMLButtonElement>("[data-preview-step]");

  controls.forEach((control) => {
    control.addEventListener("click", () => {
      const index = Number(control.dataset.previewStep ?? "0");
      setPreviewCard(index);
    });
  });

  setPreviewCard(0);

  function setPreviewCard(index: number) {
    const boundedIndex = Math.max(0, Math.min(previewCards.length - 1, index));
    const card = previewCards[boundedIndex];

    if (label) label.textContent = card.label;
    if (title) title.textContent = card.title;
    if (detail) detail.textContent = card.detail;
    if (note) note.textContent = card.note;
    if (status) status.textContent = card.status;

    controls.forEach((control) => {
      const controlIndex = Number(control.dataset.previewStep ?? "0");
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
