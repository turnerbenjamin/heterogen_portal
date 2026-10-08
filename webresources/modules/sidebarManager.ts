import constants from "./constants";
import { ErrorHandler, NullTargetErr } from "./errorHandler";
import * as utils from "./utils";
import { WindowManager } from "./windowManager";

export class SidebarManager {
  private readonly _errHandler: ErrorHandler;
  private readonly _windowManger: WindowManager;

  public constructor(errorHandler: ErrorHandler, windowManager: WindowManager) {
    this._errHandler = errorHandler;
    this._windowManger = windowManager;

    document.addEventListener(constants.layers.sidebar.events.opened, () =>
      this.showSidebar(this._windowManger.GetSidebar()),
    );

    document.addEventListener(constants.layers.sidebar.events.closed, () =>
      this.hideSidebar(this._windowManger.GetSidebar()),
    );

    document.addEventListener(
      constants.layers.nestedSidebar.events.opened,
      () => this.showSidebar(this._windowManger.GetNestedSidebar()),
    );

    document.addEventListener(
      constants.layers.nestedSidebar.events.closed,
      () => this.hideSidebar(this._windowManger.GetNestedSidebar()),
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
      this._errHandler.handle(new NullTargetErr(e));
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
