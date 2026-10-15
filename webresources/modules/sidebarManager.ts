import { AlertManager } from "./alertManager";
import constants from "./constants";
import { NullTargetErr, UnexpectedErr } from "./errors";
import * as utils from "./utils";
import { WindowManager } from "./windowManager";

export class SidebarManager {
  private readonly _windowManager: WindowManager;
  private readonly _alertManager: AlertManager;

  public constructor(windowManager: WindowManager, alertManager: AlertManager) {
    this._windowManager = windowManager;
    this._alertManager = alertManager;

    document.addEventListener("animationend", this.handleSidebarTransitionEnd);
  }

  public showSidebar = (contentIdentifier: string): Error | null => {
    const err = this.showSidebarElement(
      constants.layers.sidebar.id,
      contentIdentifier,
    );
    if (err) {
      return err;
    }
    this._windowManager.pushSidebar();
    return null;
  };

  public showNestedSidebar = (contentIdentifier: string): Error | null => {
    const err = this.showSidebarElement(
      constants.layers.nestedSidebar.id,
      contentIdentifier,
    );
    if (err) {
      return err;
    }
    this._windowManager.pushNestedSidebar();
    return null;
  };

  private showSidebarElement = (
    sidebarId: string,
    contentIdentifier: string,
  ): Error | null => {
    const sidebarElRes = utils.getElementById(sidebarId);
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

    let didFindContentElement = false;
    for (const e of sidebarEl.querySelectorAll(
      utils.classSelector(constants.sidebarCommon.contentGroupClass),
    )) {
      const elRes = utils.parseHtmlElement(e);
      if (!elRes.ok) {
        return elRes.error;
      }

      if (elRes.value.id === contentIdentifier) {
        didFindContentElement = true;
      }

      utils.setVisibility(elRes.value, elRes.value.id === contentIdentifier);
    }

    if (!didFindContentElement) {
      return new UnexpectedErr(
        `unable to show sidebar element: '${contentIdentifier} cannot be found'`,
      );
    }

    // remove close and hidden classes
    sidebarEl.classList.add(constants.layers.modifierClasses.show);
    sidebarEl.classList.remove(constants.globalModifiers.hidden);

    return null;
  };

  public closeSidebar = () => {
    this.hideSidebarEl(constants.layers.sidebar.id);
    this._windowManager.popSidebar();
  };

  public closeNestedSidebar = () => {
    this.hideSidebarEl(constants.layers.nestedSidebar.id);
    this._windowManager.popNestedSidebar();
  };

  public hideSidebarEl = (sidebarId: string) => {
    const sideBarElRes = utils.getElementById(sidebarId);
    if (!sideBarElRes.ok) {
      return this._alertManager.renderError(sideBarElRes.error);
    }
    const sidebarEl = sideBarElRes.value;

    // return early if element is in transition
    if (
      sidebarEl.classList.contains(constants.layers.modifierClasses.show) ||
      sidebarEl.classList.contains(constants.layers.modifierClasses.close)
    ) {
      return;
    }

    // remove close and hidden classes
    sidebarEl.classList.add(constants.layers.modifierClasses.close);
  };

  private handleSidebarTransitionEnd = (e: AnimationEvent) => {
    if (
      e.animationName !== constants.sidebarCommon.animations.showSidebar &&
      e.animationName !== constants.sidebarCommon.animations.closeSidebar
    ) {
      return;
    }

    const sidebarRes = utils.parseHtmlElement(e.target);
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
      sidebar.classList.add(constants.globalModifiers.hidden);
    }
  };
}
