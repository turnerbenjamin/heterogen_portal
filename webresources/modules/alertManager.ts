import constants from "./constants";
import { parseAppError, UnexpectedErr } from "./errors";
import { Fail, Result, Success } from "./result";
import { toastMyError } from "./toasts";
import {
  classSelector,
  getChildByClass,
  getElementById,
  getTemplateById,
  getTemplateContentBySelector,
  parseHtmlElement,
  setVisibility,
} from "./utils";
import { Hook, WindowManager } from "./windowManager";
import { HtmxEvent, HtmxAfterSwapEvent, onHtmxAfterSwapEvent } from "./htmx";

interface Alert {
  type: "notification" | "error" | "serverError" | "confirmation";
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

interface ServerErrorAlert extends Alert {
  type: "serverError";
  title: string;
  message: string;
  confirmButtonText: "Ok";
}

interface ConfirmationAlert extends Alert {
  type: "confirmation";
  title: string;
  message: string;
  confirmButtonText: "Continue";
  declineButtonText: "Cancel";
}

type GeneralAlert = NotificationAlert | ConfirmationAlert;
type PriorityAlert = ErrorAlert | ServerErrorAlert;

type AnyAlert = GeneralAlert | PriorityAlert;

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

  private _activeAlert: AnyAlert | null = null;
  private readonly _generalAlertStack: GeneralAlert[] = [];
  private readonly _priorityAlertStack: PriorityAlert[] = [];

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
    this.hideAlert();
  };

  public renderNotificationMsg = (title: string, message: string): void => {
    this._generalAlertStack.push({
      type: "notification",
      title: title,
      message: message,
      confirmButtonText: "Ok",
    });
    this.refreshAlerts();
  };

  public renderError = (errObj: unknown): void => {
    const err = parseAppError(errObj);
    if (err.doBlock) {
      this._priorityAlertStack.push({
        type: "error",
        title: "Error",
        message: err.message,
        detail: err.details,
        confirmButtonText: "Ok",
      });

      this.refreshAlerts();
    } else {
      const toastErr = toastMyError(err.message);
      if (toastErr) {
        this._panic(toastErr);
      }
    }
  };

  public renderSeverError = (message: string): void => {
    this._priorityAlertStack.push({
      type: "serverError",
      title: "Server Error",
      message: message,
      confirmButtonText: "Ok",
    });

    this.refreshAlerts();
  };

  public renderConfirmationDialog = (title: string, message: string): void => {
    this._generalAlertStack.push({
      type: "confirmation",
      title: title,
      message: message,
      confirmButtonText: "Continue",
      declineButtonText: "Cancel",
    });
    this.refreshAlerts();
  };

  private refreshAlerts = (): void => {
    // remove the current alert
    let err = this.clearActiveAlert();
    if (err !== null) {
      return this._panic(err);
    }

    // Get the next alert
    let nextAlert: AnyAlert | null = null;
    if (this._priorityAlertStack.length > 0) {
      nextAlert = this._priorityAlertStack[this._priorityAlertStack.length - 1];
    } else if (this._generalAlertStack.length > 0) {
      nextAlert = this._generalAlertStack[this._generalAlertStack.length - 1];
    }

    // If the stack is empty, pop the alert layer and return early
    if (nextAlert === null) {
      this._windowManager.popAlert();
      this._activeAlert = null;
      return;
    }

    // Configure the alert
    const alertElementRes = this.getAlertElement();
    if (!alertElementRes.ok) {
      return this._panic(alertElementRes.error);
    }
    const alertElement = alertElementRes.value;
    alertElement.container.classList.remove(constants.alert.modifiers.error);

    alertElement.title.textContent = nextAlert.title;
    alertElement.message.textContent = nextAlert.message;
    alertElement.confirmBtn.textContent = nextAlert.confirmButtonText;

    setVisibility(alertElement.detail, nextAlert?.type === "error");
    if (nextAlert.type === "error") {
      alertElement.detail.textContent = nextAlert.detail;
    }

    if (nextAlert.type === "error" || nextAlert.type === "serverError") {
      alertElement.container.classList.add(constants.alert.modifiers.error);
    }

    setVisibility(alertElement.declineBtn, nextAlert?.type === "confirmation");
    if (nextAlert.type === "confirmation") {
      alertElement.detail.textContent = nextAlert.declineButtonText;
    }

    // If there is not currently active alert push the alert layer
    if (this._activeAlert === null) {
      this._windowManager.pushAlertLayer();
    }
    alertElement.confirmBtn.focus();

    this._activeAlert = nextAlert;
  };

  private hideAlert() {
    if (this._activeAlert === null) {
      this._panic(
        new UnexpectedErr(`unable to hide alert: active alert is null`),
      );
    }

    if (
      this._priorityAlertStack.length > 0 &&
      this._priorityAlertStack[this._priorityAlertStack.length - 1] ===
        this._activeAlert
    ) {
      this._priorityAlertStack.pop();
      return this.refreshAlerts();
    }

    if (
      this._generalAlertStack.length > 0 &&
      this._generalAlertStack[this._generalAlertStack.length - 1] ===
        this._activeAlert
    ) {
      this._generalAlertStack.pop();
      return this.refreshAlerts();
    }

    this._panic(
      new UnexpectedErr(
        `unable to hide alert: active alert cannot be found in either stack`,
      ),
    );
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
      classSelector(constants.alert.className),
    );
    if (!alertPrototypeRes.ok) {
      return alertPrototypeRes;
    }

    const alertEl = alertPrototypeRes.value.cloneNode(true) as HTMLElement;
    alertLayerRes.value.appendChild(alertEl);

    const titleEl = getChildByClass(alertEl, constants.alert.title.className);
    if (!titleEl.ok || !(titleEl.value instanceof HTMLHeadingElement)) {
      return Fail(
        new UnexpectedErr(
          `unable to get alert element: expected alert to contain child with ` +
            `class ${constants.alert.title.className} and ` +
            `type ${HTMLHeadingElement.name}`,
        ),
      );
    }

    const msgEl = getChildByClass(alertEl, constants.alert.message.className);
    if (!msgEl.ok || !(msgEl.value instanceof HTMLParagraphElement)) {
      return Fail(
        new UnexpectedErr(
          `unable to get alert element: expected alert to contain child with ` +
            `class ${constants.alert.message.className} and ` +
            `type ${HTMLParagraphElement.name}`,
        ),
      );
    }

    const detailEl = getChildByClass(alertEl, constants.alert.detail.className);
    if (!detailEl.ok || !(detailEl.value instanceof HTMLPreElement)) {
      return Fail(
        new UnexpectedErr(
          `unable to get alert element: expected alert to contain child with ` +
            `class ${constants.alert.detail.className} and ` +
            `type ${HTMLPreElement.name}`,
        ),
      );
    }

    const confirmBtn = getChildByClass(
      alertEl,
      constants.alert.confirmButton.className,
    );
    if (!confirmBtn.ok || !(confirmBtn.value instanceof HTMLButtonElement)) {
      return Fail(
        new UnexpectedErr(
          `unable to get alert element: expected alert to contain child with ` +
            `class ${constants.alert.confirmButton.className} and ` +
            `type ${HTMLButtonElement.name}`,
        ),
      );
    }

    const declineBtn = getChildByClass(
      alertEl,
      constants.alert.declineButton.className,
    );
    if (!declineBtn.ok || !(declineBtn.value instanceof HTMLButtonElement)) {
      return Fail(
        new UnexpectedErr(
          `unable to get alert element: expected alert to contain child with ` +
            `class ${constants.alert.declineButton.className} and ` +
            `type ${HTMLButtonElement.name}`,
        ),
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

  private handleAfterHtmxSwap = (_e: HtmxEvent<HtmxAfterSwapEvent>) => {
    // if no error return early
    if (_e.detail.ctx.response?.raw.ok) {
      return;
    }

    // get the alert layer
    const alertLayer = getElementById(constants.layers.alert.id);
    if (!alertLayer.ok) {
      return this._panic(alertLayer.error);
    }

    // read server errors and assert that just one exists
    const errorSelector = ".alert[server-error]";
    const serverErrors = alertLayer.value.querySelectorAll(errorSelector);
    if (serverErrors.length !== 1) {
      return this._panic(
        `unable to handle server error: expected alert layer to contain an ` +
          `1 child matching ${errorSelector} but got ${serverErrors.length}`,
      );
    }

    // parse the server error
    const serverErrorEl = parseHtmlElement(serverErrors[0]);
    if (!serverErrorEl.ok) {
      return this._panic(serverErrorEl.error);
    }

    // remove the server error from the dom - allow normal flow to take over
    serverErrorEl.value.remove();
    this.renderSeverError(
      `The server has returned an error; please try again later. If ` +
        `the problem persists, please contact technical support`,
    );
  };

  private clearActiveAlert = (): Error | null => {
    const alertLayer = getElementById(constants.layers.alert.id);
    if (!alertLayer.ok) {
      return alertLayer.error;
    }

    const alerts = alertLayer.value.querySelectorAll(
      classSelector(constants.alert.className),
    );

    const expectedAlertCount = this._activeAlert !== null ? 1 : 0;
    if (alerts.length !== expectedAlertCount) {
      return new UnexpectedErr(
        `unable to clear active alert: expected alert count ` +
          `of ${expectedAlertCount} but got ${alerts.length}`,
      );
    }

    if (alerts.length === 0) {
      return null;
    }
    alerts[0].remove();

    return null;
  };
}
