import constants from "./constants";
import { NoElWithIdErr, UnexpectedErr } from "./errors";
import { Fail, Result, Success } from "./result";
import { classSelector, parseHtmlElement } from "./utils";

export interface Toast {
  type: "notification" | "warning" | "error" | "success";
  message: string;
}

export function InitialiseToast() {
  document.addEventListener("animationend", handleToastLifecycle);
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

export function toastMyNotification(msg: string): Error | null {
  return pushToast({ type: "notification", message: msg });
}

export function toastMyWarning(msg: string): Error | null {
  return pushToast({ type: "warning", message: msg });
}

export function toastMyError(msg: string): Error | null {
  return pushToast({ type: "error", message: msg });
}

export function toastMySuccess(msg: string): Error | null {
  return pushToast({ type: "success", message: msg });
}

function pushToast(toast: Toast): Error | null {
  const toastContainer = document.getElementById(constants.layers.toast.id);
  if (!toastContainer) {
    return new NoElWithIdErr(constants.layers.toast.id);
  }

  const typeModifier = getToastTypeModifier(toast);
  if (!typeModifier.ok) {
    return typeModifier.error;
  }

  const toastDiv = document.createElement("div");
  toastDiv.classList.add(
    constants.layers.toast.toastItem.className,
    constants.layers.toast.toastItem.modifiers.show,
    typeModifier.value,
  );

  const messageEl = document.createElement("p");
  messageEl.innerText = toast.message;
  toastDiv.appendChild(messageEl);

  const progressBar = document.createElement("div");
  progressBar.classList.add(
    constants.layers.toast.toastItem.progressBar.className,
  );
  toastDiv.appendChild(progressBar);

  toastContainer.prepend(toastDiv);
  return null;
}

function getToastTypeModifier(toast: Toast): Result<string> {
  switch (toast.type) {
    case "notification":
      return Success(constants.layers.toast.toastItem.modifiers.notification);
    case "warning":
      return Success(constants.layers.toast.toastItem.modifiers.warning);
    case "error":
      return Success(constants.layers.toast.toastItem.modifiers.error);
    default:
      return Fail(
        new UnexpectedErr(
          `unable to map toast type '${toast.type}' to a toast modifier`,
        ),
      );
  }
}

function handleToastLifecycle(e: AnimationEvent) {
  const targetRes = parseHtmlElement(e.target);
  if (!targetRes.ok) {
    return;
  }
  const target = targetRes.value;

  switch (e.animationName) {
    case constants.layers.toast.toastItem.animations.showToast:
      handleShowToastAnimationEnd(target);
      break;

    case constants.layers.toast.toastItem.progressBar.animations.depleteBar:
      handleDepleteToastProgressBarAnimationEnd(target);
      break;

    case constants.layers.toast.toastItem.animations.closeToast:
      handleCloseToastAnimationEnd(target);
      break;
  }
}

function handleShowToastAnimationEnd(toast: HTMLElement) {
  toast.classList.remove(constants.layers.toast.toastItem.modifiers.show);
  toast.classList.add(constants.layers.toast.toastItem.modifiers.visible);
  toast.addEventListener("click", () => triggerHideToast(toast));
  return;
}

function handleDepleteToastProgressBarAnimationEnd(progressBar: HTMLElement) {
  const toast = progressBar.closest(
    classSelector(constants.layers.toast.toastItem.className),
  );
  if (toast) {
    triggerHideToast(toast);
  }
}

function handleCloseToastAnimationEnd(toast: HTMLElement) {
  toast.remove();
}
