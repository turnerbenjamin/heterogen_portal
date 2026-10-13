import constants from "./constants";
import {
  NoElWithIdErr,
  UnexpectedErr,
  QSelectFail,
  NullAttributeErr,
  NullParentErr,
} from "./errors";
import {
  AfterSwapEventContext,
  ConfigRequestEventContext,
  HtmxEvent,
  onHtmxAfterSwapEvent,
  onHtmxConfigRequestEvent,
} from "./htmx";
import { Fail, Result, Success } from "./result";
import {
  setVisibility,
  parseHtmlElement,
  classSelector,
  getElementById,
} from "./utils";
import { Hook, WindowManager } from "./windowManager";

interface SelectStatement {
  statement: string;
  isDiff: boolean;
}

export class MainTable {
  private _windowManager: WindowManager;
  private editColumnsListPreEditState: HTMLElement | null = null;
  private addColumnsListPreEditState: HTMLElement | null = null;

  public constructor(windowManager: WindowManager) {
    this._windowManager = windowManager;

    this._windowManager.registerHook(
      Hook.openSidebar,
      this.snapshotColumnState,
    );

    this._windowManager.registerHook(
      Hook.closeSidebar,
      this.restoreEditColumnState,
    );

    this._windowManager.registerHook(
      Hook.closeNestedSidebar,
      this.handleCloseNestedSidebar,
    );

    onHtmxConfigRequestEvent(this.handleHtmxConfigRequestEvent);
    onHtmxAfterSwapEvent(this.handleHtmxAfterSwapEvent);
  }

  private handleHtmxConfigRequestEvent = (
    e: HtmxEvent<ConfigRequestEventContext>,
  ) => {
    const targetRes = parseHtmlElement(e.detail.elt);
    if (!targetRes.ok) {
      e.preventDefault();
      this._windowManager.displayError(targetRes.error);
      return;
    }
    const target = targetRes.value;

    if (
      target.id !==
      constants.layers.sidebar.content.editColumnsSidebar.applyButton.id
    ) {
      return;
    }

    const selectRes = this.buildSelectStatement();
    if (!selectRes.ok) {
      e.preventDefault();
      this._windowManager.displayError(selectRes.error);
      return;
    }

    if (!selectRes.value.isDiff) {
      e.preventDefault();
      // this._windowManager.hgHandlerCloseSidebar();
    }

    e.detail.path += `?select=${selectRes.value.statement}`;
    this.snapshotColumnState();
  };

  private handleHtmxAfterSwapEvent = (e: HtmxEvent<AfterSwapEventContext>) => {
    const targetRes = parseHtmlElement(e.detail.elt);
    if (!targetRes.ok) {
      e.preventDefault();
      this._windowManager.displayError(targetRes.error);
      return;
    }
    const target = targetRes.value;

    if (target.id !== constants.mainPages.mainTablePage.mainTable.id) {
      return;
    }

    this._windowManager.hgHandlerCloseSidebar();
  };

