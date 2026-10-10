import { HtmxEvent, onHtmx } from "./htmx";
import * as utils from "./utils";
import { AppError } from "./errors";

interface BeforeSwapDetails {
  isError: boolean;
  shouldSwap: boolean;
  target: Element;
  swapOverride: string;
}

export interface ErrorPresenter {
  showToast(_message: string): void;
  showBlockingError(_error: AppError): void;
}

export class htmxErrorHandler {
  // TODO: Nope
  private _errorContainerSelector = ".error-container";

  public constructor() {
    onHtmx("htmx:beforeRequest", this.handleBeforeHtmxRequest);
    onHtmx<BeforeSwapDetails>("htmx:beforeSwap", this.handleBeforeHtmxSwap);
  }

  private handleBeforeHtmxRequest = (e: HtmxEvent<unknown>) => {
    const target = utils.parseHtmlElement(e.target);
    if (!target) {
      return;
    }

    const errorContainer = target.querySelector(this._errorContainerSelector);
    if (!errorContainer) {
      return;
    }

    errorContainer.classList.remove("show");
    errorContainer.setAttribute("aria-hidden", "true");
  };

  private handleBeforeHtmxSwap = (e: HtmxEvent<BeforeSwapDetails>) => {
    const target = utils.parseHtmlElement(e.target);
    if (!target) {
      return;
    }

    if (!e.detail.isError) {
      return;
    }

    e.detail.shouldSwap = true;
    const errorContainer = target.querySelector(this._errorContainerSelector);
    if (!errorContainer) {
      return;
    }

    errorContainer.setAttribute("aria-hidden", "false");
    errorContainer.classList.add("show");
    e.detail.target = errorContainer;
    e.detail.swapOverride = "innerHTML";
  };
}
