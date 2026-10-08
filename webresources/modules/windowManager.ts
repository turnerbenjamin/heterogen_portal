import * as utils from "./utils";
import * as err from "./errorHandling";

interface Layer {
  layerId: string;
  element: HTMLElement;
  stackIndex: number;
}

export const openedSidebarEventId = "openedsidebar";
export const closedSidebarEventId = "closedsidebar";
export const openedNestedSidebarEventId = "openednestedsidebar";
export const closedNestedSidebarEventId = "closednestedsidebar";

export class WindowManager {
  private openedSidebarEvent = new CustomEvent<Event>(openedSidebarEventId);
  private closedSidebarEvent = new CustomEvent<Event>(closedSidebarEventId);
  private openedNestedSidebarEvent = new CustomEvent<Event>(
    openedNestedSidebarEventId,
  );
  private closedNestedSidebarEvent = new CustomEvent<Event>(
    closedNestedSidebarEventId,
  );

  // layer ids
  private readonly _layerIds = {
    App: "app",
    Sidebar: "sidebar",
    NestedSidebar: "nested-sidebar",
    Alert: "alert",
  } as const;

  private readonly _layers: Record<string, Layer> = {};
  private _layerStack: Layer[] = [];
  public isInitialised = false;

  public init(): Error | null {
    if (this.isInitialised) {
      return null;
    }

    for (const layerId of Object.values(this._layerIds)) {
      const [layer, getLayerErr] = this.getLayer(layerId);
      if (!layer || getLayerErr !== null) {
        return getLayerErr ?? err.NullErr("layer");
      }
      this._layers[layerId] = layer;
      if (layerId !== this._layerIds.App) {
        layer.element.setAttribute("inert", "");
      }
    }

    const appLayer = this._layers[this._layerIds.App];
    appLayer.stackIndex = 0;
    this._layerStack = [appLayer];

    document.addEventListener("keydown", this.handleEsc);

    this.isInitialised = true;
    return null;
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
      err.handle(pushLayerErr);
    }
  };

  public hgHandlerCloseSidebar = () => {
    const popLayerErr = this.popLayer(this._layerIds.Sidebar);
    if (popLayerErr !== null) {
      this.resetStack();
      err.handle(popLayerErr);
    }
  };

  public hgHandlerShowNestedSidebar = () => {
    const pushLayerErr = this.pushLayer(this._layerIds.NestedSidebar);
    if (pushLayerErr !== null) {
      this.resetStack();
      err.handle(pushLayerErr);
    }
  };

  public hgHandlerCloseNestedSidebar = () => {
    const popLayerErr = this.popLayer(this._layerIds.NestedSidebar);
    if (popLayerErr !== null) {
      this.resetStack();
      err.handle(popLayerErr);
    }
  };

  private handleEsc = (e: KeyboardEvent) => {
    if (e.key !== "Escape") {
      return;
    }

    if (this._layerStack.length > 1) {
      this.popLayer();
    }
  };

  private pushLayer = (layerId: string): Error | null => {
    const newActiveLayer = this._layers[layerId];
    if (newActiveLayer.stackIndex !== -1) {
      return err.Unexpected(
        "invalid attempt to activate a layer already in the layer stack",
      );
    }

    // Get next element to focus
    const nextElementToFocus = utils.ParseHtmlElement(
      newActiveLayer.element.querySelector(`[initial-focus]`),
    );
    if (!nextElementToFocus) {
      return err.Unexpected(
        `unable to push layer: ${layerId} should contain an element with an ` +
          "initial focus selector",
      );
    }

    // get the currently active layer
    const [activeLayer, getActiveLayerErr] = this.getActiveLayer();
    if (!activeLayer || getActiveLayerErr !== null) {
      return getActiveLayerErr ?? err.NullErr("active layer");
    }

    // If any element has focus, then blur it
    const currentFocussedElement = utils.ParseHtmlElement(
      document.activeElement,
    );
    currentFocussedElement?.blur();

    // Set the current active layer to inert
    activeLayer.element.setAttribute("inert", "");

    // Set the new active layer to active and push it to the stack
    newActiveLayer.element.removeAttribute("inert");
    newActiveLayer.stackIndex = this._layerStack.length;
    this._layerStack.push(newActiveLayer);

    this.notifyPush(newActiveLayer.layerId);
    nextElementToFocus.focus();

    return null;
  };

  private popLayer = (layerId?: string): Error | null => {
    // get current active layer
    const [currentLayer, getActiveLayerErr] = this.getActiveLayer();
    if (!currentLayer || getActiveLayerErr !== null) {
      return getActiveLayerErr ?? err.NullErr("active layer");
    }

    if (layerId && currentLayer.layerId !== layerId) {
      return err.Unexpected(
        `unable to pop layer: ${layerId} is not the active layer`,
      );
    }

    // get new active layer
    const [newActiveLayer, getNewActiveLayerErr] = this.getLayerAtIndex(
      this._layerStack.length - 2,
    );
    if (!newActiveLayer || getNewActiveLayerErr !== null) {
      return getActiveLayerErr ?? err.NullErr("new active layer");
    }

    // Get next element to focus
    const nextElementToFocus = utils.ParseHtmlElement(
      newActiveLayer.element.querySelector(`[initial-focus]`),
    );
    if (!nextElementToFocus) {
      return err.Unexpected(
        `unable to push layer: ${newActiveLayer.layerId} should contain an ` +
          "element with an initial focus selector",
      );
    }

    // If any element has focus, then blur it
    const currentFocussedElement = utils.ParseHtmlElement(
      document.activeElement,
    );
    currentFocussedElement?.blur();

    // pop the current layer and set it as inert
    this._layerStack.pop();
    currentLayer.stackIndex = -1;
    currentLayer.element.setAttribute("inert", "");

    // make the new active layer active
    newActiveLayer.element.removeAttribute("inert");
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

  private getLayerAtIndex(n: number): [Layer | null, Error | null] {
    if (this._layerStack.length === 0) {
      return [
        null,
        err.Unexpected("unable to retrieve layer: layer stack is empty"),
      ];
    }

    if (n >= this._layerStack.length) {
      return [
        null,
        err.Unexpected("unable to get layer: index is out of bounds"),
      ];
    }

    return [this._layerStack[n], null];
  }

  private getActiveLayer(): [Layer | null, Error | null] {
    return this.getLayerAtIndex(this._layerStack.length - 1);
  }

  private getLayer = (layerId: string): [Layer | null, Error | null] => {
    const layerEl = utils.ParseHtmlElement(document.getElementById(layerId));
    if (!layerEl) {
      return [null, err.NullErr(layerId)];
    }

    return [{ layerId: layerId, element: layerEl, stackIndex: -1 }, null];
  };

  private resetStack = () => {
    const appLayer = this._layers[this._layerIds.App];
    appLayer.stackIndex = 0;
    appLayer.element.removeAttribute("inert");
    this._layerStack = [appLayer];

    for (const layer of Object.values(this._layers)) {
      if (layer.layerId === this._layerIds.App) {
        continue;
      }

      layer.stackIndex = -1;
      layer.element.setAttribute("inert", "");
      layer.element.classList.add("hidden");
      layer.element.classList.remove("show");
      layer.element.classList.remove("close");
    }
  };
}
