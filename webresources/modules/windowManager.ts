import constants from "./constants";
import { NoElWithIdErr, UnexpectedErr, WindowManagerErr } from "./errors";
import { Fail, Result, Success } from "./result";
import { attributeSelector, parseHtmlElement, setBoolAttribute } from "./utils";

interface Layer {
  layerId: string;
  element: HTMLElement;
  stackIndex: number;
  depth: number;
}

export enum Hook {
  layerChange,
  openSidebar,
  closeSidebar,
  openNestedSidebar,
  closeNestedSidebar,
  escSidebar,
  escNestedSidebar,
  escAlert,
}

type AriaNotify = (
  announcement: string,
  options?: {
    priority?: "normal" | "high";
  },
) => void;

interface DocumentWithAriaNotify extends Document {
  ariaNotify?: AriaNotify;
}

export class WindowManager {
  private _hooks: Partial<Record<Hook, Array<() => void>>> = {};
  private readonly _panic: (errObj: unknown) => void;
  private _ariaNotify: AriaNotify | null = null;

  private readonly _layerIds = {
    App: constants.layers.app.id,
    Sidebar: constants.layers.sidebar.id,
    NestedSidebar: constants.layers.nestedSidebar.id,
    Alert: constants.layers.alert.id,
  } as const;

  private readonly _layerDepths: Partial<Record<string, number>> = {
    [constants.layers.app.id]: 0,
    [constants.layers.sidebar.id]: 1,
    [constants.layers.nestedSidebar.id]: 2,
    [constants.layers.alert.id]: 3,
  } as const;

  private readonly _layers: Record<string, Layer> = {};

  private _layerStack: Layer[] = [];

  public constructor(panic: (errObj: unknown) => void) {
    this._panic = panic;

    const documentWithAriaNotify = document as DocumentWithAriaNotify;
    if (documentWithAriaNotify.ariaNotify !== undefined) {
      this._ariaNotify = documentWithAriaNotify.ariaNotify.bind(
        documentWithAriaNotify,
      );
    }

    for (const layerId of Object.values(this._layerIds)) {
      const getLayerResult = this.getLayer(layerId);
      if (!getLayerResult.ok) {
        throw getLayerResult.error;
      }
      const layer = getLayerResult.value;

      this._layers[layerId] = layer;
      if (layerId !== this._layerIds.App) {
        layer.element.setAttribute(
          constants.layers.modifierAttributes.inert,
          "",
        );
      }
    }

    const appLayer = this._layers[this._layerIds.App];
    appLayer.stackIndex = 0;
    this._layerStack = [appLayer];

    document.addEventListener("keydown", this.handleEsc);
  }

  public getActiveLayer(): Result<Layer> {
    return this.getLayerAtIndex(this._layerStack.length - 1);
  }

  public pushAlertLayer = () => {
    const pushErr = this.pushLayer(this._layerIds.Alert);
    if (pushErr) {
      this.resetStack();
      this._panic(pushErr);
    }
  };

  public pushSidebar = () => {
    const pushLayerErr = this.pushLayer(this._layerIds.Sidebar);
    if (pushLayerErr !== null) {
      this.resetStack();
      this._panic(pushLayerErr);
    }
  };

  public pushNestedSidebar = () => {
    const pushLayerErr = this.pushLayer(this._layerIds.NestedSidebar);
    if (pushLayerErr !== null) {
      this.resetStack();
      this._panic(pushLayerErr);
    }
  };

  public popAlert = () => {
    const popLayerErr = this.popLayer(this._layerIds.Alert);
    if (popLayerErr !== null) {
      this.resetStack();
      this._panic(popLayerErr);
    }
  };

  public popSidebar = () => {
    const popLayerErr = this.popLayer(this._layerIds.Sidebar);
    if (popLayerErr !== null) {
      this.resetStack();
      this._panic(popLayerErr);
    }
  };

  public popNestedSidebar = () => {
    const popLayerErr = this.popLayer(this._layerIds.NestedSidebar);
    if (popLayerErr !== null) {
      this.resetStack();
      this._panic(popLayerErr);
    }
  };

  public announcePolitely = (s: string) => {
    if (this._ariaNotify !== null) {
      this._ariaNotify(s, { priority: "normal" });
    }
  };

  public announceAssertively = (s: string) => {
    if (this._ariaNotify !== null) {
      this._ariaNotify(s, { priority: "high" });
    }
  };

  private pushLayer = (layerId: string): Error | null => {
    const newActiveLayer = this._layers[layerId];
    if (newActiveLayer.stackIndex !== -1) {
      return new WindowManagerErr(
        `Invalid attempt to push layer ${layerId} to the stack: ${layerId} ` +
          `is already in the layer stack`,
      );
    }

    // Blur the current element
    const activeEl = parseHtmlElement(document.activeElement);
    if (activeEl.ok) {
      activeEl.value.blur();
    }

    // get the currently active layer
    const getActiveLayerResult = this.getActiveLayer();
    if (!getActiveLayerResult.ok) {
      return getActiveLayerResult.error;
    }
    const activeLayer = getActiveLayerResult.value;

    // Guard that active layer is higher than the current layer
    if (newActiveLayer.depth < activeLayer.depth) {
      return new UnexpectedErr(
        `unable to push layer to stack: ${layerId} has a lower depth value ` +
          `than the current active layer, ${activeLayer.layerId}`,
      );
    }

    // Set the current active layer to inert
    setBoolAttribute(
      activeLayer.element,
      constants.layers.modifierAttributes.inert,
    );

    // Set the new active layer to active and push it to the stack
    newActiveLayer.element.removeAttribute(
      constants.layers.modifierAttributes.inert,
    );
    newActiveLayer.stackIndex = this._layerStack.length;
    this._layerStack.push(newActiveLayer);

    this.notifyPush(newActiveLayer.layerId);

    return null;
  };

