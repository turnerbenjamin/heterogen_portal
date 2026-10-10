import constants from "./constants";
import { NullTargetErr } from "./errors";
import * as utils from "./utils";
import { WindowManager } from "./windowManager";

export class SidebarManager {
  private readonly _windowManager: WindowManager;

  public constructor(windowManager: WindowManager) {
    this._windowManager = windowManager;

    document.addEventListener(constants.layers.sidebar.events.opened, () =>
      this.showSidebar(this._windowManager.GetSidebar()),
    );

    document.addEventListener(constants.layers.sidebar.events.closed, () =>
      this.hideSidebar(this._windowManager.GetSidebar()),
    );

    document.addEventListener(
      constants.layers.nestedSidebar.events.opened,
      () => this.showSidebar(this._windowManager.GetNestedSidebar()),
    );

    document.addEventListener(
      constants.layers.nestedSidebar.events.closed,
      () => this.hideSidebar(this._windowManager.GetNestedSidebar()),
    );

    document.addEventListener("animationend", this.handleSidebarTransitionEnd);
  }

  private showSidebar = (sideBar: HTMLElement) => {
    // return early if element is in transition
    if (
      sideBar.classList.contains(constants.layers.modifierClasses.show) ||
      sideBar.classList.contains(constants.layers.modifierClasses.close)
    ) {
      return;
    }

    // remove close and hidden classes
    sideBar.classList.add(constants.layers.modifierClasses.show);
    sideBar.classList.remove(constants.globalModifiers.hidden);
  };

  private hideSidebar = (sideBar: HTMLElement) => {
    // return early if element is in transition
    if (
      sideBar.classList.contains(constants.layers.modifierClasses.show) ||
      sideBar.classList.contains(constants.layers.modifierClasses.close)
    ) {
      return;
    }

    // remove close and hidden classes
    sideBar.classList.add(constants.layers.modifierClasses.close);
  };

  private handleSidebarTransitionEnd = (e: AnimationEvent) => {
    if (
      e.animationName !== constants.sidebarCommon.animations.showSidebar &&
      e.animationName !== constants.sidebarCommon.animations.closeSidebar
    ) {
      return;
    }

    const sidebar = utils.parseHtmlElement(e.target);
    if (sidebar === null) {
      this._windowManager.displayError(new NullTargetErr(e));
      return;
    }

    if (e.animationName === constants.sidebarCommon.animations.showSidebar) {
      sidebar.classList.remove(constants.layers.modifierClasses.show);
    }

    if (e.animationName === constants.sidebarCommon.animations.closeSidebar) {
      sidebar.classList.remove(constants.layers.modifierClasses.close);
      sidebar.classList.add(constants.globalModifiers.hidden);
    }
  };
}
