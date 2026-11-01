import constants from "./constants";
import { NoElWithIdAndTypeErr, parseAppError } from "./errors";
import { Fail, Result, Success } from "./result";
import { toastMyError } from "./toasts";
import { getElementById, setVisibility } from "./utils";
import { WindowManager } from "./windowManager";
import { HtmxEvent, HtmxAfterSwapEvent, onHtmxAfterSwapEvent } from "./htmx";

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
  private readonly _windowManager: WindowManager;
  private readonly _panic: (errObj: unknown) => void;

  public constructor(
    windowManager: WindowManager,
    panic: (errObj: unknown) => void,
  ) {
    this._windowManager = windowManager;
    this._panic = panic;

    onHtmxAfterSwapEvent(this.handleAfterHtmxSwap);
    document.addEventListener("keydown", this.handleEsc);
  }

  public hgHandlerCloseAlert = () => {
    this.hideAlert();
    this._windowManager.popAlert();
  };

  public renderNotificationMsg = (title: string, message: string): void => {
    this.renderAlert({
      type: "notification",
      title: title,
      message: message,
      confirmButtonText: "Ok",
    });
    this._windowManager.pushAlertLayer();
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
      this._windowManager.pushAlertLayer();
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

    this._windowManager.pushAlertLayer();
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

    setVisibility(alertElement.container, true);
  };

  public hideAlert() {
    const alertElementRes = this.getAlertElement();
    if (!alertElementRes.ok) {
      return this._panic(alertElementRes.error);
    }
    const alertElement = alertElementRes.value;
    setVisibility(alertElement.container, false);
  }

  private getAlertElement = (): Result<AlertElement> => {
    const alertConstants = constants.alert;

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

  private handleAfterHtmxSwap = (e: HtmxEvent<HtmxAfterSwapEvent>) => {
    if (e.detail.ctx.response?.raw.ok) {
      return;
    }

    const getAlertElRes = getElementById(constants.layers.alert.id);
    if (!getAlertElRes.ok) {
      return this._panic(getAlertElRes.error);
    }
    if (
      !getAlertElRes.value.hasAttribute(constants.alert.attributes.serverAlert)
    ) {
      return;
    }

    getAlertElRes.value.removeAttribute(constants.alert.attributes.serverAlert);
    this.renderAlert();
    this._windowManager.pushAlertLayer();
  };

  private handleEsc = (e: KeyboardEvent) => {
    if (e.key !== constants.keyCodes.Escape) {
      return;
    }

    const activeLayer = this._windowManager.getActiveLayer();
    if (!activeLayer.ok) {
      return this.renderError(activeLayer.error);
    }

    if (activeLayer.value.layerId === constants.layers.alert.id) {
      this._windowManager.popAlert();
    }
  };
}
