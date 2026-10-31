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

export function htmlElementAs<T extends HTMLElement>(
  el: Element | null | undefined,
  elementType: abstract new (...args: never[]) => T,
): Result<T> {
  if (!el) {
    return Fail(
      new UnexpectedErr("Unable to parse HTML element: element is null"),
    );
  }

  if (!(el instanceof elementType)) {
    return Fail(
      new UnexpectedErr(
        "unable to parse HTML element: element is not an instance of the expected type",
      ),
    );
  }

  return Success(el);
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

export function setBoolAttribute(
  el: HTMLElement,
  attributeName: string,
  value = true,
): void {
  if (value === true) {
    el.setAttribute(attributeName, "");
  } else {
    el.removeAttribute(attributeName);
  }
}

export function getElementById(id: string): Result<HTMLElement> {
  const elRes = parseHtmlElement(document.getElementById(id));
  if (!elRes.ok) {
    return Fail(new NoElWithIdErr(id));
  }
  return Success(elRes.value);
}

export function getTarget(e: Event): Result<HTMLElement> {
  const t = parseHtmlElement(e.target);
  if (!t.ok) {
    return Fail(
      new UnexpectedErr(`Unable to parse target for event: ${e.type}`),
    );
  }
  return t;
}

export function getCurrentTarget(e: Event): Result<HTMLElement> {
  const t = parseHtmlElement(e.currentTarget);
  if (!t.ok) {
    return Fail(
      new UnexpectedErr(`Unable to parse current target for event: ${e.type}`),
    );
  }
  return t;
}

export function getTargetInput(e: Event): Result<HTMLInputElement> {
  const t = getTarget(e);
  if (!t.ok) {
    return t;
  }

  return htmlElementAs(t.value, HTMLInputElement);
}

export function getTargetSelect(e: Event): Result<HTMLSelectElement> {
  const t = getTarget(e);
  if (!t.ok) {
    return t;
  }

  return htmlElementAs(t.value, HTMLSelectElement);
}

export function getChildByClass(
  parent: HTMLElement,
  childClass: string,
): Result<HTMLElement> {
  const c = parseHtmlElement(parent.querySelector(classSelector(childClass)));
  if (!c.ok) {
    return Fail(
      new UnexpectedErr(
        `Unable to access element with class name ${childClass} from ${elString(parent)}`,
      ),
    );
  }
  return c;
}

export function getChildBySelector(
  parent: HTMLElement,
  selector: string,
): Result<HTMLElement> {
  const c = parseHtmlElement(parent.querySelector(selector));
  if (!c.ok) {
    return Fail(
      new UnexpectedErr(
        `Unable to access element matching selector '${selector}' from ${elString(parent)}`,
      ),
    );
  }
  return c;
}

export function getTemplateContentByClass(
  parent: HTMLTemplateElement,
  childClass: string,
): Result<HTMLElement> {
  const c = parseHtmlElement(
    parent.content.querySelector(classSelector(childClass)),
  );
  if (!c.ok) {
    return Fail(
      new UnexpectedErr(
        `Unable to access element with class name ${childClass} from ${elString(parent)}`,
      ),
    );
  }
  return c;
}

export function getTemplateContentBySelector(
  parent: HTMLTemplateElement,
  selector: string,
): Result<HTMLElement> {
  const c = parseHtmlElement(parent.content.querySelector(selector));
  if (!c.ok) {
    return Fail(
      new UnexpectedErr(
        `Unable to access element matching '${selector}' from ${elString(parent)}`,
      ),
    );
  }
  return c;
}

export function getClosestByClass(
  child: HTMLElement,
  parentClass: string,
): Result<HTMLElement> {
  const c = parseHtmlElement(child.closest(classSelector(parentClass)));
  if (!c.ok) {
    return Fail(
      new UnexpectedErr(
        `Unable to access parent with class name ${parentClass} from ${elString(child)}`,
      ),
    );
  }
  return c;
}

export function getSelectedOption(
  el: HTMLSelectElement,
): Result<HTMLOptionElement | null> {
  const options = el.selectedOptions;
  if (options.length === 0) {
    return Success(null);
  }

  if (options.length > 1) {
    return Fail(
      new UnexpectedErr(
        `expected ${elString(el)} to have 0-1 selected options but found ${options.length}`,
      ),
    );
  }

  return Success(options[0]);
}

export function getTemplateById(id: string): Result<HTMLTemplateElement> {
  const templ = getElementById(id);
  if (!templ.ok) {
    return templ;
  }

  if (templ.value.tagName !== "TEMPLATE") {
    return Fail(
      new UnexpectedErr(
        `expected ${id} tag to be template but got ${templ.value.tagName}`,
      ),
    );
  }

  return Success(templ.value as HTMLTemplateElement);
}

export function elString(el: HTMLElement): string {
  if (el.id) {
    return `${el.id}:${el.tagName}(${[el.classList.toString()]})`;
  }

  return `${el.tagName}(${[el.classList.toString()]})`;
}
