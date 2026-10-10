import constants from "./constants";
import {
  NoElWithIdErr,
  UnexpectedErr,
  QSelectFail,
  NullCurrentTargetErr,
  NullAttributeErr,
  NullParentErr,
} from "./errors";
import { Fail, Result, Success } from "./result";
import { setVisibility, parseHtmlElement, classSelector } from "./utils";
import { WindowManager } from "./windowManager";

export class MainTable {
  private _windowManager: WindowManager;

  public constructor(windowManager: WindowManager) {
    this._windowManager = windowManager;

    document.addEventListener(
      constants.layers.nestedSidebar.events.closed,
      this.handleCloseNestedSidebar,
    );
  }

  public hgHandlerAddColumns = () => {
    const editColumnsOptionsSet = document.getElementById(
      constants.layers.sidebar.content.editColumnsSidebar.editColumnsList.id,
    );
    if (!editColumnsOptionsSet) {
      this._windowManager.displayError(
        new NoElWithIdErr(
          constants.layers.sidebar.content.editColumnsSidebar.editColumnsList
            .id,
        ),
      );
      return;
    }
    const editColumnOptionSetItems =
      editColumnsOptionsSet.getElementsByClassName(
        constants.classes.sortableList.items.className,
      );
    if (!editColumnOptionSetItems || !editColumnOptionSetItems.length) {
      this._windowManager.displayError(
        new UnexpectedErr(
          `expected ${constants.layers.sidebar.content.editColumnsSidebar.editColumnsList.id} ` +
            `to contain at least one child of type ${constants.classes.sortableList.items.className}`,
        ),
      );
    }
    const protoEditColumn = editColumnOptionSetItems[0];

    const getAddColumnSetResult = this.getAddColumnSet();
    if (!getAddColumnSetResult.ok) {
      this._windowManager.displayError(getAddColumnSetResult.error);
      return;
    }

    let uncheckedCount = 0;
    for (const optionSetEl of getAddColumnSetResult.value) {
      const labelEl = optionSetEl.querySelector("label");
      if (!labelEl) {
        this._windowManager.displayError(
          new QSelectFail(
            "label",
            constants.classes.optionGroup.items.className,
          ),
        );
        return;
      }

      const checkboxEl = optionSetEl.querySelector("input");
      if (!checkboxEl || checkboxEl.type !== "checkbox") {
        this._windowManager.displayError(
          new UnexpectedErr(
            `expected ${constants.classes.optionGroup.items.className} to ` +
              `contain an input element of type checkbox`,
          ),
        );
        return;
      }

      if (!checkboxEl.checked) {
        uncheckedCount++;
        continue;
      }

      // 1. Reset and hide the column
      checkboxEl.checked = false;
      setVisibility(optionSetEl, false);

      // 2. Add to edit columns
      const newEditItem = protoEditColumn.cloneNode(true) as HTMLElement;
      newEditItem.setAttribute(
        constants.classes.sortableList.items.attributes.columnPath,
        checkboxEl.value,
      );

      // 3. Set text field for new edit item
      const newItemTextContainer = newEditItem.querySelector(
        classSelector(
          constants.classes.sortableList.items.textContainer.className,
        ),
      );
      if (!newItemTextContainer) {
        this._windowManager.displayError(
          new QSelectFail(
            constants.classes.sortableList.items.textContainer.className,
            constants.classes.sortableList.items.className,
          ),
        );
        return;
      }
      newItemTextContainer.textContent = labelEl.textContent;

      // 4. Append new edit column to edit columns
      editColumnsOptionsSet.appendChild(newEditItem);
    }

    const openAddColumnScreenBtn = document.getElementById(
      constants.layers.sidebar.content.editColumnsSidebar
        .openAddColumnScreenButton.id,
    );
    if (!openAddColumnScreenBtn) {
      this._windowManager.displayError(
        new NoElWithIdErr(
          constants.layers.sidebar.content.editColumnsSidebar
            .openAddColumnScreenButton.id,
        ),
      );
      return;
    }

    setVisibility(openAddColumnScreenBtn, uncheckedCount > 0);
    this._windowManager.hgHandlerCloseNestedSidebar();
  };

