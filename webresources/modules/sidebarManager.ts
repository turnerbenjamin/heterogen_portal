import { AlertManager } from "./alertManager";
import constants from "./constants";
import { NullTargetErr, UnexpectedErr } from "./errors";
import { onHtmxAfterSwapEvent } from "./htmx";
import {
  elString,
  getChildBySelector,
  getElementById,
  parseHtmlElement,
  setBoolAttribute,
  setVisibility,
} from "./utils";
import { Hook, WindowManager } from "./windowManager";

export class SidebarManager {
  private readonly _windowManager: WindowManager;
  private readonly _alertManager: AlertManager;

  private activeSidebarId: string | null = null;
  private activeNestedSidebarId: string | null = null;

  private onEscapeHooks: Partial<Record<string, Array<() => void>>> = {};

  public constructor(windowManager: WindowManager, alertManager: AlertManager) {
    this._windowManager = windowManager;
    this._alertManager = alertManager;

    document.addEventListener("animationend", this.handleSidebarTransitionEnd);
    onHtmxAfterSwapEvent(this.updateSidebarVisibility);
    this.updateSidebarVisibility();

    this._windowManager.registerHook(Hook.escSidebar, this.handleEscapeSidebar);
    this._windowManager.registerHook(
      Hook.escNestedSidebar,
      this.handleEscapeNestedSidebar,
    );
  }

  public registerOnEscapeHook(sidebarId: string, callback: () => void) {
    if (this.onEscapeHooks[sidebarId] === undefined) {
      this.onEscapeHooks[sidebarId] = [];
    }
    this.onEscapeHooks[sidebarId].push(callback);
  }

  public showSidebar = (sidebarId: string): Error | null => {
    this.activeSidebarId = sidebarId;
    const err = this.showSidebarElement(constants.layers.sidebar.id, sidebarId);
    if (err) {
      return err;
    }
    return null;
  };

  public showNestedSidebar = (sidebarId: string): Error | null => {
    const err = this.showSidebarElement(
      constants.layers.nestedSidebar.id,
      sidebarId,
    );
    if (err) {
      return err;
    }

    this.activeNestedSidebarId = sidebarId;

    return null;
  };

  public closeSidebar = () => {
    const err = this.hideSidebarEl(this.activeSidebarId);
    if (err !== null) {
      this._alertManager.renderError(err);
      return;
    }
  };

  public closeNestedSidebar = () => {
    const err = this.hideSidebarEl(this.activeNestedSidebarId);
    if (err !== null) {
      this._alertManager.renderError(err);
      return;
    }
  };

  private showSidebarElement = (
    layerId: string,
    sidebarId: string,
  ): Error | null => {
    const sidebarElRes = getElementById(sidebarId);
    if (!sidebarElRes.ok) {
      return sidebarElRes.error;
    }
    const sidebarEl = sidebarElRes.value;

    if (
      sidebarEl.classList.contains(constants.layers.modifierClasses.show) ||
      sidebarEl.classList.contains(constants.layers.modifierClasses.close)
    ) {
      // return early if element is in transition
      return null;
    }

    // remove close and hidden classes
    setVisibility(sidebarEl, true);
    setBoolAttribute(sidebarEl, "inert", true);

    sidebarEl.classList.add(constants.layers.modifierClasses.show);

    const err = this.setInitialFocusAttribute(sidebarEl);
    if (err !== null) {
      return err;
    }

    switch (layerId) {
      case constants.layers.sidebar.id:
        this._windowManager.pushSidebar();
        break;
      case constants.layers.nestedSidebar.id:
        this._windowManager.pushNestedSidebar();
        break;
      default:
        return new UnexpectedErr(
          `unable to show sidebar el: unexpected layer id '${layerId}'`,
        );
    }

    return null;
  };

  public hideSidebarEl = (activeSidebarId: string | null): Error | null => {
    if (activeSidebarId === null) {
      return new UnexpectedErr(
        `unable to hide sidebar: active sidebar is null`,
      );
    }
    const activeSidebarRes = getElementById(activeSidebarId);
    if (!activeSidebarRes.ok) {
      return activeSidebarRes.error;
    }
    const activeSidebar = activeSidebarRes.value;

    // return early if element is in transition
    if (
      activeSidebar.classList.contains(constants.layers.modifierClasses.show) ||
      activeSidebar.classList.contains(constants.layers.modifierClasses.close)
    ) {
      return null;
    }

    setBoolAttribute(activeSidebar, "inert", true);
    activeSidebar.classList.add(constants.layers.modifierClasses.close);

    return null;
  };

