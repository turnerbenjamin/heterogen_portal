import * as utils from "./utils";
import * as err from "./errorHandling";
import {
  closedNestedSidebarEventId,
  closedSidebarEventId,
  openedNestedSidebarEventId,
  openedSidebarEventId,
  WindowManager,
} from "./windowManager";

export class SidebarManager {
  private readonly _windowManger: WindowManager;

  public constructor(windowManager: WindowManager) {
    this._windowManger = windowManager;

    document.addEventListener(openedSidebarEventId, () =>
      this.showSidebar(this._windowManger.GetSidebar()),
    );

    document.addEventListener(closedSidebarEventId, () =>
      this.hideSidebar(this._windowManger.GetSidebar()),
    );

    document.addEventListener(openedNestedSidebarEventId, () =>
      this.showSidebar(this._windowManger.GetNestedSidebar()),
    );

    document.addEventListener(closedNestedSidebarEventId, () =>
      this.hideSidebar(this._windowManger.GetNestedSidebar()),
    );

    document.addEventListener("animationend", this.handleSidebarTransitionEnd);
  }

  private showSidebar = (sideBar: HTMLElement) => {
    // return early if element is in transition
    if (
      sideBar.classList.contains("show") ||
      sideBar.classList.contains("close")
    ) {
      return;
    }

    // remove close and hidden classes
    sideBar.classList.add("show");
    sideBar.classList.remove("hidden");
  };

  private hideSidebar = (sideBar: HTMLElement) => {
    // return early if element is in transition
    if (
      sideBar.classList.contains("show") ||
      sideBar.classList.contains("close")
    ) {
      return;
    }

    // remove close and hidden classes
    sideBar.classList.add("close");
  };

  private handleSidebarTransitionEnd = (e: AnimationEvent) => {
    if (
      e.animationName !== "show-sidebar" &&
      e.animationName !== "close-sidebar"
    ) {
      return;
    }

    const sidebar = utils.ParseHtmlElement(e.target);
    if (sidebar === null) {
      err.handle(
        err.Unexpected("unable to access sidebar from animation event"),
      );
      return;
    }

    if (e.animationName === "show-sidebar") {
      sidebar.classList.remove("show");
    }

    if (e.animationName === "close-sidebar") {
      sidebar.classList.remove("close");
      sidebar.classList.add("hidden");
    }
  };
}
