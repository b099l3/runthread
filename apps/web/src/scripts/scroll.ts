import { gsap } from "gsap";
import { ScrollTrigger } from "gsap/ScrollTrigger";

gsap.registerPlugin(ScrollTrigger);

/* ------------------------------------------------------------------ */
/* Thread: a stitched path generated from [data-thread-anchor] points  */
/* ------------------------------------------------------------------ */

interface ThreadRefs {
  layer: HTMLElement;
  svg: SVGSVGElement;
  line: SVGPathElement;
  shadow: SVGPathElement;
  needle: HTMLElement;
}

function getThreadRefs(): ThreadRefs | null {
  const layer = document.querySelector<HTMLElement>("[data-thread-layer]");
  const svg = document.querySelector<SVGSVGElement>("[data-thread-svg]");
  const line = document.querySelector<SVGPathElement>("[data-thread-line]");
  const shadow = document.querySelector<SVGPathElement>("[data-thread-shadow]");
  const needle = document.querySelector<HTMLElement>("[data-thread-needle]");
  if (!layer || !svg || !line || !shadow || !needle) return null;
  return { layer, svg, line, shadow, needle };
}

function buildThreadGeometry(refs: ThreadRefs): number {
  const anchors = gsap.utils.toArray<HTMLElement>("[data-thread-anchor]");
  if (anchors.length < 2) return 0;

  const vw = document.documentElement.clientWidth;
  const docHeight = document.documentElement.scrollHeight;
  const narrow = vw < 720;
  const sideX = (side: string, rect: DOMRect): number => {
    switch (side) {
      case "left":
        return vw * (narrow ? 0.2 : 0.1);
      case "right":
        return vw * (narrow ? 0.8 : 0.9);
      case "tie":
        return rect.left + rect.width / 2;
      default:
        return vw * 0.5;
    }
  };

  const scrollY = window.scrollY;
  const points = anchors.map((el) => {
    const mode = el.dataset.threadAnchor ?? "center";
    const rect = el.getBoundingClientRect();
    return {
      x: sideX(mode, rect),
      // The tie-off point sits just above its card so the knot stays visible
      // instead of disappearing behind the card surface.
      y: mode === "tie" ? rect.top + scrollY - 42 : rect.top + scrollY + rect.height * 0.4,
    };
  });
  points.sort((a, b) => a.y - b.y);

  // Weave: when consecutive anchors sit on similar x, swing through the
  // opposite side between them so the thread always crosses the page.
  const woven: { x: number; y: number }[] = [points[0]];
  for (let i = 1; i < points.length; i += 1) {
    const prev = points[i - 1];
    const next = points[i];
    if (Math.abs(prev.x - next.x) < vw * 0.3 && next.y - prev.y > 220) {
      woven.push({
        x: vw - (prev.x + next.x) / 2,
        y: (prev.y + next.y) / 2,
      });
    }
    woven.push(next);
  }

  // Smooth path with vertical tangents at each point.
  const start = { x: woven[0].x, y: Math.max(0, woven[0].y - 260) };
  let d = `M ${start.x.toFixed(1)} ${start.y.toFixed(1)}`;
  let last = start;
  for (const point of woven) {
    const dy = point.y - last.y;
    d += ` C ${last.x.toFixed(1)} ${(last.y + dy * 0.45).toFixed(1)}, ${point.x.toFixed(1)} ${(point.y - dy * 0.45).toFixed(1)}, ${point.x.toFixed(1)} ${point.y.toFixed(1)}`;
    last = point;
  }

  // Tie off: a small loop knotted after the final anchor.
  const r = narrow ? 20 : 30;
  d += ` c ${r * 0.2} ${r * 1.4}, ${r * 1.8} ${r * 1.2}, ${r * 1.6} ${r * 0.1}`;
  d += ` c ${-r * 0.2} ${-r * 1.1}, ${-r * 2.4} ${-r * 0.9}, ${-r * 2.2} ${r * 0.3}`;

  refs.layer.style.height = `${docHeight}px`;
  refs.svg.setAttribute("viewBox", `0 0 ${vw} ${docHeight}`);
  refs.line.setAttribute("d", d);
  refs.shadow.setAttribute("d", d);

  return refs.line.getTotalLength();
}

function setThreadProgress(refs: ThreadRefs, length: number, progress: number) {
  const offset = length * (1 - progress);
  refs.line.style.strokeDasharray = `${length}`;
  refs.line.style.strokeDashoffset = `${offset}`;
  refs.shadow.style.strokeDasharray = `${length}`;
  refs.shadow.style.strokeDashoffset = `${offset}`;

  const head = refs.line.getPointAtLength(length * progress);
  gsap.set(refs.needle, { x: head.x, y: head.y, xPercent: -50, yPercent: -50 });
}