  private handleSidebarTransitionEnd = (e: AnimationEvent) => {
    if (
      e.animationName !== constants.sidebarCommon.animations.showSidebar &&
      e.animationName !== constants.sidebarCommon.animations.closeSidebar
    ) {
      return;
    }

    const sidebarRes = parseHtmlElement(e.target);
    if (!sidebarRes.ok) {
      this._alertManager.renderError(new NullTargetErr(e));
      return;
    }
    const sidebar = sidebarRes.value;

    if (e.animationName === constants.sidebarCommon.animations.showSidebar) {
      sidebar.classList.remove(constants.layers.modifierClasses.show);

      // sidebar is made inert during transition
      setBoolAttribute(sidebar, "inert", false);
      const parentLayer = sidebarRes.value.parentElement;
      const err = this.setInitialFocus(parentLayer);
      if (err !== null) {
        this._alertManager.renderError(err);
        return;
      }
    }

    if (e.animationName === constants.sidebarCommon.animations.closeSidebar) {
      sidebar.classList.remove(constants.layers.modifierClasses.close);

      const parentLayer = sidebarRes.value.parentElement;

      switch (parentLayer?.id) {
        case constants.layers.sidebar.id:
          this._windowManager.popSidebar();
          this.activeSidebarId = null;
          break;

        case constants.layers.nestedSidebar.id:
          this._windowManager.popNestedSidebar();
          this.activeNestedSidebarId = null;
          break;

        default:
          this._alertManager.renderError(
            new UnexpectedErr(
              `unable to pop sidebar layer: invalid parent id: '${parentLayer?.id}'`,
            ),
          );
          return;
      }

      setVisibility(sidebar, false);
    }
  };

  private updateSidebarVisibility = () => {
    let err = this.updateSidebarVisibilityForLayer(
      constants.layers.sidebar.id,
      this.activeSidebarId,
    );
    if (err !== null) {
      this._alertManager.renderError(err);
      return;
    }

    err = this.updateSidebarVisibilityForLayer(
      constants.layers.nestedSidebar.id,
      this.activeNestedSidebarId,
    );
    if (err !== null) {
      this._alertManager.renderError(err);
      return;
    }
  };

  private updateSidebarVisibilityForLayer = (
    layerId: string,
    activeElId: string | null,
  ): Error | null => {
    const sidebarLayerRes = getElementById(layerId);
    if (!sidebarLayerRes.ok) {
      return sidebarLayerRes.error;
    }

    let activeSidebarFound = false;
    for (const child of sidebarLayerRes.value.children) {
      const isActive = activeElId !== null && child.id === activeElId;

      if (isActive) {
        activeSidebarFound = true;
      }
      setVisibility(child, isActive);
      setBoolAttribute(child, "inert", !isActive);
    }

    if (activeElId !== null && !activeSidebarFound) {
      return new UnexpectedErr(
        `unable to set sidebar visibility: active sidebar, '${activeElId}' not found`,
      );
    }

    return null;
  };

  private setInitialFocusAttribute(sidebarElement: HTMLElement): Error | null {
    const parentLayer = sidebarElement.parentElement;
    if (parentLayer === null) {
      return new UnexpectedErr(
        `unable to set initial focus attribute: ${elString(sidebarElement)} parent is null`,
      );
    }

    const current = getChildBySelector(parentLayer, "[initial-focus]");
    if (current.ok) {
      setBoolAttribute(current.value, "initial-focus", false);
    }

    const nextActiveEl = getChildBySelector(
      sidebarElement,
      "[initial-focus-sidebar]",
    );
    if (!nextActiveEl.ok) {
      return nextActiveEl.error;
    }
    setBoolAttribute(nextActiveEl.value, "initial-focus", true);

    return null;
  }

  private setInitialFocus(parentLayer: HTMLElement | null): Error | null {
    if (parentLayer === null) {
      return new UnexpectedErr(
        `unable to set initial focus: parent layer is null`,
      );
    }

    const current = getChildBySelector(parentLayer, "[initial-focus]");
    if (!current.ok) {
      return current.error;
    }

    current.value.focus();
    return null;
  }

  private handleEscapeSidebar = () => {
    if (this.activeSidebarId === null) {
      return this._alertManager.renderError(
        new UnexpectedErr(
          `unable to handle escape sidebar: active sidebar id is null`,
        ),
      );
    }

    const activeSidebar = getElementById(this.activeSidebarId);
    if (!activeSidebar.ok) {
      return this._alertManager.renderError(activeSidebar.error);
    }

    for (const cb of this.onEscapeHooks[this.activeSidebarId] ?? []) {
      cb();
    }
  };

  private handleEscapeNestedSidebar = () => {
    if (this.activeNestedSidebarId === null) {
      return this._alertManager.renderError(
        new UnexpectedErr(
          `unable to handle escape nested sidebar: active sidebar id is null`,
        ),
      );
    }

    for (const cb of this.onEscapeHooks[this.activeNestedSidebarId] ?? []) {
      cb();
    }
  };
}
