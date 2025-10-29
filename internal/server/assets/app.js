document.addEventListener("DOMContentLoaded", () => {
  const highlightLib =
    typeof window !== "undefined" ? (window.hljs ?? null) : null;

  if (highlightLib?.highlightAll) {
    highlightLib.highlightAll();
  }

  const hljsThemeLink = document.getElementById("codetheme");

  const HLJS_THEMES = {
    light: "github.min.css",
    dark: "github-dark.min.css",
  };

  function setHljsTheme(scheme) {
    if (!hljsThemeLink) return;
    const file = HLJS_THEMES[scheme] || HLJS_THEMES.light;
    hljsThemeLink.href = `https://unpkg.com/@highlightjs/cdn-assets@11.11.1/styles/${file}`;
  }

  const form = document.getElementById("query-form");
  const resultContent = document.getElementById("result-content");
  const resultContainer = resultContent.parentElement;
  const resultOverlayHost = resultContainer
    ? wrapResultContainer(resultContainer)
    : null;
  const errorMessage = document.querySelector('[data-role="error"]');
  const copyButton = document.querySelector('[data-role="copy"]');
  const queryTypeField = document.getElementById("query-type");
  const submitButton = document.querySelector('[data-role="submit"]');
  const queryField = document.getElementById("query-text");
  const editor = document.querySelector(".editor");
  const highlightOverlay = editor?.querySelector(".highlight");
  const themeToggle = document.querySelector('[data-role="theme-toggle"]');
  const rootElement = document.documentElement;
  const spinnerOverlay = resultOverlayHost
    ? createSpinnerOverlay(resultOverlayHost)
    : null;
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
      const prefersDark = window.matchMedia?.(
        "(prefers-color-scheme: dark)",
      )?.matches;
      return prefersDark ? "dark" : "light";
    };

    applyScheme(detectInitialScheme());

    themeToggle.addEventListener("click", () => {
      const current =
        rootElement.dataset.colorScheme === "dark" ? "dark" : "light";
      applyScheme(current === "dark" ? "light" : "dark");
    });
  }

  const safeHighlight = (value, language) => {
    if (!highlightLib?.highlight) {
      return null;
    }
    try {
      return highlightLib.highlight(value, { language }).value;
    } catch (error) {
      console.warn("Highlighting failed:", { language, error });
      return null;
    }
  };

  const safeHighlightElement = (element) => {
    if (!highlightLib?.highlightElement || !element) {
      return false;
    }
    try {
      highlightLib.highlightElement(element);
      return true;
    } catch (error) {
      console.warn("Element highlighting failed:", error);
      return false;
    }
  };

  if (editor && highlightOverlay && queryField && queryTypeField) {
    const currentLanguage = () =>
      queryTypeField.value === "cypher" ? "cypher" : "groovy";

    const updateHighlight = (value) => {
      if (!value) {
        highlightOverlay.innerHTML = "";
        return;
      }

      const highlightedValue = safeHighlight(value, currentLanguage());
      if (highlightedValue) {
        highlightOverlay.innerHTML = highlightedValue;
      } else {
        highlightOverlay.textContent = value;
      }
    };

    queryField.addEventListener("input", () =>
      updateHighlight(queryField.value),
    );
    queryField.addEventListener("scroll", () => {
      highlightOverlay.scrollTop = queryField.scrollTop;
      highlightOverlay.scrollLeft = queryField.scrollLeft;
    });

    queryTypeField.addEventListener("change", () => {
      queryField.value = "";
      updateHighlight("");
    });

    updateHighlight(queryField.value);
    rootElement?.setAttribute("data-editor-ready", "true");
  }

  function wrapResultContainer(container) {
    if (!container || !container.parentElement) {
      return null;
    }

    const existingHost = container.parentElement.closest(
      ".result-overlay-host",
    );
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
      <span class="loading-spinner" aria-hidden="true"></span>
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

      var data = await response.json();

      if (!response.ok) {
        throw new Error(data.error || "Request failed");
      }

      const processed = data.processed || "(empty response)";
      resultContent.textContent = processed;
      hideSpinnerOverlay();
      resultContent.removeAttribute("data-highlighted");
      resultContent.classList.remove("hljs");
      if (processed.length <= MAX_HIGHLIGHT_LENGTH) {
        const highlighted = safeHighlightElement(resultContent);
        if (!highlighted) {
          resultContent.textContent = processed;
        }
      } else {
        console.warn(
          `Skipping syntax highlighting (length=${processed.length}, hljs ready=${Boolean(highlightLib)})`,
        );
      }
      subtleScroll(resultContent);
      copyButton.hidden = processed.length === 0;
      copyButton.classList.remove("is-copied");
      copyButton.textContent = "Copy";
    } catch (error) {
      errorMessage.textContent = error.message;
      errorMessage.hidden = false;
      resultContent.textContent = JSON.stringify(data, null, 2);
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
});
