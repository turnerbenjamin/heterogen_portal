import { AlertManager } from "./alertManager";
import constants from "./constants";
import { NullTargetErr, UnexpectedErr } from "./errors";
import { onHtmxAfterSwapEvent } from "./htmx";
import { getElementById, parseHtmlElement, setVisibility } from "./utils";
import { WindowManager } from "./windowManager";

export class SidebarManager {
  private readonly _windowManager: WindowManager;
  private readonly _alertManager: AlertManager;

  private activeSidebarId: string | null = null;
  private activeNestedSidebarId: string | null = null;

  public constructor(windowManager: WindowManager, alertManager: AlertManager) {
    this._windowManager = windowManager;
    this._alertManager = alertManager;

    document.addEventListener("animationend", this.handleSidebarTransitionEnd);
    onHtmxAfterSwapEvent(this.updateSidebarVisibility);
    this.updateSidebarVisibility();
  }

  public showSidebar = (sidebarId: string): Error | null => {
    const err = this.showSidebarElement(constants.layers.sidebar.id, sidebarId);
    if (err) {
      return err;
    }
    this.activeSidebarId = sidebarId;
    this._windowManager.pushSidebar();
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
    this._windowManager.pushNestedSidebar();
    return null;
  };

  public closeSidebar = () => {
    const err = this.hideSidebarEl(this.activeSidebarId);
    if (err !== null) {
      this._alertManager.renderError(err);
      return;
    }

    this.activeSidebarId = null;
  };

  public closeNestedSidebar = () => {
    const err = this.hideSidebarEl(this.activeNestedSidebarId);
    if (err !== null) {
      this._alertManager.renderError(err);
      return;
    }

    this.activeNestedSidebarId = null;
  };

  private showSidebarElement = (
    _layerId: string,
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
    sidebarEl.classList.add(constants.layers.modifierClasses.show);
    setVisibility(sidebarEl, true);

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
    }

    if (e.animationName === constants.sidebarCommon.animations.closeSidebar) {
      sidebar.classList.remove(constants.layers.modifierClasses.close);

      const parentLayer = sidebarRes.value.parentElement;
      switch (parentLayer?.id) {
        case constants.layers.sidebar.id:
          this._windowManager.popSidebar();
          break;
        case constants.layers.nestedSidebar.id:
          this._windowManager.popNestedSidebar();
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
      setVisibility(child, isActive);

      if (isActive) {
        activeSidebarFound = true;
      }
    }

    if (activeElId !== null && !activeSidebarFound) {
      return new UnexpectedErr(
        `unable to set sidebar visibility: active sidebar, '${activeElId}' not found`,
      );
    }

    return null;
  };
}