/* ------------------------------------------------------------------ */
/* Hero title word splitting                                           */
/* ------------------------------------------------------------------ */

function splitWords(element: HTMLElement) {
  const wrap = (node: Node, target: DocumentFragment | HTMLElement) => {
    if (node.nodeType === Node.TEXT_NODE) {
      const words = (node.textContent ?? "").split(/(\s+)/);
      for (const word of words) {
        if (!word) continue;
        if (/^\s+$/.test(word)) {
          target.append(document.createTextNode(" "));
          continue;
        }
        const outer = document.createElement("span");
        outer.className = "word";
        const inner = document.createElement("span");
        inner.className = "word-inner";
        inner.textContent = word;
        outer.append(inner);
        target.append(outer);
      }
      return;
    }
    if (node.nodeType === Node.ELEMENT_NODE) {
      const clone = (node as HTMLElement).cloneNode(false) as HTMLElement;
      node.childNodes.forEach((child) => wrap(child, clone));
      target.append(clone);
    }
  };

  const fragment = document.createDocumentFragment();
  [...element.childNodes].forEach((node) => wrap(node, fragment));
  element.replaceChildren(fragment);
}

/* ------------------------------------------------------------------ */
/* Product demo states (shared by clicks and scroll scrubbing)         */
/* ------------------------------------------------------------------ */

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
    body: "Your run happens outside Runthread first, then flows back into the plan on its own.",
    meter: "48%",
  },
  {
    heading: "Activity imported",
    status: "Strava synced",
    label: "Import",
    title: "The details come back into context.",
    body: "Date, distance, duration, and activity type are enough to start checking the planned workout.",
    meter: "62%",
  },
  {
    heading: "Workout matched",
    status: "92% confidence",
    label: "Match",
    title: "The run is linked to the likely workout.",
    body: "Runthread shows the matching confidence before the plan adapts around the completed activity.",
    meter: "80%",
  },
  {
    heading: "Week adapted",
    status: "Change explained",
    label: "Adapt",
    title: "Adjust the next useful thing, not the whole plan.",
    body: "A small long-run change is explained while the rest of the week stays exactly as planned.",
    meter: "100%",
  },
];

function initDemo(): ((index: number) => void) | null {
  const demo = document.querySelector<HTMLElement>("[data-product-demo]");
  if (!demo) return null;

  const heading = demo.querySelector<HTMLElement>("[data-demo-heading]");
  const status = demo.querySelector<HTMLElement>("[data-demo-status]");
  const label = demo.querySelector<HTMLElement>("[data-demo-label]");
  const title = demo.querySelector<HTMLElement>("[data-demo-title]");
  const body = demo.querySelector<HTMLElement>("[data-demo-body]");
  const meter = demo.querySelector<HTMLElement>("[data-demo-meter]");
  const controls = demo.querySelectorAll<HTMLButtonElement>("[data-demo-step]");
  let activeIndex = -1;

  const setDemoState = (index: number) => {
    const boundedIndex = Math.max(0, Math.min(demoStates.length - 1, index));
    if (boundedIndex === activeIndex) return;
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
      setDemoState(Number(control.dataset.demoStep ?? "0"));
    });
  });

  setDemoState(0);
  return setDemoState;
}

/* ------------------------------------------------------------------ */
/* Setup                                                               */
/* ------------------------------------------------------------------ */

const setDemoState = initDemo();
const threadRefs = getThreadRefs();
let threadLength = 0;
let threadProgress = 0;

const mm = gsap.matchMedia();

mm.add("(prefers-reduced-motion: reduce)", () => {
  document
    .querySelectorAll("[data-reveal]")
    .forEach((element) => element.classList.add("is-visible"));

  const ring = document.querySelector<HTMLElement>("[data-ring]");
  const ringCount = document.querySelector<HTMLElement>("[data-ring-count]");
  if (ring && ringCount) {
    const value = Number(ring.dataset.ringValue ?? "92");
    ring.style.setProperty("--ring-fill", String(value));
    ringCount.textContent = String(value);
  }

  if (threadRefs) {
    threadRefs.needle.style.display = "none";
    const draw = () => {
      threadLength = buildThreadGeometry(threadRefs);
      if (threadLength > 0) setThreadProgress(threadRefs, threadLength, 1);
    };
    window.addEventListener("load", draw);
    draw();
  }
});

