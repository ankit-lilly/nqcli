hljs.highlightAll();

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
  // Force reflow so animation can restart
  void resultContainer.offsetWidth;
  resultContainer.classList.add("result-highlight");
  subtleScroll(resultContent);
}

function subtleScroll(element) {
  const initial = element.scrollTop;
  const available = element.scrollHeight - element.clientHeight;
  if (available <= 0) {
    return;
  }
  const delta = Math.min(available, 40);
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
