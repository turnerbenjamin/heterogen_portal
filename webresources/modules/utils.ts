/**
 * Helper method to extract the target element from an event
 * @param e Event
 * @returns Element or null
 */
export function ParseHtmlElement(
  t: EventTarget | Element | undefined | null,
): HTMLElement | null {
  if (!(t instanceof HTMLElement)) {
    return null;
  }
  return t as HTMLElement;
}

export function SwitchSiblingElements(
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

export function idSelector(id: string) {
  return `#${id}`;
}

export function getAppBase(): string {
  const m = document.querySelector(
    'meta[name="app-base-url"]',
  ) as HTMLMetaElement | null;
  return (m && m.content) || window.location.origin;
}

export function setVisibility(el: Element, isVisible: boolean) {
  if (isVisible) {
    el.classList.remove("hidden");
    el.removeAttribute("aria-hidden");
  } else {
    el.classList.add("hidden");
    el.setAttribute("aria-hidden", "true");
  }
}