mm.add(
  {
    motion: "(prefers-reduced-motion: no-preference)",
    desktop: "(min-width: 1021px)",
    mobile: "(max-width: 1020px)",
  },
  (context) => {
    const { motion, desktop } = context.conditions as {
      motion: boolean;
      desktop: boolean;
    };
    if (!motion) return;

    /* Generic reveals (works on every page) */
    const revealElements = gsap.utils.toArray<HTMLElement>("[data-reveal]");
    if (revealElements.length > 0) {
      ScrollTrigger.batch(revealElements, {
        start: "top 88%",
        once: true,
        onEnter: (batch) => {
          batch.forEach((element, index) => {
            window.setTimeout(() => element.classList.add("is-visible"), index * 90);
          });
        },
      });
    }

    /* Header compression */
    const header = document.querySelector<HTMLElement>("[data-site-header]");
    if (header) {
      ScrollTrigger.create({
        start: 90,
        end: "max",
        onToggle: (self) => header.classList.toggle("is-compressed", self.isActive),
      });
    }

    /* Hero entrance + parallax */
    const heroTitle = document.querySelector<HTMLElement>("[data-hero-title]");
    const heroProduct = document.querySelector<HTMLElement>("[data-hero-product]");
    if (heroTitle) {
      splitWords(heroTitle);
      const intro = gsap.timeline({ defaults: { ease: "power3.out" } });
      intro
        .from("[data-hero-eyebrow]", { y: 18, autoAlpha: 0, duration: 0.5 }, 0.1)
        .from(
          heroTitle.querySelectorAll(".word-inner"),
          { yPercent: 112, duration: 0.9, stagger: 0.07 },
          0.18,
        )
        .from("[data-hero-lede]", { y: 26, autoAlpha: 0, duration: 0.7 }, 0.62)
        .from("[data-hero-form]", { y: 26, autoAlpha: 0, duration: 0.7 }, 0.74);
      if (heroProduct) {
        intro.from(
          heroProduct,
          { y: 70, rotate: 2.5, autoAlpha: 0, duration: 1.1, ease: "power2.out" },
          0.45,
        );
      }
    }
    if (heroProduct) {
      gsap.to(heroProduct, {
        y: desktop ? 90 : 40,
        ease: "none",
        scrollTrigger: {
          trigger: ".hero",
          start: "top top",
          end: "bottom top",
          scrub: 0.6,
        },
      });
    }

    /* Scroll-driven product demo states while the hero is in view */
    if (setDemoState) {
      ScrollTrigger.create({
        trigger: ".hero",
        start: "top top",
        end: "bottom 25%",
        onUpdate: (self) => {
          setDemoState(Math.round(self.progress * (demoStates.length - 1)));
        },
      });
    }

    /* Thread statement: words brighten as you scroll through */
    const statement = document.querySelector<HTMLElement>("[data-statement]");
    if (statement) {
      splitWords(statement);
      gsap.fromTo(
        statement.querySelectorAll(".word-inner"),
        { opacity: 0.16 },
        {
          opacity: 1,
          stagger: 0.12,
          ease: "none",
          scrollTrigger: {
            trigger: statement,
            start: "top 82%",
            end: "bottom 45%",
            scrub: 0.4,
          },
        },
      );
    }

    /* Marquee: endless loop, speed nudged by scroll velocity */
    const marqueeTrack = document.querySelector<HTMLElement>("[data-marquee-track]");
    if (marqueeTrack) {
      const loop = gsap.to(marqueeTrack, {
        xPercent: -50,
        ease: "none",
        duration: 26,
        repeat: -1,
      });
      ScrollTrigger.create({
        trigger: "[data-marquee]",
        start: "top bottom",
        end: "bottom top",
        onUpdate: (self) => {
          const velocity = gsap.utils.clamp(-2400, 2400, self.getVelocity());
          const boost = 1 + Math.abs(velocity) / 900;
          gsap.to(loop, { timeScale: boost, duration: 0.3, overwrite: true });
          gsap.to(loop, { timeScale: 1, duration: 1.2, delay: 0.35 });
        },
      });
    }

    /* How it works: pinned scroll-telling on desktop, reveals on mobile */
    const flowCards = gsap.utils.toArray<HTMLElement>("[data-flow-card]");
    if (flowCards.length > 0) {
      if (desktop) {
        gsap.set(flowCards, { opacity: 0.32, y: 26, scale: 0.985 });
        const steps = gsap.timeline({
          scrollTrigger: {
            trigger: "[data-hiw]",
            start: "top top+=70",
            end: "+=180%",
            pin: true,
            scrub: 0.5,
          },
        });
        flowCards.forEach((card, index) => {
          steps.to(
            card,
            { opacity: 1, y: 0, scale: 1, duration: 0.55, ease: "power2.out" },
            index * 0.75,
          );
          if (index > 0) {
            steps.to(
              flowCards[index - 1],
              { opacity: 0.55, scale: 0.99, duration: 0.4, ease: "power1.out" },
              index * 0.75 + 0.1,
            );
          }
        });
        steps.to({}, { duration: 0.5 });
      } else {
        flowCards.forEach((card) => {
          gsap.from(card, {
            y: 34,
            autoAlpha: 0,
            duration: 0.7,
            ease: "power2.out",
            scrollTrigger: { trigger: card, start: "top 88%", once: true },
          });
        });
      }
    }

    /* Feature modules: staggered rise */
    const modules = gsap.utils.toArray<HTMLElement>(".feature-module");
    if (modules.length > 0) {
      gsap.set(modules, { y: 44, autoAlpha: 0 });
      ScrollTrigger.batch(modules, {
        start: "top 86%",
        once: true,
        onEnter: (batch) =>
          gsap.to(batch, {
            y: 0,
            autoAlpha: 1,
            duration: 0.8,
            stagger: 0.09,
            ease: "power3.out",
          }),
      });
    }

    /* Detail panels: gentle alternating parallax */
    gsap.utils.toArray<HTMLElement>("[data-parallax]").forEach((panel) => {
      const factor = Number(panel.dataset.parallax ?? "0");
      gsap.to(panel, {
        y: factor * (desktop ? 900 : 380),
        ease: "none",
        scrollTrigger: {
          trigger: panel,
          start: "top bottom",
          end: "bottom top",
          scrub: 0.7,
        },
      });
    });

    /* Confidence ring count-up */
    const ring = document.querySelector<HTMLElement>("[data-ring]");
    const ringCount = document.querySelector<HTMLElement>("[data-ring-count]");
    if (ring && ringCount) {
      const value = Number(ring.dataset.ringValue ?? "92");
      const counter = { current: 0 };
      ring.style.setProperty("--ring-fill", "0");
      ScrollTrigger.create({
        trigger: ring,
        start: "top 82%",
        once: true,
        onEnter: () =>
          gsap.to(counter, {
            current: value,
            duration: 1.4,
            ease: "power2.out",
            onUpdate: () => {
              ringCount.textContent = String(Math.round(counter.current));
              ring.style.setProperty("--ring-fill", String(counter.current));
            },
          }),
      });
    }

    /* Timeline rows stitch in one by one */
    const timelineRows = gsap.utils.toArray<HTMLElement>("[data-timeline] span");
    if (timelineRows.length > 0) {
      gsap.from(timelineRows, {
        x: -22,
        autoAlpha: 0,
        duration: 0.55,
        stagger: 0.14,
        ease: "power2.out",
        scrollTrigger: { trigger: "[data-timeline]", start: "top 82%", once: true },
      });
    }

    /* Dark band: scales/rounds in as it enters */
    const band = document.querySelector<HTMLElement>("[data-band]");
    if (band) {
      gsap.fromTo(
        band,
        { scale: 0.965, borderRadius: "26px" },
        {
          scale: 1,
          borderRadius: "0px",
          ease: "none",
          scrollTrigger: {
            trigger: band,
            start: "top 92%",
            end: "top 38%",
            scrub: 0.5,
          },
        },
      );
    }

    /* Thread drawing, tied to overall page progress */
    if (threadRefs) {
      threadRefs.needle.style.display = "";
      ScrollTrigger.create({
        start: 0,
        end: "max",
        onUpdate: (self) => {
          threadProgress = gsap.utils.clamp(0, 1, self.progress * 1.06);
          if (threadLength > 0) {
            setThreadProgress(threadRefs, threadLength, threadProgress);
          }
        },
      });
    }
  },
);

/* Rebuild thread geometry whenever ScrollTrigger recalculates layout */
if (threadRefs) {
  const rebuild = () => {
    threadLength = buildThreadGeometry(threadRefs);
    if (threadLength > 0) {
      const reduced = window.matchMedia("(prefers-reduced-motion: reduce)").matches;
      setThreadProgress(threadRefs, threadLength, reduced ? 1 : threadProgress);
    }
  };
  ScrollTrigger.addEventListener("refresh", rebuild);
  rebuild();
}

window.addEventListener("load", () => ScrollTrigger.refresh());
document.fonts?.ready.then(() => ScrollTrigger.refresh());
