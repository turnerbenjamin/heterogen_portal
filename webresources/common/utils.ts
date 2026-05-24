/**
 * Helper method to extract the target element from an event
 * @param e Event
 * @returns Element or null
 */
export function GetEventTarget(e: Event): Element | null {
  if (!(e.target instanceof Element)) {
    return null;
  }
  return e.target as Element;
}

export function idSelector(id: string) {
  return `#${id}`;
}

export function attachToHeterogenNamespace(data: Record<string, object>): void {
  const g = globalThis as any;
  if (typeof g.HETEROGEN_SCRIPTS !== "object" || g.HETEROGEN_SCRIPTS === null) {
    g.HETEROGEN_SCRIPTS = {};
  }

  for (const [key, value] of Object.entries(data)) {
    g.HETEROGEN_SCRIPTS[key] = value;
  }
}

export function getAppBase(): string {
  const m = document.querySelector(
    'meta[name="app-base-url"]',
  ) as HTMLMetaElement | null;
  return (m && m.content) || window.location.origin;
}

export function handleScriptError(errorObj: unknown) {
  let error: Error;
  if ((errorObj as any).message && (errorObj as any).stack) {
    error = errorObj as Error;
  } else {
    error = new Error(JSON.stringify(errorObj));
  }

  const toastContainer = document.getElementById("toast-container");
  if (!toastContainer) {
    return;
  }

  const toastDiv = document.createElement("div");
  toastDiv.classList.add("toast", "show", "error");

  const errorMessage = document.createElement("p");
  errorMessage.innerText = error.message;

  const progressBar = document.createElement("div");
  progressBar.classList.add("progress-bar");

  toastDiv.appendChild(errorMessage);
  toastDiv.appendChild(progressBar);

  toastContainer.prepend(toastDiv);
}
