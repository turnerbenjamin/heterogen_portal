declare const htmx: typeof import("htmx.org").default;

export type HtmxEvent<T> = CustomEvent<T>;

export function onHtmx<T>(
  eventName: string,
  handler: (_event: HtmxEvent<T>) => void,
): void {
  htmx.on(eventName, handler as EventListener);
}
