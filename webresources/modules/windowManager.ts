import constants from "./constants";
import { NoElWithIdErr, NullChildErr, WindowManagerErr } from "./errors";
import { Fail, Result, Success } from "./result";
import { attributeSelector, parseHtmlElement, setBoolAttribute } from "./utils";

interface Layer {
  layerId: string;
  element: HTMLElement;
  stackIndex: number;
}

export enum Hook {
  layerChange,
  openSidebar,
  closeSidebar,
  openNestedSidebar,
  closeNestedSidebar,
}

export class WindowManager {
  private _hooks: Partial<Record<Hook, Array<() => void>>> = {};
  private readonly _panic: (errObj: unknown) => void;

  private readonly _layerIds = {
    App: constants.layers.app.id,
    Sidebar: constants.layers.sidebar.id,
    NestedSidebar: constants.layers.nestedSidebar.id,
    Alert: constants.layers.alert.id,
  } as const;

  private readonly _layers: Record<string, Layer> = {};

  private _layerStack: Layer[] = [];

  public constructor(panic: (errObj: unknown) => void) {
    this._panic = panic;

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
  }

  public getActiveLayer(): Result<Layer> {
    return this.getLayerAtIndex(this._layerStack.length - 1);
  }

  public pushAlertLayer = () => {
    const pushErr = this.pushLayer(this._layerIds.Alert);
    if (pushErr) {
      this.resetStack();
      this._panic;
    }
  };

  public pushSidebar = () => {
    const pushLayerErr = this.pushLayer(this._layerIds.Sidebar);
    if (pushLayerErr !== null) {
      this.resetStack();
      this._panic;
    }
  };

  public pushNestedSidebar = () => {
    const pushLayerErr = this.pushLayer(this._layerIds.NestedSidebar);
    if (pushLayerErr !== null) {
      this.resetStack();
      this._panic;
    }
  };

  public popAlert = () => {
    const popLayerErr = this.popLayer(this._layerIds.Alert);
    if (popLayerErr !== null) {
      this.resetStack();
      this._panic;
    }
  };

  public popSidebar = () => {
    const popLayerErr = this.popLayer(this._layerIds.Sidebar);
    if (popLayerErr !== null) {
      this.resetStack();
      this._panic;
    }
  };

  public popNestedSidebar = () => {
    const popLayerErr = this.popLayer(this._layerIds.NestedSidebar);
    if (popLayerErr !== null) {
      this.resetStack();
      this._panic;
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

    // Get next element to focus
    const nextElementToFocusRes = parseHtmlElement(
      newActiveLayer.element.querySelector(
        attributeSelector(constants.layers.contentShape.initialFocusAttribute),
      ),
    );
    if (!nextElementToFocusRes.ok) {
      return new NullChildErr(
        newActiveLayer.element,
        attributeSelector(constants.layers.contentShape.initialFocusAttribute),
      );
    }
    const nextElementToFocus = nextElementToFocusRes.value;

    // get the currently active layer
    const getActiveLayerResult = this.getActiveLayer();
    if (!getActiveLayerResult.ok) {
      return getActiveLayerResult.error;
    }
    const activeLayer = getActiveLayerResult.value;

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
    nextElementToFocus.focus();

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

    return Success({
      layerId: layerId,
      element: layerElRes.value,
      stackIndex: -1,
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
}
