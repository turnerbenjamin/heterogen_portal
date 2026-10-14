import { HtmxRequestCtx } from "htmx.org";

declare const htmx: typeof import("htmx.org").default;

export type HtmxEvent<T extends BaseHtmxEvent> = CustomEvent<T>;

interface BaseHtmxEvent {
  ctx: HtmxRequestCtx;
}

export interface HtmxErrorEventContext extends BaseHtmxEvent {
  error: unknown;
}

export interface HtmxBeforeRequestEventContext extends BaseHtmxEvent {}
export interface HtmxConfigRequestEvent extends BaseHtmxEvent {}
export interface HtmxAfterSwapEvent extends BaseHtmxEvent {}
export interface HtmxResponseErrorEventContext extends BaseHtmxEvent {}

export function onHtmxError(
  handler: (_event: HtmxEvent<HtmxErrorEventContext>) => void,
) {
  return htmx.on("htmx:error", handler as EventListener);
}

export function onHtmxResponseError(
  handler: (_event: HtmxEvent<HtmxResponseErrorEventContext>) => void,
) {
  return htmx.on("htmx:response:error", handler as EventListener);
}

export function onHtmxAfterSwapEvent(
  handler: (_event: HtmxEvent<HtmxAfterSwapEvent>) => void,
) {
  return htmx.on("htmx:after:swap", handler as EventListener);
}

export function onHtmxBeforeRequestEvent(
  handler: (_event: HtmxEvent<HtmxBeforeRequestEventContext>) => void,
) {
  return htmx.on("htmx:before:request", handler as EventListener);
}

export function onHtmxConfigRequestEvent(
  handler: (_event: HtmxEvent<HtmxConfigRequestEvent>) => void,
) {
  return htmx.on("htmx:config:request", handler as EventListener);
}
