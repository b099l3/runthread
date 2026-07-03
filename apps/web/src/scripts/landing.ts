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
