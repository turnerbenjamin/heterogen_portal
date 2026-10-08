import { HtmxBeforeSwapDetails } from "htmx.org";
import { onHtmx } from "./htmx";
import * as utils from "./utils";

export function InitialiseErrorHandling() {
  const errorContainerSelector = ".error-container";

  addEventListener("error", (e: ErrorEvent) => {
    e.preventDefault();
    handle(e.error);
  });

  // Hide error container before a new request
  onHtmx("htmx:beforeRequest", (e) => {
    const target = utils.ParseHtmlElement(e.target);
    if (!target) {
      return;
    }

    const errorContainer = target.querySelector(errorContainerSelector);
    if (!errorContainer) {
      return;
    }

    errorContainer.classList.remove("show");
    errorContainer.setAttribute("aria-hidden", "true");
  });

  onHtmx<HtmxBeforeSwapDetails>("htmx:beforeSwap", (e) => {
    const target = utils.ParseHtmlElement(e.target);
    if (!target) {
      return;
    }

    if (!e.detail.isError) {
      return;
    }

    e.detail.shouldSwap = true;
    const errorContainer = target.querySelector(errorContainerSelector);
    if (!errorContainer) {
      return;
    }

    errorContainer.setAttribute("aria-hidden", "false");
    errorContainer.classList.add("show");
    e.detail.target = errorContainer;
    e.detail.swapOverride = "innerHTML";
  });
}

export function handle(errorObj: unknown) {
  let error: Error;
  if ((errorObj as Error).message && (errorObj as Error).stack) {
    error = errorObj as Error;
  } else {
    if (errorObj === undefined || errorObj === null) {
      error = new Error("An unexpected error has occurred");
    } else {
      error = new Error(JSON.stringify(errorObj));
    }
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

export function Unexpected(msg: string): Error {
  return new Error(`Unexpected error: ${msg}`);
}

export function NullErr(elName: string): Error {
  return Unexpected(`unable to access: ${elName}`);
}

export function UndefinedDataAttribute(attrName: string): Error {
  return Unexpected(`expected data attribute to be set: ${attrName}`);
}
