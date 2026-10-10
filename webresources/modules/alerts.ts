import constants from "./constants";
import { AppError, NoElWithIdAndTypeErr } from "./errors";
import { Fail, Result, Success } from "./result";
import { setVisibility } from "./utils";

interface Alert {
  type: "notification" | "error" | "confirmation";
  title: string;
  message: string;
  confirmButtonText: string;
}

interface NotificationAlert extends Alert {
  type: "notification";
  title: string;
  message: string;
  confirmButtonText: "Ok";
}

interface ErrorAlert extends Alert {
  type: "error";
  title: string;
  message: string;
  detail: string;
  confirmButtonText: "Ok";
}

interface ConfirmationAlert extends Alert {
  type: "confirmation";
  title: string;
  message: string;
  confirmButtonText: "Continue";
  declineButtonText: "Cancel";
}

interface AlertElement {
  container: HTMLDivElement;
  title: HTMLHeadingElement;
  message: HTMLParagraphElement;
  detail: HTMLDivElement;
  confirmBtn: HTMLButtonElement;
  declineBtn: HTMLButtonElement;
}

let alertElement: AlertElement | null = null;

export function renderNotificationAlert(
  title: string,
  message: string,
): Error | null {
  return renderAlert({
    type: "notification",
    title: title,
    message: message,
    confirmButtonText: "Ok",
  });
}

export function renderErrorAlert(err: AppError): Error | null {
  return renderAlert({
    type: "error",
    title: "Error",
    message: err.message,
    detail: err.details,
    confirmButtonText: "Ok",
  });
}

export function renderConfirmationAlert(
  title: string,
  message: string,
): Error | null {
  return renderAlert({
    type: "confirmation",
    title: title,
    message: message,
    confirmButtonText: "Continue",
    declineButtonText: "Cancel",
  });
}

function renderAlert(
  alert: NotificationAlert | ErrorAlert | ConfirmationAlert,
): Error | null {
  if (!alertElement) {
    const getAlertElementResult = getAlertElement();
    if (!getAlertElementResult.ok) {
      return getAlertElementResult.error;
    }
    alertElement = getAlertElementResult.value;
  }

  alertElement.title.textContent = alert.title;
  alertElement.message.textContent = alert.message;
  alertElement.confirmBtn.textContent = alert.confirmButtonText;

  setVisibility(alertElement.detail, alert.type === "error");
  if (alert.type === "error") {
    alertElement.detail.textContent = alert.detail;
  }

  setVisibility(alertElement.declineBtn, alert.type === "confirmation");
  if (alert.type === "confirmation") {
    alertElement.detail.textContent = alert.declineButtonText;
  }

  setVisibility(alertElement.container, true);

  return null;
}

function getAlertElement(): Result<AlertElement> {
  const alertConstants = constants.layers.alert;

  const containerEl = document.getElementById(alertConstants.dialog.id);
  if (!containerEl || !(containerEl instanceof HTMLDivElement)) {
    return Fail(new NoElWithIdAndTypeErr(alertConstants.dialog.id, "div"));
  }

  const titleEl = document.getElementById(alertConstants.dialog.title.id);
  if (!titleEl || !(titleEl instanceof HTMLHeadingElement)) {
    return Fail(
      new NoElWithIdAndTypeErr(alertConstants.dialog.title.id, "heading"),
    );
  }
  HTMLTitleElement;
  const msgEl = document.getElementById(alertConstants.dialog.message.id);
  if (!msgEl || !(msgEl instanceof HTMLParagraphElement)) {
    return Fail(
      new NoElWithIdAndTypeErr(alertConstants.dialog.message.id, "paragraph"),
    );
  }

  const detailEl = document.getElementById(alertConstants.dialog.detail.id);
  if (!detailEl || !(detailEl instanceof HTMLDivElement)) {
    return Fail(
      new NoElWithIdAndTypeErr(alertConstants.dialog.detail.id, "div"),
    );
  }

  const confirmBtn = document.getElementById(
    alertConstants.dialog.confirmButton.id,
  );
  if (!confirmBtn || !(confirmBtn instanceof HTMLButtonElement)) {
    return Fail(
      new NoElWithIdAndTypeErr(
        alertConstants.dialog.confirmButton.id,
        "button",
      ),
    );
  }

  const declineBtn = document.getElementById(
    alertConstants.dialog.declineButton.id,
  );
  if (!declineBtn || !(declineBtn instanceof HTMLButtonElement)) {
    throw new NoElWithIdAndTypeErr(
      alertConstants.dialog.declineButton.id,
      "button",
    );
  }

  return Success({
    container: containerEl,
    title: titleEl,
    message: msgEl,
    detail: detailEl,
    confirmBtn: confirmBtn,
    declineBtn: declineBtn,
  });
}
