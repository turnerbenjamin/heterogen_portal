declare const htmx: typeof import("htmx.org").default;
// export const HTMX = htmx;

type HtmxEvent<T> = CustomEvent<T>;

export function configureCacheSize(n: number) {
  htmx.config.historyCacheSize = n;
}

export function onHtmx<T>(
  eventName: string,
  handler: (_event: HtmxEvent<T>) => void,
): void {
  htmx.on(eventName, handler as EventListener);
}
