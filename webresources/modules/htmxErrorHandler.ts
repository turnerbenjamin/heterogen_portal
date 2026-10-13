import {
  HtmxEvent,
  onHtmxBeforeSwapEvent,
  BeforeSwapEventContext,
  onHtmxBeforeRequestEvent,
} from "./htmx";
import * as utils from "./utils";
import { AppError } from "./errors";
import { HtmxRequestCtx } from "htmx.org";

export interface ErrorPresenter {
  showToast(_message: string): void;
  showBlockingError(_error: AppError): void;
}

export class htmxErrorHandler {
  // TODO: Nope
  private _errorContainerSelector = ".error-container";

  public constructor() {
    onHtmxBeforeRequestEvent(this.handleBeforeHtmxRequest);
    onHtmxBeforeSwapEvent(this.handleBeforeHtmxSwap);
  }

  private handleBeforeHtmxRequest = (e: HtmxEvent<HtmxRequestCtx>) => {
    const targetRes = utils.parseHtmlElement(e.target);
    if (!targetRes.ok) {
      return;
    }
    const target = targetRes.value;

    const errorContainer = target.querySelector(this._errorContainerSelector);
    if (!errorContainer) {
      return;
    }

    errorContainer.classList.remove("show");
    errorContainer.setAttribute("aria-hidden", "true");
  };

  private handleBeforeHtmxSwap = (e: HtmxEvent<BeforeSwapEventContext>) => {
    const targetRes = utils.parseHtmlElement(e.target);
    if (!targetRes.ok) {
      return;
    }
    const target = targetRes.value;

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
