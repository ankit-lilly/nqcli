hljs.highlightAll();

const hljsThemeLink = document.getElementById("codetheme");

const HLJS_THEMES = {
  light: "github.min.css",
  dark:  "github-dark.min.css",
};

function setHljsTheme(scheme) {
  if (!hljsThemeLink) return;
  const file = HLJS_THEMES[scheme] || HLJS_THEMES.light;
  hljsThemeLink.href = `https://unpkg.com/@highlightjs/cdn-assets@11.11.1/styles/${file}`;
}


const form = document.getElementById("query-form");
const resultContent = document.getElementById("result-content");
const resultContainer = resultContent.parentElement;
const resultOverlayHost = resultContainer ? wrapResultContainer(resultContainer) : null;
const errorMessage = document.querySelector('[data-role="error"]');
const copyButton = document.querySelector('[data-role="copy"]');
const queryTypeField = document.getElementById("query-type");
const submitButton = document.querySelector('[data-role="submit"]');
const queryField = document.getElementById("query-text");
const editor = document.querySelector(".editor");
const highlightOverlay = editor?.querySelector(".highlight");
const themeToggle = document.querySelector('[data-role="theme-toggle"]');
const rootElement = document.documentElement;
const spinnerOverlay = resultOverlayHost ? createSpinnerOverlay(resultOverlayHost) : null;
const MAX_HIGHLIGHT_LENGTH = 25000;

if (themeToggle) {
  const schemes = new Set(["light", "dark"]);
  const storageKey = "nqcli-color-scheme";

  const readStoredScheme = () => {
    try {
      return window.localStorage?.getItem(storageKey) ?? null;
    } catch (error) {
      console.warn("Color scheme storage read failed:", error);
      return null;
    }
  };

  const writeStoredScheme = (scheme) => {
    try {
      window.localStorage?.setItem(storageKey, scheme);
    } catch (error) {
      console.warn("Color scheme storage write failed:", error);
    }
  };

  const applyScheme = (scheme) => {
    const nextScheme = schemes.has(scheme) ? scheme : "light";
    rootElement.dataset.colorScheme = nextScheme;
    themeToggle.setAttribute("aria-pressed", nextScheme === "dark");
    writeStoredScheme(nextScheme);
    setHljsTheme(nextScheme);
  };

  const detectInitialScheme = () => {
    const stored = readStoredScheme();
    if (schemes.has(stored)) {
      return stored;
    }
    const prefersDark = window.matchMedia?.("(prefers-color-scheme: dark)")?.matches;
    return prefersDark
      ? "dark"
      : "light";
  };

  applyScheme(detectInitialScheme());

  themeToggle.addEventListener("click", () => {
    const current = rootElement.dataset.colorScheme === "dark" ? "dark" : "light";
    applyScheme(current === "dark" ? "light" : "dark");
  });
}

if (editor && highlightOverlay && queryField && queryTypeField) {
  const currentLanguage = () =>
    queryTypeField.value === "cypher" ? "cypher" : "groovy";

  const updateHighlight = (value) => {
    if (!value) {
      highlightOverlay.innerHTML = "";
      return;
    }

    highlightOverlay.innerHTML = hljs.highlight(value, {
      language: currentLanguage(),
    }).value;
  };

  queryField.addEventListener("input", () => updateHighlight(queryField.value));
  queryField.addEventListener("scroll", () => {
    highlightOverlay.scrollTop = queryField.scrollTop;
    highlightOverlay.scrollLeft = queryField.scrollLeft;
  });

  queryTypeField.addEventListener("change", () => {
    queryField.value = "";
    updateHighlight("");
  });

  updateHighlight(queryField.value);
}

function wrapResultContainer(container) {
  if (!container || !container.parentElement) {
    return null;
  }

  const existingHost = container.parentElement.closest(".result-overlay-host");
  if (existingHost && existingHost.contains(container)) {
    return existingHost;
  }

  const host = document.createElement("div");
  host.className = "result-overlay-host";
  container.parentElement.insertBefore(host, container);
  host.appendChild(container);
  return host;
}

