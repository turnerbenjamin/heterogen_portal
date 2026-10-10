import { htmxErrorHandler } from "./htmxErrorHandler";
import { InitialiseToast } from "./toasts";
import { MainTable } from "./mainTable";
import * as htmx from "./htmx";
import { SidebarManager } from "./sidebarManager";
import { lastDitchHandler, WindowManager } from "./windowManager";
import { SortableListManager } from "./sortableListManager";

export interface HgNamespace {
  closeAlert: () => void;
  showSidebar: () => void;
  closeSidebar: () => void;
  showNestedSidebar: () => void;
  closeNestedSidebar: () => void;
  editColumns: {
    addColumns: () => void;
    removeColumn: (_e: MouseEvent) => void;
  };
}

export class App {
  public hgNamespace: HgNamespace | undefined;

  public readonly htmxErrorHandler: htmxErrorHandler | undefined;
  public readonly windowManager: WindowManager | undefined;
  public readonly sidebarManager: SidebarManager | undefined;
  public readonly sortableListManager: SortableListManager | undefined;
  public readonly mainTable: MainTable | undefined;

  public constructor() {
    try {
      htmx.configureCacheSize(0);

      InitialiseToast();
      this.windowManager = new WindowManager();

      this.htmxErrorHandler = new htmxErrorHandler();

      this.mainTable = new MainTable(this.windowManager);
      this.sidebarManager = new SidebarManager(this.windowManager);
      this.sortableListManager = new SortableListManager(this.windowManager);

      // Global try/catch
      addEventListener("error", (e: ErrorEvent) => {
        e.preventDefault();
        this.windowManager?.displayError(e.error);
      });

      this.hgNamespace = {
        closeAlert: this.windowManager.hgHandlerCloseAlert,
        showSidebar: this.windowManager.hgHandlerShowSidebar,
        closeSidebar: this.windowManager.hgHandlerCloseSidebar,
        showNestedSidebar: this.windowManager.hgHandlerShowNestedSidebar,
        closeNestedSidebar: this.windowManager.hgHandlerCloseNestedSidebar,
        editColumns: {
          addColumns: this.mainTable.hgHandlerAddColumns,
          removeColumn: this.mainTable.hgHandlerRemoveColumn,
        },
      };
    } catch (err: unknown) {
      lastDitchHandler(err);
    }
  }
}
