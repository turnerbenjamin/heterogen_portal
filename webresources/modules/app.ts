import { htmxErrorHandler } from "./htmxErrorHandler";
import { InitialiseToast } from "./toasts";
import { MainTable } from "./mainTable";
import { SidebarManager } from "./sidebarManager";
import { WindowManager } from "./windowManager";
import { SortableListManager } from "./sortableListManager";
import { AlertManager } from "./alertManager";
import { parseAppError } from "./errors";
import { TableFilterManager } from "./tableFilterManager";

export interface HgNamespace {
  closeAlert: () => void;
  editColumns: {
    addColumns: () => void;
    removeColumn: (_e: MouseEvent) => void;
    showAddColumnsSidebar: () => void;
    closeAddColumnsSidebar: () => void;
  };
  mainTable: {
    showEditColumnsSidebar: () => void;
    showEditFiltersSidebar: () => void;
    closeEditColumnsSidebar: () => void;
    closeEditFiltersSidebar: () => void;
    filters: {
      newRule: (_e: MouseEvent) => void;
      newRuleGroup: (_e: MouseEvent) => void;
      deleteRule: (_e: MouseEvent) => void;
      deleteRuleGroup: (_e: MouseEvent) => void;
    };
  };
}

export class App {
  public hgNamespace: HgNamespace | undefined;

  public readonly htmxErrorHandler: htmxErrorHandler | undefined;
  public readonly alertManager: AlertManager | undefined;
  public readonly windowManager: WindowManager | undefined;
  public readonly sidebarManager: SidebarManager | undefined;
  public readonly sortableListManager: SortableListManager | undefined;
  public readonly mainTable: MainTable | undefined;
  public readonly tableFilterManager: TableFilterManager | undefined;

  public constructor() {
    try {
      InitialiseToast();
      this.htmxErrorHandler = new htmxErrorHandler();

      this.windowManager = new WindowManager(this.panic);

      this.alertManager = new AlertManager(this.windowManager, this.panic);
      this.sidebarManager = new SidebarManager(
        this.windowManager,
        this.alertManager,
      );

      this.mainTable = new MainTable(this.alertManager, this.sidebarManager);
      this.tableFilterManager = new TableFilterManager(this.alertManager);

      this.sortableListManager = new SortableListManager(
        this.windowManager,
        this.alertManager,
      );

      // Global try/catch
      addEventListener("error", (e: ErrorEvent) => {
        e.preventDefault();
        this.alertManager?.renderError(e.error);
      });

      this.hgNamespace = {
        closeAlert: this.alertManager.hgHandlerCloseAlert,
        editColumns: {
          addColumns: this.mainTable.hgHandlerAddColumns,
          removeColumn: this.mainTable.hgHandlerRemoveColumn,
          showAddColumnsSidebar: this.mainTable.hgHandlerOpenAddColumnsSidebar,
          closeAddColumnsSidebar:
            this.mainTable.hgHandlerCloseAddColumnsSidebar,
        },
        mainTable: {
          showEditColumnsSidebar:
            this.mainTable.hgHandlerShowEditColumnsSidebar,
          closeEditColumnsSidebar:
            this.mainTable.hgHandlerCloseEditColumnsSidebar,
          showEditFiltersSidebar:
            this.mainTable.hgHandlerShowEditFiltersSidebar,
          closeEditFiltersSidebar:
            this.mainTable.hgHandlerCloseEditColumnsSidebar,
          filters: {
            newRule: this.tableFilterManager.hgHandlerNewRule,
            newRuleGroup: this.tableFilterManager.hgHandlerNewRuleGroup,
            deleteRule: this.tableFilterManager.hgHandlerDeleteRule,
            deleteRuleGroup: this.tableFilterManager.hgHandlerDeleteRuleGroup,
          },
        },
      };
    } catch (err: unknown) {
      this.panic(err);
    }
  }

  private panic(errObj: unknown) {
    const err = parseAppError(errObj);

    const lastDitchErrContainerEl = document.createElement("div");
    lastDitchErrContainerEl.id = "last-ditch-err-container";

    const lastDitchTitleEl = document.createElement("h1");
    lastDitchTitleEl.id = "last-ditch-err-title";
    lastDitchTitleEl.textContent = "Unable to initialise application";

    const lastDitchMsgEl = document.createElement("p");
    lastDitchMsgEl.id = "last-ditch-err-message";
    lastDitchMsgEl.textContent = err.message;

    const lastDitchDetailEl = document.createElement("pre");
    lastDitchDetailEl.id = "last-ditch-err-detail";
    lastDitchDetailEl.classList.add("error-detail");
    lastDitchDetailEl.textContent = err.details;

    lastDitchErrContainerEl.appendChild(lastDitchTitleEl);
    lastDitchErrContainerEl.appendChild(lastDitchMsgEl);
    lastDitchErrContainerEl.appendChild(lastDitchDetailEl);

    document.body.innerHTML = "";
    document.body.appendChild(lastDitchErrContainerEl);
  }
}
