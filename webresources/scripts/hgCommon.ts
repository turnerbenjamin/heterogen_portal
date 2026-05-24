declare const htmx: typeof import("htmx.org").default;

import * as utils from "../common/utils";

class HeterogenCommon {
  constructor() {
    this.initialiseErrorHandling();
    this.initialiseToast();
  }

  /**
   * Add event handlers for error handling. This includes hiding errors before a
   * new request and displaying any errors returned in an error container.
   */
  private initialiseErrorHandling = () => {
    const errorContainerSelector = ".error-container";
    // Hide error container before a new request
    htmx.on("htmx:beforeRequest", (e: Event) => {
      const target = utils.GetEventTarget(e);
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

    // Show and populate error container on error
    htmx.on("htmx:beforeSwap", (e: Event) => {
      const target = utils.GetEventTarget(e);
      if (!target) {
        return;
      }

      const detail = (e as any)?.detail as any;
      if (!detail.isError) {
        return;
      }

      detail.shouldSwap = true;
      const errorContainer = target.querySelector(errorContainerSelector);
      if (!errorContainer) {
        return;
      }

      errorContainer.setAttribute("aria-hidden", "false");
      errorContainer.classList.add("show");
      detail.target = errorContainer;
      detail.swapOverride = "innerHTML";
    });
  };

  /**
   * Add handlers to manage the lifecycle of a toast element from initial render
   * to removal after the toast-hide animation has ended
   */
  private initialiseToast = () => {
    document.addEventListener("animationend", (e) => {
      const target = utils.GetEventTarget(e);
      if (!target) {
        return;
      }

      let toast: Element | undefined = undefined;
      switch (e.animationName) {
        case "toast-show":
          toast = target;
          target.classList.remove("show");
          target.classList.add("visible");
          target.addEventListener("click", () => this.triggerHideToast(toast));
          return;
        case "deplete-toast-progress-bar":
          toast = target.closest(".toast") ?? undefined;
          this.triggerHideToast(toast);
          return;
        case "toast-hide":
          toast = target;
          target.remove();
      }
    });
  };

  /**
   * Update a toast element's classes to trigger the toast-hide animation
   * @param toastEl Toast element to hide
   */
  public triggerHideToast = (toastEl: Element | undefined) => {
    if (!toastEl) {
      return;
    }

    toastEl.classList.remove("show", "visible");
    toastEl.classList.add("hide");
  };
}

utils.attachToHeterogenNamespace({
  HeterogenCommon: new HeterogenCommon(),
});
