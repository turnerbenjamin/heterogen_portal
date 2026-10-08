import * as err from "./errorHandling";
import { InitialiseToast } from "./toast";
import { AppState } from "./state";
import { InitSortableLists } from "./sortableLists";
import { MainTable } from "./mainTable";
import * as htmx from "./htmx";
import { SidebarManager } from "./sidebarManager";
import { WindowManager } from "./windowManager";

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
  public hgNamespace: HgNamespace;

  public readonly appState: AppState;
  public readonly windowManager: WindowManager;
  public readonly sidebarManager: SidebarManager;
  public readonly mainTable: MainTable;

  public constructor() {
    InitialiseToast();
    err.InitialiseErrorHandling();

    htmx.configureCacheSize(0);

    this.appState = new AppState();
    this.windowManager = new WindowManager();
    const initWindowMangerErr = this.windowManager.init();
    if (initWindowMangerErr !== null) {
      err.handle(initWindowMangerErr);
    }

    this.mainTable = new MainTable(this.windowManager);
    this.sidebarManager = new SidebarManager(this.windowManager);

    InitSortableLists(
      this.appState.getDraggedItem,
      this.appState.setDraggedItem,
    );

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
