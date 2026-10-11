import constants from "./constants";
import { NoElWithIdErr, UnexpectedErr } from "./errors";
import { Fail, Result, Success } from "./result";

export function parseHtmlElement(
  e: EventTarget | Element | undefined | null,
): Result<HTMLElement> {
  if (!e) {
    return Fail(
      new UnexpectedErr("Unable to parse HTML element: element is null"),
    );
  }

  if (!(e instanceof HTMLElement)) {
    return Fail(
      new UnexpectedErr(
        "unable to parse HTML element: element is not instance of HTML element",
      ),
    );
  }

  return Success(e);
}

export function switchSiblingElements(
  parentElement: HTMLElement,
  childSelector: string,
  a: HTMLElement,
  b: HTMLElement,
): Error | null {
  const childElements = parentElement.querySelectorAll(childSelector);

  let aIndex = -1;
  let bIndex = -1;
  for (let i = 0; i < childElements.length && (aIndex < 0 || bIndex < 0); i++) {
    const el = childElements[i];
    if (el === a) {
      aIndex = i;
    }
    if (el === b) {
      bIndex = i;
    }
  }

  if (aIndex < childElements.length - 1) {
    parentElement.insertBefore(b, childElements[aIndex + 1]);
  } else {
    parentElement.appendChild(b);
  }

  if (bIndex < childElements.length - 1) {
    parentElement.insertBefore(a, childElements[bIndex + 1]);
  } else {
    parentElement.appendChild(a);
  }

  return null;
}

export function getAppBase(): string {
  const m = document.querySelector(
    'meta[name="app-base-url"]',
  ) as HTMLMetaElement | null;
  return (m && m.content) || window.location.origin;
}

export function setVisibility(el: Element, isVisible: boolean) {
  if (isVisible) {
    el.classList.remove(constants.globalModifiers.hidden);
    el.removeAttribute("aria-hidden");
  } else {
    el.classList.add(constants.globalModifiers.hidden);
    el.setAttribute("aria-hidden", "true");
  }
}

export function classSelector(className: string): string {
  return `.${className}`;
}

export function idSelector(className: string): string {
  return `.${className}`;
}

export function attributeSelector(
  attributeName: string,
  value?: string,
): string {
  if (value === undefined) {
    return `[${attributeName}]`;
  }

  return `[${attributeName}="${value}"]`;
}

export function setBoolAttribute(el: HTMLElement, attributeName: string): void {
  el.setAttribute(attributeName, "");
}

export function getElementById(id: string): Result<HTMLElement> {
  const elRes = parseHtmlElement(document.getElementById(id));
  if (!elRes.ok) {
    return Fail(new NoElWithIdErr(id));
  }
  return Success(elRes.value);
}