  private handleCloseNestedSidebar = () => {
    if (!this.isMainTablePage) {
      return;
    }

    const addColumnOptionEl = parseHtmlElement(
      document.getElementById(
        constants.layers.nestedSidebar.content.addColumnsSidebar
          .addColumnsOptions.id,
      ),
    );
    if (!addColumnOptionEl) {
      this._windowManager.displayError(
        new NoElWithIdErr(
          constants.layers.nestedSidebar.content.addColumnsSidebar
            .addColumnsOptions.id,
        ),
      );
      return;
    }

    const checkboxEls = addColumnOptionEl.getElementsByTagName("input");
    for (const checkboxEl of checkboxEls) {
      checkboxEl.checked = false;
    }
  };

  public hgHandlerRemoveColumn = (e: MouseEvent) => {
    const target = parseHtmlElement(e.currentTarget);
    if (!target) {
      this._windowManager.displayError(new NullCurrentTargetErr(e));
      return;
    }

    const sortableColumn = parseHtmlElement(
      target.closest(
        classSelector(constants.classes.sortableList.items.className),
      ),
    );
    if (!sortableColumn) {
      return;
    }
    const headerPath = sortableColumn.getAttribute(
      constants.classes.sortableList.items.attributes.columnPath,
    );
    if (!headerPath) {
      this._windowManager.displayError(
        new NullAttributeErr(
          sortableColumn,
          constants.classes.sortableList.items.attributes.columnPath,
        ),
      );
      return;
    }

    const sortableList = sortableColumn.closest(
      classSelector(constants.classes.sortableList.className),
    );
    if (!sortableList) {
      this._windowManager.displayError(
        new NullParentErr(
          sortableColumn,
          constants.classes.sortableList.className,
        ),
      );
      return;
    }

    const addColumnSetResult = this.getAddColumnSet();
    if (!addColumnSetResult.ok) {
      this._windowManager.displayError(addColumnSetResult.error);
      return;
    }

    // find the column in the add columns set and mark it as visible
    let addColumnElement: HTMLElement | null = null;
    for (const currentEl of addColumnSetResult.value) {
      const addColumnElementPath = currentEl.getAttribute(
        constants.classes.optionGroup.items.attributes.columnPath,
      );
      if (!addColumnElementPath) {
        this._windowManager.displayError(
          new NullAttributeErr(
            currentEl,
            constants.classes.optionGroup.items.attributes.columnPath,
          ),
        );
        return;
      }

      if (addColumnElementPath === headerPath) {
        addColumnElement = currentEl;
        break;
      }
    }

    if (!addColumnElement) {
      this._windowManager.displayError(
        new UnexpectedErr(
          `A ${constants.classes.optionGroup.items.className} cannot be ` +
            `found for the column ${headerPath}`,
        ),
      );
      return;
    }

    // remove element from the edit list and reshow the column in the add list
    setVisibility(addColumnElement, true);
    sortableList.removeChild(sortableColumn);
  };

  public getAddColumnSet = (): Result<HTMLElement[]> => {
    const element = document.getElementById(
      constants.layers.nestedSidebar.content.addColumnsSidebar.addColumnsOptions
        .id,
    );
    if (!element) {
      return Fail(
        new NoElWithIdErr(
          constants.layers.nestedSidebar.content.addColumnsSidebar
            .addColumnsOptions.id,
        ),
      );
    }

    return Success(
      Array.from(
        element.getElementsByClassName(
          constants.classes.optionGroup.items.className,
        ),
      ).filter(
        (element): element is HTMLElement => element instanceof HTMLElement,
      ),
    );
  };

  private isMainTablePage = (): boolean => {
    return document.getElementById(constants.mainPageIds.mainTable) !== null;
  };
}
