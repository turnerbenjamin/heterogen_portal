import constants from "./constants";
import { NoElWithIdAndTypeErr, parseAppError } from "./errors";
import { Fail, Result, Success } from "./result";
import { toastMyError } from "./toasts";
import {
  getChildBySelector,
  getElementById,
  getTemplateById,
  getTemplateContentBySelector,
  idSelector,
  setVisibility,
} from "./utils";
import { Hook, WindowManager } from "./windowManager";
import { onHtmxAfterSwapEvent } from "./htmx";

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
  container: HTMLElement;
  title: HTMLHeadingElement;
  message: HTMLParagraphElement;
  detail: HTMLPreElement;
  confirmBtn: HTMLButtonElement;
  declineBtn: HTMLButtonElement;
}

export class AlertManager {
  private readonly _windowManager: WindowManager;
  private readonly _panic: (errObj: unknown) => void;

  public constructor(
    windowManager: WindowManager,
    panic: (errObj: unknown) => void,
  ) {
    this._windowManager = windowManager;
    this._panic = panic;

    onHtmxAfterSwapEvent(this.handleAfterHtmxSwap);
    windowManager.registerHook(Hook.escAlert, this.hgHandlerCloseAlert);
  }

  public hgHandlerCloseAlert = () => {
    this._windowManager.popAlert();
    this.hideAlert();
  };

  public renderNotificationMsg = (title: string, message: string): void => {
    this.renderAlert({
      type: "notification",
      title: title,
      message: message,
      confirmButtonText: "Ok",
    });
  };

  public renderError = (errObj: unknown): void => {
    const err = parseAppError(errObj);
    if (err.doBlock) {
      this.renderAlert({
        type: "error",
        title: "Error",
        message: err.message,
        detail: err.details,
        confirmButtonText: "Ok",
      });
    } else {
      const toastErr = toastMyError(err.message);
      if (toastErr) {
        this._panic(toastErr);
      }
    }
  };

  public renderConfirmationDialog = (title: string, message: string): void => {
    this.renderAlert({
      type: "confirmation",
      title: title,
      message: message,
      confirmButtonText: "Continue",
      declineButtonText: "Cancel",
    });
  };

  private renderAlert = (
    alert?: NotificationAlert | ErrorAlert | ConfirmationAlert,
  ): void => {
    const alertElementRes = this.getAlertElement();
    if (!alertElementRes.ok) {
      return this._panic(alertElementRes.error);
    }
    const alertElement = alertElementRes.value;
    alertElement.container.classList.remove(constants.alert.modifiers.error);

    if (alert) {
      alertElement.title.textContent = alert.title;
      alertElement.message.textContent = alert.message;
      alertElement.confirmBtn.textContent = alert.confirmButtonText;
    }

    setVisibility(alertElement.detail, alert?.type === "error");
    if (alert?.type === "error") {
      alertElement.detail.textContent = alert.detail;
      alertElement.container.classList.add(constants.alert.modifiers.error);
    }

    setVisibility(alertElement.declineBtn, alert?.type === "confirmation");
    if (alert?.type === "confirmation") {
      alertElement.detail.textContent = alert.declineButtonText;
    }

    this._windowManager.pushAlertLayer();

    const announcement = `${alertElement.title.textContent}: ${alertElement.message.textContent}`;
    this._windowManager.announceAssertively(announcement);

    alertElement.confirmBtn.focus();
  };

  public hideAlert() {
    const alertElementRes = getElementById(constants.alert.id);
    if (!alertElementRes.ok) {
      return this._panic(alertElementRes.error);
    }
    alertElementRes.value.remove();
  }

  private getAlertElement = (): Result<AlertElement> => {
    const alertLayerRes = getElementById(constants.layers.alert.id);
    if (!alertLayerRes.ok) {
      return alertLayerRes;
    }

    const alertTemplateRes = getTemplateById(constants.alert.templateId);
    if (!alertTemplateRes.ok) {
      return alertTemplateRes;
    }

    const alertPrototypeRes = getTemplateContentBySelector(
      alertTemplateRes.value,
      idSelector(constants.alert.id),
    );
    if (!alertPrototypeRes.ok) {
      return alertPrototypeRes;
    }

    const alertEl = alertPrototypeRes.value.cloneNode(true) as HTMLElement;
    alertLayerRes.value.appendChild(alertEl);

    const titleEl = getChildBySelector(
      alertEl,
      idSelector(constants.alert.title.id),
    );
    if (!titleEl.ok || !(titleEl.value instanceof HTMLHeadingElement)) {
      return Fail(
        new NoElWithIdAndTypeErr(constants.alert.title.id, "heading"),
      );
    }

    const msgEl = getChildBySelector(
      alertEl,
      idSelector(constants.alert.message.id),
    );
    if (!msgEl.ok || !(msgEl.value instanceof HTMLParagraphElement)) {
      return Fail(
        new NoElWithIdAndTypeErr(constants.alert.message.id, "paragraph"),
      );
    }

    const detailEl = getChildBySelector(
      alertEl,
      idSelector(constants.alert.detail.id),
    );
    if (!detailEl.ok || !(detailEl.value instanceof HTMLPreElement)) {
      return Fail(new NoElWithIdAndTypeErr(constants.alert.detail.id, "pre"));
    }

    const confirmBtn = getChildBySelector(
      alertEl,
      idSelector(constants.alert.confirmButton.id),
    );
    if (!confirmBtn.ok || !(confirmBtn.value instanceof HTMLButtonElement)) {
      return Fail(
        new NoElWithIdAndTypeErr(constants.alert.confirmButton.id, "button"),
      );
    }

    const declineBtn = getChildBySelector(
      alertEl,
      idSelector(constants.alert.declineButton.id),
    );
    if (!declineBtn.ok || !(declineBtn.value instanceof HTMLButtonElement)) {
      throw new NoElWithIdAndTypeErr(
        constants.alert.declineButton.id,
        "button",
      );
    }

    return Success({
      container: alertEl,
      title: titleEl.value,
      message: msgEl.value,
      detail: detailEl.value,
      confirmBtn: confirmBtn.value,
      declineBtn: declineBtn.value,
    });
  };

  private handleAfterHtmxSwap = () => {
    const alert = getElementById(constants.alert.id);
    if (!alert.ok) {
      return;
    }

    const alertConfirmationButton = getChildBySelector(
      alert.value,
      idSelector(constants.alert.confirmButton.id),
    );
    if (!alertConfirmationButton.ok) {
      this._panic(alertConfirmationButton.error);
      return;
    }

    this._windowManager.pushAlertLayer();
    alertConfirmationButton.value.focus();
  };
}
