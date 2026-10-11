import { HtmxRequestCtx } from "htmx.org";

declare const htmx: typeof import("htmx.org").default;

export type HtmxEvent<T extends HtmxRequestCtx> = CustomEvent<T>;

export interface BeforeSwapEventContext extends HtmxRequestCtx {
  isError: boolean;
  shouldSwap: boolean;
  swapOverride: string;
}

export interface ConfigRequestEventContext extends HtmxRequestCtx {
  elt: Element;
  formData: FormData;
  headers: Partial<Record<string, string | null>>;
  path: string;
  useUrlParams: boolean;
  verb: string;
  withCredentials: boolean;
}

export function onHtmxBeforeSwapEvent(
  handler: (_event: HtmxEvent<BeforeSwapEventContext>) => void,
) {
  return htmx.on("htmx:beforeSwap", handler as EventListener);
}

export function onHtmxBeforeRequestEvent(
  handler: (_event: HtmxEvent<HtmxRequestCtx>) => void,
) {
  return htmx.on("htmx:beforeRequest", handler as EventListener);
}

export function onHtmxConfigRequestEvent(
  handler: (_event: HtmxEvent<ConfigRequestEventContext>) => void,
) {
  return htmx.on("htmx:configRequest", handler as EventListener);
}