  private popLayer = (layerId?: string): Error | null => {
    // get current active layer
    const getCurrentLayerResult = this.getActiveLayer();
    if (!getCurrentLayerResult.ok) {
      return getCurrentLayerResult.error;
    }
    const currentLayer = getCurrentLayerResult.value;

    if (layerId && currentLayer.layerId !== layerId) {
      return new WindowManagerErr(`
        Unable to pop layer ${layerId} from the stack as it is not currently the active layer
        `);
    }

    // get new active layer
    const getNextActiveLayerRes = this.getLayerAtIndex(
      this._layerStack.length - 2,
    );
    if (!getNextActiveLayerRes.ok) {
      return getNextActiveLayerRes.error;
    }
    const nextActiveLayer = getNextActiveLayerRes.value;

    // Blur the current element
    const activeEl = parseHtmlElement(document.activeElement);
    if (activeEl.ok) {
      activeEl.value.blur();
    }

    // Get next element to focus
    const nextElementToFocusRes = parseHtmlElement(
      nextActiveLayer.element.querySelector(
        attributeSelector(constants.layers.contentShape.initialFocusAttribute),
      ),
    );
    if (!nextElementToFocusRes.ok) {
      return new WindowManagerErr(
        `Unable to pop layer ${layerId}. ${nextActiveLayer.layerId} must ` +
          `contain an child element with ` +
          `the '${constants.layers.contentShape.initialFocusAttribute}' attribute`,
      );
    }
    const nextElementToFocus = nextElementToFocusRes.value;

    // pop the current layer and set it as inert
    this._layerStack.pop();
    currentLayer.stackIndex = -1;
    currentLayer.element.setAttribute(
      constants.layers.modifierAttributes.inert,
      "",
    );

    // make the new active layer active
    nextActiveLayer.element.removeAttribute(
      constants.layers.modifierAttributes.inert,
    );
    nextElementToFocus.focus();

    this.notifyPop(currentLayer.layerId);

    return null;
  };

  public registerHook = (h: Hook, fn: () => void) => {
    if (!this._hooks[h]) {
      this._hooks[h] = [];
    }
    this._hooks[h].push(fn);
  };

  private invokeHook = (h: Hook) => {
    if (!this._hooks[h]) {
      return;
    }

    for (const callback of this._hooks[h]) {
      callback();
    }
  };

  private notifyPush(layerId: string) {
    this.invokeHook(Hook.layerChange);

    switch (layerId) {
      case this._layerIds.Sidebar:
        this.invokeHook(Hook.openSidebar);
        break;
      case this._layerIds.NestedSidebar:
        this.invokeHook(Hook.openNestedSidebar);
        break;
      default:
        return;
    }
  }

  private notifyPop(layerId: string) {
    this.invokeHook(Hook.layerChange);

    switch (layerId) {
      case this._layerIds.Sidebar:
        this.invokeHook(Hook.closeSidebar);
        break;
      case this._layerIds.NestedSidebar:
        this.invokeHook(Hook.closeNestedSidebar);
        break;
      default:
        return;
    }
  }

  private getLayerAtIndex(n: number): Result<Layer> {
    if (this._layerStack.length === 0) {
      return Fail(
        new WindowManagerErr(
          `Unable to retrieve layer ${n} from the layer stack: The stack is ` +
            `empty`,
        ),
      );
    }

    if (n >= this._layerStack.length) {
      return Fail(
        new WindowManagerErr(
          `Unable to retrieve layer ${n} from the layer stack: The index is ` +
            `out of bounds`,
        ),
      );
    }

    return Success(this._layerStack[n]);
  }

  private getLayer = (layerId: string): Result<Layer> => {
    const layerElRes = parseHtmlElement(document.getElementById(layerId));
    if (!layerElRes.ok) {
      return Fail(new NoElWithIdErr(layerId));
    }

    const depth = this._layerDepths[layerId];
    if (depth === undefined) {
      return Fail(
        new UnexpectedErr(
          `unable to get layer: cannot access depth for ${layerId}`,
        ),
      );
    }

    return Success({
      layerId: layerId,
      element: layerElRes.value,
      stackIndex: -1,
      depth: depth,
    });
  };

  private resetStack = () => {
    const appLayer = this._layers[this._layerIds.App];
    appLayer.stackIndex = 0;
    appLayer.element.removeAttribute(constants.layers.modifierAttributes.inert);
    this._layerStack = [appLayer];

    for (const layer of Object.values(this._layers)) {
      if (layer.layerId === this._layerIds.App) {
        continue;
      }

      layer.stackIndex = -1;
      setBoolAttribute(
        layer.element,
        constants.layers.modifierAttributes.inert,
      );
      layer.element.classList.add(constants.globalModifiers.hidden);
      layer.element.classList.remove(constants.layers.modifierClasses.show);
      layer.element.classList.remove(constants.layers.modifierClasses.close);
    }
  };

  private handleEsc = (e: KeyboardEvent) => {
    if (e.key !== constants.keyCodes.Escape) {
      return;
    }

    const activeLayerReq = this.getActiveLayer();
    if (!activeLayerReq.ok) {
      return;
    }

    switch (activeLayerReq.value.layerId) {
      case this._layerIds.Sidebar:
        this.invokeHook(Hook.escSidebar);
        break;
      case this._layerIds.NestedSidebar:
        this.invokeHook(Hook.escNestedSidebar);
        break;
      case this._layerIds.Alert:
        this.invokeHook(Hook.escAlert);
        break;
    }
  };
}
