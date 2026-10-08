import constants from "./constants";
import {
  ErrorHandler,
  NoElWithIdErr,
  NullChildErr,
  WindowManagerErr,
} from "./errorHandler";
import { Fail, Result, Success } from "./result";
import { attributeSelector, parseHtmlElement, setBoolAttribute } from "./utils";

interface Layer {
  layerId: string;
  element: HTMLElement;
  stackIndex: number;
}

export class WindowManager {
  private readonly _errorHandler: ErrorHandler;

  private readonly _layerIds = {
    App: constants.layers.app.id,
    Sidebar: constants.layers.sidebar.id,
    NestedSidebar: constants.layers.nestedSidebar.id,
    Alert: constants.layers.alert.id,
  } as const;

  private readonly _layers: Record<string, Layer> = {};

  private _layerStack: Layer[] = [];

  private openedSidebarEvent = new CustomEvent<Event>(
    constants.layers.sidebar.events.opened,
  );
  private closedSidebarEvent = new CustomEvent<Event>(
    constants.layers.sidebar.events.closed,
  );
  private openedNestedSidebarEvent = new CustomEvent<Event>(
    constants.layers.nestedSidebar.events.opened,
  );
  private closedNestedSidebarEvent = new CustomEvent<Event>(
    constants.layers.nestedSidebar.events.closed,
  );

  public constructor(errorHandler: ErrorHandler) {
    this._errorHandler = errorHandler;

    for (const layerId of Object.values(this._layerIds)) {
      const getLayerResult = this.getLayer(layerId);
      if (!getLayerResult.ok) {
        // Generally avoiding throw, however, an error is possible and I don't
        // want to use an init function - This is a panic throw, if it throws
        // it throws on app load and it should never throw if the HTML layout
        // is correct
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

  public GetSidebar = (): HTMLElement => {
    return this._layers[this._layerIds.Sidebar].element;
  };

  public GetNestedSidebar = (): HTMLElement => {
    return this._layers[this._layerIds.NestedSidebar].element;
  };

  public hgHandlerShowSidebar = () => {
    const pushLayerErr = this.pushLayer(this._layerIds.Sidebar);
    if (pushLayerErr !== null) {
      this.resetStack();
      this._errorHandler.handle(pushLayerErr);
    }
  };

  public hgHandlerCloseSidebar = () => {
    const popLayerErr = this.popLayer(this._layerIds.Sidebar);
    if (popLayerErr !== null) {
      this.resetStack();
      this._errorHandler.handle(popLayerErr);
    }
  };

  public hgHandlerShowNestedSidebar = () => {
    const pushLayerErr = this.pushLayer(this._layerIds.NestedSidebar);
    if (pushLayerErr !== null) {
      this.resetStack();
      this._errorHandler.handle(pushLayerErr);
    }
  };

  public hgHandlerCloseNestedSidebar = () => {
    const popLayerErr = this.popLayer(this._layerIds.NestedSidebar);
    if (popLayerErr !== null) {
      this.resetStack();
      this._errorHandler.handle(popLayerErr);
    }
  };

  private handleEsc = (e: KeyboardEvent) => {
    if (e.key !== constants.keyCodes.Escape) {
      return;
    }

    if (this._layerStack.length > 1) {
      this.popLayer();
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
    const nextElementToFocus = parseHtmlElement(
      newActiveLayer.element.querySelector(
        attributeSelector(constants.layers.contentShape.initialFocusAttribute),
      ),
    );
    if (!nextElementToFocus) {
      return new NullChildErr(
        newActiveLayer.element,
        attributeSelector(constants.layers.contentShape.initialFocusAttribute),
      );
    }

    // get the currently active layer
    const getActiveLayerResult = this.getActiveLayer();
    if (!getActiveLayerResult.ok) {
      return getActiveLayerResult.error;
    }
    const activeLayer = getActiveLayerResult.value;

    // If any element has focus, then blur it
    const currentFocussedElement = parseHtmlElement(document.activeElement);
    currentFocussedElement?.blur();

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
    const nextElementToFocus = parseHtmlElement(
      nextActiveLayer.element.querySelector(
        attributeSelector(constants.layers.contentShape.initialFocusAttribute),
      ),
    );
    if (!nextElementToFocus) {
      return new WindowManagerErr(
        `Unable to pop layer ${layerId}. ${nextActiveLayer.layerId} must ` +
          `contain an child element with ` +
          `the '${constants.layers.contentShape.initialFocusAttribute}' attribute`,
      );
    }

    // If any element has focus, then blur it
    const currentFocussedElement = parseHtmlElement(document.activeElement);
    currentFocussedElement?.blur();

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

  private notifyPush(layerId: string) {
    switch (layerId) {
      case this._layerIds.Sidebar:
        document.dispatchEvent(this.openedSidebarEvent);
        break;
      case this._layerIds.NestedSidebar:
        document.dispatchEvent(this.openedNestedSidebarEvent);
        break;
      default:
        return;
    }
  }

  private notifyPop(layerId: string) {
    switch (layerId) {
      case this._layerIds.Sidebar:
        document.dispatchEvent(this.closedSidebarEvent);
        break;
      case this._layerIds.NestedSidebar:
        document.dispatchEvent(this.closedNestedSidebarEvent);
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

  private getActiveLayer(): Result<Layer> {
    return this.getLayerAtIndex(this._layerStack.length - 1);
  }

  private getLayer = (layerId: string): Result<Layer> => {
    const layerEl = parseHtmlElement(document.getElementById(layerId));
    if (!layerEl) {
      return Fail(new NoElWithIdErr(layerId));
    }

    return Success({ layerId: layerId, element: layerEl, stackIndex: -1 });
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
