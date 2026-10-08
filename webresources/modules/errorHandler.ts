import { HtmxBeforeSwapDetails } from "htmx.org";
import { HtmxEvent, onHtmx } from "./htmx";
import * as utils from "./utils";

export class ErrorHandler {
  private _errorContainerSelector = ".error-container";

  public constructor() {
    // Global try/catch
    addEventListener("error", (e: ErrorEvent) => {
      e.preventDefault();
      this.handle(e.error);
    });

    onHtmx("htmx:beforeRequest", this.handleBeforeHtmxRequest);
    onHtmx<HtmxBeforeSwapDetails>("htmx:beforeSwap", this.handleBeforeHtmxSwap);
  }

  public handle = (errorObj: unknown) => {
    let error: Error;
    if (errorObj instanceof Error) {
      error = errorObj;
    } else {
      if (errorObj === undefined || errorObj === null) {
        error = new Error("An unexpected error has occurred");
      } else {
        error = new Error(JSON.stringify(errorObj));
      }
    }

    const toastContainer = document.getElementById("toast-container");
    if (!toastContainer) {
      return;
    }

    const toastDiv = document.createElement("div");
    toastDiv.classList.add("toast", "show", "error");

    const errorMessage = document.createElement("p");
    errorMessage.innerText = error.message;

    const progressBar = document.createElement("div");
    progressBar.classList.add("progress-bar");

    toastDiv.appendChild(errorMessage);
    toastDiv.appendChild(progressBar);

    toastContainer.prepend(toastDiv);
  };

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

  private handleBeforeHtmxSwap = (e: HtmxEvent<HtmxBeforeSwapDetails>) => {
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

class AppError extends Error {
  public details: string;
  public doBlock: boolean;

  public constructor(msg: string, details: string, doBlock: boolean) {
    super(msg);

    this.name = new.target.name;
    this.doBlock = doBlock;
    this.details = this.addContextToDetails(details, this.stack);
  }

  private addContextToDetails = (details: string, stack?: string): string => {
    return JSON.stringify(
      {
        OccurredAt: new Date().toISOString(),

        Message: details,
        Stack: stack,

        Host: window.location.host,
        PathName: window.location.pathname,

        Browser: navigator.userAgent,
      },
      null,
      2,
    );
  };
}

export class UnexpectedErr extends AppError {
  private static readonly genericErrorMessage =
    "An unexpected error has occured, the app may not function as expected; " +
    "please refresh the page. If this issue persists pass the details of " +
    "this error on to technical support";

  public constructor(details: string) {
    super(UnexpectedErr.genericErrorMessage, details, true);
  }
}

export class NoElWithIdErr extends UnexpectedErr {
  public constructor(elementId: string) {
    super(`Expected element with id ${elementId} to be in the DOM`);
  }
}

export class NullParentErr extends UnexpectedErr {
  public constructor(el: HTMLElement, parentSelector: string) {
    super(
      `Expected ${elString(el)} to have a parent matching the selector ` +
        `${parentSelector}, but found null`,
    );
  }
}

export class NullChildErr extends UnexpectedErr {
  public constructor(el: HTMLElement, childSelector: string) {
    super(
      `Expected ${elString(el)} to have a child matching the selector ` +
        `${childSelector}, but found null`,
    );
  }
}

export class QSelectFail extends UnexpectedErr {
  public constructor(selector: string, elementId: string) {
    super(
      `Expected ${elementId} to contain at least one element matching the ` +
        `css selector ${selector}`,
    );
  }
}

export class NullTargetErr extends UnexpectedErr {
  public constructor(e: Event) {
    super(
      `Expected ${e.type} event to contain a target of type HTMLElement, but ` +
        `recevied null`,
    );
  }
}

export class NullCurrentTargetErr extends UnexpectedErr {
  public constructor(e: Event) {
    super(
      `Expected ${e.type} event to contain a current target of type ` +
        `HTMLElement, but recevied null`,
    );
  }
}

export class NullAttributeErr extends UnexpectedErr {
  public constructor(el: HTMLElement, attributeName: string) {
    super(`Expected ${elString(el)} to contain a ${attributeName} attribute`);
  }
}

export class WindowManagerErr extends UnexpectedErr {
  public constructor(detail: string) {
    super(`Window manager error: ${detail}`);
  }
}

function elString(el: HTMLElement): string {
  if (el.id) {
    return `${el.id}:${el.tagName}(${[el.classList.toString()]})`;
  }

  return `${el.tagName}(${[el.classList.toString()]})`;
}
