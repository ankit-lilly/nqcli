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
const errorMessage = document.querySelector('[data-role="error"]');
const copyButton = document.querySelector('[data-role="copy"]');
const queryTypeField = document.getElementById("query-type");
const submitButton = document.querySelector('[data-role="submit"]');
const queryField = document.getElementById("query-text");
const editor = document.querySelector(".editor");
const highlightOverlay = editor?.querySelector(".highlight");
const themeToggle = document.querySelector('[data-role="theme-toggle"]');
const rootElement = document.documentElement;

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
    resultContent.removeAttribute("data-highlighted");
    resultContent.classList.remove("hljs");
    hljs.highlightElement(resultContent);
    triggerResultFeedback();
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

function triggerResultFeedback() {
  resultContainer.classList.remove("result-highlight");
  // Force reflow so animation can restart ( https://stackoverflow.com/questions/60686489/what-purpose-does-void-element-offsetwidth-serve )
  void resultContainer.offsetWidth;
  resultContainer.classList.add("result-highlight");
  subtleScroll(resultContent);
}

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

resultContainer.addEventListener("animationend", () => {
  resultContainer.classList.remove("result-highlight");
});
