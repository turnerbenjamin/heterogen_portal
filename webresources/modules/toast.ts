import constants from "./constants";
import { classSelector, parseHtmlElement } from "./utils";

export function InitialiseToast() {
  document.addEventListener("animationend", (e) => {
    const target = parseHtmlElement(e.target);
    if (!target) {
      return;
    }

    let toast: Element | undefined = undefined;
    switch (e.animationName) {
      case constants.layers.toast.toastItem.animations.showToast:
        toast = target;
        target.classList.remove(
          constants.layers.toast.toastItem.modifiers.show,
        );
        target.classList.add(
          constants.layers.toast.toastItem.modifiers.visible,
        );
        target.addEventListener("click", () => triggerHideToast(toast));
        return;
      case constants.layers.toast.toastItem.progressBar.animations.depleteBar:
        toast =
          target.closest(
            classSelector(constants.layers.toast.toastItem.className),
          ) ?? undefined;
        triggerHideToast(toast);
        return;
      case constants.layers.toast.toastItem.animations.closeToast:
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

  toastEl.classList.remove(
    constants.layers.toast.toastItem.modifiers.show,
    constants.layers.toast.toastItem.modifiers.visible,
  );
  toastEl.classList.add(constants.layers.toast.toastItem.modifiers.hide);
}