function createSpinnerOverlay(hostElement) {
  if (!hostElement) {
    return null;
  }

  hostElement.classList.add("result-overlay-host");

  const overlay = document.createElement("div");
  overlay.className = "spinner-overlay";
  overlay.innerHTML = `
    <div class="spinner-overlay__content" role="status" aria-live="polite" aria-label="Loading">
      <svg xmlns="http://www.w3.org/2000/svg" width="48" height="48" viewBox="0 0 24 24">
        <path fill="currentColor" d="M20.27,4.74a4.93,4.93,0,0,1,1.52,4.61,5.32,5.32,0,0,1-4.1,4.51,5.12,5.12,0,0,1-5.2-1.5,5.53,5.53,0,0,0,6.13-1.48A5.66,5.66,0,0,0,20.27,4.74ZM12.32,11.53a5.49,5.49,0,0,0-1.47-6.2A5.57,5.57,0,0,0,4.71,3.72,5.17,5.17,0,0,1,9.53,2.2,5.52,5.52,0,0,1,13.9,6.45,5.28,5.28,0,0,1,12.32,11.53ZM19.2,20.29a4.92,4.92,0,0,1-4.72,1.49,5.32,5.32,0,0,1-4.34-4.05A5.2,5.2,0,0,1,11.6,12.5a5.6,5.6,0,0,0,1.51,6.13A5.63,5.63,0,0,0,19.2,20.29ZM3.79,19.38A5.18,5.18,0,0,1,2.32,14a5.3,5.3,0,0,1,4.59-4,5,5,0,0,1,4.58,1.61,5.55,5.55,0,0,0-6.32,1.69A5.46,5.46,0,0,0,3.79,19.38ZM12.23,12a5.11,5.11,0,0,0,3.66-5,5.75,5.75,0,0,0-3.18-6,5,5,0,0,1,4.42,2.3,5.21,5.21,0,0,1,.24,5.92A5.4,5.4,0,0,1,12.23,12ZM11.76,12a5.18,5.18,0,0,0-3.68,5.09,5.58,5.58,0,0,0,3.19,5.79c-1,.35-2.9-.46-4-1.68A5.51,5.51,0,0,1,11.76,12ZM23,12.63a5.07,5.07,0,0,1-2.35,4.52,5.23,5.23,0,0,1-5.91.2,5.24,5.24,0,0,1-2.67-4.77,5.51,5.51,0,0,0,5.45,3.33A5.52,5.52,0,0,0,23,12.63ZM1,11.23a5,5,0,0,1,2.49-4.5,5.23,5.23,0,0,1,5.81-.06,5.3,5.3,0,0,1,2.61,4.74A5.56,5.56,0,0,0,6.56,8.06,5.71,5.71,0,0,0,1,11.23Z">
          <animateTransform attributeName="transform" dur="1.5s" repeatCount="indefinite" type="rotate" values="0 12 12;360 12 12"/>
        </path>
      </svg>
    </div>
  `;
  hostElement.appendChild(overlay);
  return overlay;
}

function showSpinnerOverlay() {
  spinnerOverlay?.classList.add("is-active");
}

function hideSpinnerOverlay() {
  spinnerOverlay?.classList.remove("is-active");
}

form.addEventListener("submit", async (event) => {
  event.preventDefault();

  const payload = {
    type: queryTypeField.value,
    query: queryField.value,
  };

  console.log("Submitting payload:", payload);

  errorMessage.hidden = true;
  submitButton.textContent = "Running...";
  submitButton.disabled = true;
  submitButton.setAttribute("aria-busy", "true");
  resultContent.setAttribute("aria-busy", "true");
  showSpinnerOverlay();
  flashButton(submitButton, "btn-flash");

  try {
    const response = await fetch("/queries", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(payload),
    });

    const data = await response.json();

    if (!response.ok) {
      throw new Error(data.error || "Request failed");
    }

    const processed = data.processed || "(empty response)";
    resultContent.textContent = processed;
    hideSpinnerOverlay();
    resultContent.removeAttribute("data-highlighted");
    resultContent.classList.remove("hljs");
    if (processed.length <= MAX_HIGHLIGHT_LENGTH) {
      hljs.highlightElement(resultContent);
    } else {
      console.warn(`Skipping syntax highlighting for response length ${processed.length}`);
    }
    subtleScroll(resultContent);
    copyButton.hidden = processed.length === 0;
    copyButton.classList.remove("is-copied");
    copyButton.textContent = "Copy";
  } catch (error) {
    errorMessage.textContent = error.message;
    errorMessage.hidden = false;
  } finally {
    submitButton.disabled = false;
    submitButton.textContent = "Run";
    submitButton.removeAttribute("aria-busy");
    resultContent.removeAttribute("aria-busy");
    hideSpinnerOverlay();
  }
});

copyButton.addEventListener("click", async () => {
  const textToCopy = resultContent.textContent;
  if (!textToCopy) {
    return;
  }

  try {
    await navigator.clipboard.writeText(textToCopy);
    flashButton(copyButton, "btn-flash--copy");
    copyButton.classList.add("is-copied");
    copyButton.textContent = "Copied!";
    subtleScroll(resultContent);
    setTimeout(() => {
      copyButton.classList.remove("is-copied");
      copyButton.textContent = "Copy";
    }, 1500);
  } catch (error) {
    errorMessage.textContent = "Failed to copy to clipboard.";
    errorMessage.hidden = false;
  }
});

function subtleScroll(element) {
  //scrollTop: Current vertical scroll position of the element.
  //clientHeight: Visible height of the element (viewport).
  //scrollHeight: Total height of the content, including the part not visible.
  const initial = 0;
  const available = element.scrollHeight - element.clientHeight;
  if (available <= 0) {
    return;
  }
  const delta = Math.min(available, 70);
  smoothScrollTo(element, initial + delta);
  setTimeout(() => {
    smoothScrollTo(element, initial);
  }, 250);
}

function smoothScrollTo(element, top) {
  if (typeof element.scrollTo === "function") {
    try {
      element.scrollTo({ top, behavior: "smooth" });
      return;
    } catch (err) {
      // ignore and fall back
    }
  }
  element.scrollTop = top;
}

function flashButton(button, animationClass) {
  button.classList.remove(animationClass);
  // Force reflow so animation restarts
  void button.offsetWidth;
  button.classList.add(animationClass);
  button.addEventListener(
    "animationend",
   () => button.classList.remove(animationClass),
   { once: true },
 );
}
