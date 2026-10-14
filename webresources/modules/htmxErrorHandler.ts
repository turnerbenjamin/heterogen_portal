import {
  HtmxEvent,
  HtmxResponseErrorEventContext,
  onHtmxResponseError,
} from "./htmx";
import { AppError } from "./errors";

export interface ErrorPresenter {
  showToast(_message: string): void;
  showBlockingError(_error: AppError): void;
}

export class htmxErrorHandler {
  public constructor() {
    onHtmxResponseError(this.handleHtmxResponseError);
  }

  private handleHtmxResponseError = (
    e: HtmxEvent<HtmxResponseErrorEventContext>,
  ) => {
    e.detail.ctx.push = false;
  };
}
