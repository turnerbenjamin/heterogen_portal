import * as err from "./errorHandling";
import * as utils from "./utils";
import { closedNestedSidebarEventId, WindowManager } from "./windowManager";

export class MainTable {
  private _windowManager: WindowManager;

  public constructor(windowManager: WindowManager) {
    this._windowManager = windowManager;

    document.addEventListener(
      closedNestedSidebarEventId,
      this.handleCloseNestedSidebar,
    );
  }

  public hgHandlerAddColumns = () => {
    const editColumnsOptionsSet = document.getElementById("edit-columns-list");
    if (!editColumnsOptionsSet) {
      err.handle(err.NullErr("edit-columns-list"));
      return;
    }
    const editColumnOptionSetItems =
      editColumnsOptionsSet.getElementsByClassName("sortable-item");
    if (!editColumnOptionSetItems || !editColumnOptionSetItems.length) {
      err.handle(
        err.Unexpected("expected at least one item in edit columns list"),
      );
    }
    const protoEditColumn = editColumnOptionSetItems[0];

    const [addColumnOptionSets, getAddColumnSetErr] = this.getAddColumnSet();
    if (getAddColumnSetErr || !addColumnOptionSets) {
      err.handle(getAddColumnSetErr);
      return;
    }

    let uncheckedCount = 0;
    for (const optionSetEl of addColumnOptionSets) {
      const labelEl = optionSetEl.querySelector("label");
      if (!labelEl) {
        err.handle(err.NullErr("label"));
        return;
      }

      const checkboxEl = optionSetEl.querySelector("input");
      if (!checkboxEl || checkboxEl.type !== "checkbox") {
        err.handle(err.NullErr("checkbox"));
        return;
      }

      if (!checkboxEl.checked) {
        uncheckedCount++;
        continue;
      }

      // 1. Reset and hide the column
      checkboxEl.checked = false;
      utils.setVisibility(optionSetEl, false);

      // 2. Add to edit columns
      const newEditItem = protoEditColumn.cloneNode(true) as HTMLElement;
      newEditItem.setAttribute("data-column-path", checkboxEl.value);

      // 3. Set text field for new edit item
      const newItemTextContainer = newEditItem.querySelector(
        ".sortable-item-text",
      );
      if (!newItemTextContainer) {
        err.handle(err.NullErr(".sortable-item-text"));
        return;
      }
      newItemTextContainer.textContent = labelEl.textContent;

      // 4. Append new edit column to edit columns
      editColumnsOptionsSet.appendChild(newEditItem);
    }

    const openAddColumnScreenBtn = document.getElementById(
      "open-add-column-screen-button",
    );
    if (!openAddColumnScreenBtn) {
      err.handle(err.NullErr("#open-add-column-screen-button"));
      return;
    }

    utils.setVisibility(openAddColumnScreenBtn, uncheckedCount > 0);
    this._windowManager.hgHandlerCloseNestedSidebar();
  };

  private handleCloseNestedSidebar = () => {
    if (!this.isMainTablePage) {
      return;
    }

    const addColumnOptionEl = utils.ParseHtmlElement(
      document.getElementById("add-columns-options"),
    );
    if (!addColumnOptionEl) {
      err.handle(err.NullErr("add-columns-options"));
      return;
    }

    const checkboxEls = addColumnOptionEl.getElementsByTagName("input");
    for (const checkboxEl of checkboxEls) {
      checkboxEl.checked = false;
    }
  };

  public hgHandlerRemoveColumn = (e: MouseEvent) => {
    const target = utils.ParseHtmlElement(e.currentTarget);
    if (!target) {
      err.handle(err.NullErr("remove column event target"));
      return;
    }

    const sortableColumn = utils.ParseHtmlElement(
      target.closest(".sortable-item"),
    );
    if (!sortableColumn) {
      return;
    }
    const headerPath = this.getColumnPath(sortableColumn);
    if (!headerPath) {
      err.handle(err.NullErr("column-path attribute"));
      return;
    }

    const sortableList = sortableColumn.closest(".sortable-list");
    if (!sortableList) {
      err.handle(err.NullErr("sortable list container for selected item"));
      return;
    }

    const [addColumnOptionSets, getAddColumnSetErr] = this.getAddColumnSet();
    if (getAddColumnSetErr || !addColumnOptionSets) {
      err.handle(getAddColumnSetErr);
      return;
    }

    // find the column in the add columns set and mark it as visible
    let addColumnElement: HTMLElement | null = null;
    for (const addColumnSet of addColumnOptionSets) {
      const currentEl = utils.ParseHtmlElement(addColumnSet);
      if (!currentEl) {
        err.handle(
          err.Unexpected("unable to parse column set as HTML element"),
        );
        return;
      }

      const addColumnElementPath = this.getColumnPath(currentEl);
      if (addColumnElementPath === headerPath) {
        addColumnElement = currentEl;
        break;
      }
    }

    if (!addColumnElement) {
      err.handle(
        err.Unexpected(`unable to find add column item for ${headerPath}`),
      );
      return;
    }

    // remove element from the edit list and reshow the column in the add list
    utils.setVisibility(addColumnElement, true);
    sortableList.removeChild(sortableColumn);
  };

  public getAddColumnSet = (): [
    HTMLCollectionOf<Element> | null,
    Error | null,
  ] => {
    const addColumnOptionSetEl = document.getElementById("add-columns-options");
    if (!addColumnOptionSetEl) {
      return [null, err.NullErr("add-columns-options")];
    }

    const addColumnOptionSets =
      addColumnOptionSetEl.getElementsByClassName("option-group-item");

    return [addColumnOptionSets, null];
  };

  private getColumnPath = (el: HTMLElement): string => {
    return el.getAttribute("data-column-path") ?? "";
  };

  private isMainTablePage = (): boolean => {
    return document.getElementById("main-table-content") !== null;
  };
}
