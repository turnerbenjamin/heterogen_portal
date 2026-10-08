import * as utils from "./utils";

export function InitialiseToast() {
  document.addEventListener("animationend", (e) => {
    const target = utils.ParseHtmlElement(e.target);
    if (!target) {
      return;
    }

    let toast: Element | undefined = undefined;
    switch (e.animationName) {
      case "toast-show":
        toast = target;
        target.classList.remove("show");
        target.classList.add("visible");
        target.addEventListener("click", () => triggerHideToast(toast));
        return;
      case "deplete-toast-progress-bar":
        toast = target.closest(".toast") ?? undefined;
        triggerHideToast(toast);
        return;
      case "toast-hide":
        toast = target;
        target.remove();
    }
  });
}

/**
 * Update a toast element's classes to trigger the toast-hide animation
 * @param toastEl Toast element to hide
 */
export function triggerHideToast(toastEl: Element | undefined) {
  if (!toastEl) {
    return;
  }

  toastEl.classList.remove("show", "visible");
  toastEl.classList.add("hide");
}
