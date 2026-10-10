export function parseAppError(errorObj: unknown): AppError {
  if (errorObj instanceof AppError) {
    return errorObj;
  }

  if (errorObj instanceof Error) {
    return new UnexpectedErr(errorObj.message);
  }

  if (errorObj === undefined || errorObj === null) {
    return new UnexpectedErr(`A null/undefined error has been passed`);
  }

  return new UnexpectedErr(JSON.stringify(errorObj));
}

export class AppError extends Error {
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
    super(`Expected element with id '${elementId}' to be in the DOM`);
  }
}

export class NoElWithIdAndTypeErr extends UnexpectedErr {
  public constructor(elementId: string, type: string) {
    super(
      `Expected element with id '${elementId}' and type '${type}'to be in the DOM`,
    );
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
