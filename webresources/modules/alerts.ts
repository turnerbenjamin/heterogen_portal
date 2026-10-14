import constants from "./constants";
import { AppError, NoElWithIdAndTypeErr, parseAppError } from "./errors";
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
  detail: HTMLPreElement;
  confirmBtn: HTMLButtonElement;
  declineBtn: HTMLButtonElement;
}

export class AlertManager {
  public renderNotificationAlert = (title: string, message: string): void => {
    return this.renderAlert({
      type: "notification",
      title: title,
      message: message,
      confirmButtonText: "Ok",
    });
  };

  public renderErrorAlert = (err: AppError): void => {
    return this.renderAlert({
      type: "error",
      title: "Error",
      message: err.message,
      detail: err.details,
      confirmButtonText: "Ok",
    });
  };

  public renderConfirmationAlert = (title: string, message: string): void => {
    return this.renderAlert({
      type: "confirmation",
      title: title,
      message: message,
      confirmButtonText: "Continue",
      declineButtonText: "Cancel",
    });
  };

  public renderServerAlert() {
    this.renderAlert();
  }

  private renderAlert = (
    alert?: NotificationAlert | ErrorAlert | ConfirmationAlert,
  ): void => {
    const alertElementRes = this.getAlertElement();
    if (!alertElementRes.ok) {
      return lastDitchHandler(alertElementRes.error);
    }
    const alertElement = alertElementRes.value;

    alertElement.container.classList.remove(
      constants.layers.alert.modifiers.error,
    );

    if (alert) {
      alertElement.title.textContent = alert.title;
      alertElement.message.textContent = alert.message;
      alertElement.confirmBtn.textContent = alert.confirmButtonText;
    }

    setVisibility(alertElement.detail, alert?.type === "error");
    if (alert?.type === "error") {
      alertElement.detail.textContent = alert.detail;
      alertElement.container.classList.add(
        constants.layers.alert.modifiers.error,
      );
    }

    setVisibility(alertElement.declineBtn, alert?.type === "confirmation");
    if (alert?.type === "confirmation") {
      alertElement.detail.textContent = alert.declineButtonText;
    }

    setVisibility(alertElement.container, true);
  };

  public hideAlert() {
    const alertElementRes = this.getAlertElement();
    if (!alertElementRes.ok) {
      return lastDitchHandler(alertElementRes.error);
    }
    const alertElement = alertElementRes.value;
    setVisibility(alertElement.container, false);
  }

  private getAlertElement = (): Result<AlertElement> => {
    const alertConstants = constants.layers.alert;

    const containerEl = document.getElementById(alertConstants.id);
    if (!containerEl || !(containerEl instanceof HTMLDivElement)) {
      return Fail(new NoElWithIdAndTypeErr(alertConstants.id, "div"));
    }

    const titleEl = document.getElementById(alertConstants.title.id);
    if (!titleEl || !(titleEl instanceof HTMLHeadingElement)) {
      return Fail(new NoElWithIdAndTypeErr(alertConstants.title.id, "heading"));
    }
    HTMLTitleElement;
    const msgEl = document.getElementById(alertConstants.message.id);
    if (!msgEl || !(msgEl instanceof HTMLParagraphElement)) {
      return Fail(
        new NoElWithIdAndTypeErr(alertConstants.message.id, "paragraph"),
      );
    }

    const detailEl = document.getElementById(alertConstants.detail.id);
    if (!detailEl || !(detailEl instanceof HTMLPreElement)) {
      return Fail(new NoElWithIdAndTypeErr(alertConstants.detail.id, "pre"));
    }

    const confirmBtn = document.getElementById(alertConstants.confirmButton.id);
    if (!confirmBtn || !(confirmBtn instanceof HTMLButtonElement)) {
      return Fail(
        new NoElWithIdAndTypeErr(alertConstants.confirmButton.id, "button"),
      );
    }

    const declineBtn = document.getElementById(alertConstants.declineButton.id);
    if (!declineBtn || !(declineBtn instanceof HTMLButtonElement)) {
      throw new NoElWithIdAndTypeErr(alertConstants.declineButton.id, "button");
    }

    return Success({
      container: containerEl,
      title: titleEl,
      message: msgEl,
      detail: detailEl,
      confirmBtn: confirmBtn,
      declineBtn: declineBtn,
    });
  };
}

export function lastDitchHandler(errObj: unknown) {
  const err = parseAppError(errObj);

  const lastDitchErrContainerEl = document.createElement("div");
  lastDitchErrContainerEl.id = "last-ditch-err-container";

  const lastDitchTitleEl = document.createElement("h1");
  lastDitchTitleEl.id = "last-ditch-err-title";
  lastDitchTitleEl.textContent = "Unable to initialise application";

  const lastDitchMsgEl = document.createElement("p");
  lastDitchMsgEl.id = "last-ditch-err-message";
  lastDitchMsgEl.textContent = err.message;

  const lastDitchDetailEl = document.createElement("pre");
  lastDitchDetailEl.id = "last-ditch-err-detail";
  lastDitchDetailEl.classList.add("error-detail");
  lastDitchDetailEl.textContent = err.details;

  lastDitchErrContainerEl.appendChild(lastDitchTitleEl);
  lastDitchErrContainerEl.appendChild(lastDitchMsgEl);
  lastDitchErrContainerEl.appendChild(lastDitchDetailEl);

  document.body.innerHTML = "";
  document.body.appendChild(lastDitchErrContainerEl);
}
