import { ErrorHandler } from "./errorHandler";
import { InitialiseToast } from "./toast";
import { MainTable } from "./mainTable";
import * as htmx from "./htmx";
import { SidebarManager } from "./sidebarManager";
import { WindowManager } from "./windowManager";
import { SortableListManager } from "./sortableLists";

export interface HgNamespace {
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

  public readonly errorHandler: ErrorHandler | undefined;
  public readonly windowManager: WindowManager | undefined;
  public readonly sidebarManager: SidebarManager | undefined;
  public readonly sortableListManager: SortableListManager | undefined;
  public readonly mainTable: MainTable | undefined;

  public constructor() {
    htmx.configureCacheSize(0);

    InitialiseToast();
    this.errorHandler = new ErrorHandler();

    // panic recovery - window manager should not throw unless there is a
    // fundamental problem with the layout on load
    try {
      this.windowManager = new WindowManager(this.errorHandler);
    } catch (e: unknown) {
      this.errorHandler.handle(e);
      return;
    }

    this.mainTable = new MainTable(this.errorHandler, this.windowManager);

    this.sidebarManager = new SidebarManager(
      this.errorHandler,
      this.windowManager,
    );

    this.sortableListManager = new SortableListManager(this.errorHandler);

    this.hgNamespace = {
      showSidebar: this.windowManager.hgHandlerShowSidebar,
      closeSidebar: this.windowManager.hgHandlerCloseSidebar,
      showNestedSidebar: this.windowManager.hgHandlerShowNestedSidebar,
      closeNestedSidebar: this.windowManager.hgHandlerCloseNestedSidebar,
      editColumns: {
        addColumns: this.mainTable.hgHandlerAddColumns,
        removeColumn: this.mainTable.hgHandlerRemoveColumn,
      },
    };
  }
}