  public hgHandlerAddColumns = () => {
    const getListElementsRes = this.getEditAndAddColumnListEls();
    if (!getListElementsRes.ok) {
      this._windowManager.displayError(getListElementsRes.ok);
      return;
    }
    const [editColumnsList, addColumnsList] = getListElementsRes.value;

    const editColumnsListItemsRes =
      this.getEditColumnsListItems(editColumnsList);
    if (!editColumnsListItemsRes.ok) {
      this._windowManager.displayError(editColumnsListItemsRes.error);
      return;
    }
    const editColumnsListItems = editColumnsListItemsRes.value;

    const protoEditColumn = editColumnsListItems[0];

    const addColumnsListItemsRes = this.getAddColumnsListItems(addColumnsList);
    if (!addColumnsListItemsRes.ok) {
      this._windowManager.displayError(addColumnsListItemsRes.error);
      return;
    }

    let uncheckedCount = 0;
    for (const optionSetEl of addColumnsListItemsRes.value) {
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
      editColumnsList.appendChild(newEditItem);
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

    const getAddColumnOptionElRes = getElementById(
      constants.layers.nestedSidebar.content.addColumnsSidebar.addColumnsOptions
        .id,
    );
    if (!getAddColumnOptionElRes.ok) {
      this._windowManager.displayError(getAddColumnOptionElRes.error);
      return;
    }
    const addColumnOptionEl = getAddColumnOptionElRes.value;

    const checkboxEls = addColumnOptionEl.getElementsByTagName("input");
    for (const checkboxEl of checkboxEls) {
      checkboxEl.checked = false;
    }
  };

  public hgHandlerRemoveColumn = (e: MouseEvent) => {
    const targetRes = parseHtmlElement(e.currentTarget);
    if (!targetRes.ok) {
      this._windowManager.displayError(targetRes.error);
      return;
    }
    const target = targetRes.value;

    const sortableColumnRes = parseHtmlElement(
      target.closest(
        classSelector(constants.classes.sortableList.items.className),
      ),
    );
    if (!sortableColumnRes.ok) {
      return;
    }
    const sortableColumn = sortableColumnRes.value;

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

    const addColumnListRes = this.getAddColumnsList();
    if (!addColumnListRes.ok) {
      this._windowManager.displayError(addColumnListRes.error);
      return;
    }

    const addColumnListItemsResult = this.getAddColumnsListItems(
      addColumnListRes.value,
    );
    if (!addColumnListItemsResult.ok) {
      this._windowManager.displayError(addColumnListItemsResult.error);
      return;
    }

    // find the column in the add columns set and mark it as visible
    let addColumnElement: HTMLElement | null = null;
    for (const currentEl of addColumnListItemsResult.value) {
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

  private snapshotColumnState = () => {
    const res = this.getEditAndAddColumnListEls();
    if (!res.ok) {
      this._windowManager.displayError(res.error);
      return;
    }
    const [editColumnsListEl, addColumnsListEl] = res.value;

    const editColumnListCloneRes = parseHtmlElement(
      editColumnsListEl.cloneNode(true),
    );
    if (!editColumnListCloneRes.ok) {
      this._windowManager.displayError(editColumnListCloneRes.error);
      return;
    }

    const addColumnsListCloneRes = parseHtmlElement(
      addColumnsListEl.cloneNode(true),
    );
    if (!addColumnsListCloneRes.ok) {
      this._windowManager.displayError(addColumnsListCloneRes);
      return;
    }

    this.editColumnsListPreEditState = editColumnListCloneRes.value;
    this.addColumnsListPreEditState = addColumnsListCloneRes.value;
  };

  private restoreEditColumnState = () => {
    if (!this.editColumnsListPreEditState || !this.addColumnsListPreEditState) {
      this._windowManager.displayError(
        new UnexpectedErr(
          "Unable to restore state in the edit column sidebar; one or both " +
            "snapshots are null",
        ),
      );
      return;
    }

    const res = this.getEditAndAddColumnListEls();
    if (!res.ok) {
      this._windowManager.displayError(res.error);
      return;
    }
    const [editColumnsListEl, addColumnsListEl] = res.value;

    editColumnsListEl.replaceWith(this.editColumnsListPreEditState);
    addColumnsListEl.replaceWith(this.addColumnsListPreEditState);

    this.editColumnsListPreEditState = null;
    this.addColumnsListPreEditState = null;
  };

  private buildSelectStatement = (): Result<SelectStatement> => {
    const editColumnsListElRes = this.getEditColumnsList();
    if (!editColumnsListElRes.ok) {
      return Fail(editColumnsListElRes.error);
    }

    if (this.editColumnsListPreEditState === null) {
      return Fail(
        new UnexpectedErr(
          "Unable to build select statement: Cannot access pre-edit column " +
            "list snapshot",
        ),
      );
    }

    const currentEditColumnItemsRes = this.getEditColumnsListItems(
      editColumnsListElRes.value,
    );
    if (!currentEditColumnItemsRes.ok) {
      return Fail(currentEditColumnItemsRes.error);
    }

    const storedEditColumnItemsRes = this.getEditColumnsListItems(
      this.editColumnsListPreEditState,
    );
    if (!storedEditColumnItemsRes.ok) {
      return Fail(storedEditColumnItemsRes.error);
    }

    const select = [];
    let isDiff =
      currentEditColumnItemsRes.value.length !==
      storedEditColumnItemsRes.value.length;

    for (let i = 0; i < currentEditColumnItemsRes.value.length; i++) {
      const currentColEl = currentEditColumnItemsRes.value[i];
      const storedColEl =
        !isDiff && i < storedEditColumnItemsRes.value.length
          ? storedEditColumnItemsRes.value[i]
          : null;

      const currentColPath = currentColEl.getAttribute(
        constants.classes.sortableList.items.attributes.columnPath,
      );
      if (!currentColPath) {
        return Fail(
          new NullAttributeErr(
            currentColEl,
            constants.classes.sortableList.items.attributes.columnPath,
          ),
        );
      }

      const storedColPath =
        storedColEl?.getAttribute(
          constants.classes.sortableList.items.attributes.columnPath,
        ) ?? null;

      if (currentColPath !== storedColPath) {
        isDiff = true;
      }
      select.push(currentColPath);
    }

    return Success({
      statement: select.join(","),
      isDiff: isDiff,
    });
  };

  private getEditAndAddColumnListEls = (): Result<
    [HTMLElement, HTMLElement]
  > => {
    const editColumnsListElRes = this.getEditColumnsList();
    if (!editColumnsListElRes.ok) {
      return Fail(editColumnsListElRes.error);
    }

    const addColumnsListElRes = this.getAddColumnsList();
    if (!addColumnsListElRes.ok) {
      return Fail(addColumnsListElRes.error);
    }

    return Success([editColumnsListElRes.value, addColumnsListElRes.value]);
  };

  private getEditColumnsList(): Result<HTMLElement> {
    const editColumnsListEl = document.getElementById(
      constants.layers.sidebar.content.editColumnsSidebar.editColumnsList.id,
    );
    if (!editColumnsListEl) {
      return Fail(
        new NoElWithIdErr(
          constants.layers.sidebar.content.editColumnsSidebar.editColumnsList
            .id,
        ),
      );
    }
    return Success(editColumnsListEl);
  }

  private getEditColumnsListItems(listEl: HTMLElement): Result<HTMLElement[]> {
    const listItems = Array.from(
      listEl.getElementsByClassName(
        constants.classes.sortableList.items.className,
      ),
    ).filter(
      (element): element is HTMLElement => element instanceof HTMLElement,
    );

    if (!listItems?.length) {
      return Fail(
        new UnexpectedErr(
          `expected ${listEl.id} to contain at least one child of type ` +
            constants.classes.sortableList.items.className,
        ),
      );
    }

    return Success(listItems);
  }

  private getAddColumnsList(): Result<HTMLElement> {
    const addColumnsListEl = document.getElementById(
      constants.layers.nestedSidebar.content.addColumnsSidebar.addColumnsOptions
        .id,
    );
    if (!addColumnsListEl) {
      return Fail(
        new NoElWithIdErr(
          constants.layers.nestedSidebar.content.addColumnsSidebar
            .addColumnsOptions.id,
        ),
      );
    }
    return Success(addColumnsListEl);
  }

  private getAddColumnsListItems(listEl: HTMLElement): Result<HTMLElement[]> {
    const listItems = Array.from(
      listEl.getElementsByClassName(
        constants.classes.optionGroup.items.className,
      ),
    ).filter(
      (element): element is HTMLElement => element instanceof HTMLElement,
    );

    if (!listItems?.length) {
      return Fail(
        new UnexpectedErr(
          `expected ${listEl.id} to contain at least one child of type ` +
            constants.classes.optionGroup.items.className,
        ),
      );
    }

    return Success(listItems);
  }

  private isMainTablePage = (): boolean => {
    return (
      document.getElementById(constants.mainPages.mainTablePage.id) !== null
    );
  };
}
