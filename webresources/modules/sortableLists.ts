import constants from "./constants";
import { ErrorHandler, NullParentErr } from "./errorHandler";
import {
  classSelector,
  parseHtmlElement,
  switchSiblingElements,
} from "./utils";

export interface DraggedItem {
  sortableList: HTMLElement;
  item: HTMLElement;
}

export class SortableListManager {
  private readonly _errHandler: ErrorHandler;
  private _currentDraggedItem: DraggedItem | null = null;

  public constructor(errorHandler: ErrorHandler) {
    this._errHandler = errorHandler;

    // Drag controls
    document.addEventListener("dragstart", this.handleDragStart);
    document.addEventListener("dragenter", this.handleDragEnter);
    document.addEventListener("dragend", this.handleDragEnd);

    // Keyboard controls
    document.addEventListener("keydown", this.handleKeyDown);
  }

  private handleDragStart = (e: Event) => {
    const t = parseHtmlElement(e.target);
    if (
      !t ||
      !t.classList.contains(constants.classes.sortableList.items.className)
    ) {
      return;
    }

    const parentList = parseHtmlElement(
      t.closest(classSelector(constants.classes.sortableList.className)),
    );

    if (!parentList) {
      this._errHandler.handle(
        new NullParentErr(
          t,
          classSelector(constants.classes.sortableList.className),
        ),
      );
      return;
    }

    t.setAttribute("aria-pressed", "true");

    this._currentDraggedItem = {
      sortableList: parentList,
      item: t,
    };
  };

  private handleDragEnter = (e: Event) => {
    if (!this._currentDraggedItem) {
      return;
    }

    const dropZone = parseHtmlElement(e.target);
    if (
      !dropZone ||
      !dropZone.classList.contains(
        constants.classes.sortableList.items.className,
      ) ||
      this._currentDraggedItem.item === dropZone
    ) {
      return;
    }

    const dropZoneParentList = parseHtmlElement(
      dropZone.closest(classSelector(constants.classes.sortableList.className)),
    );
    if (!dropZoneParentList) {
      this._errHandler.handle(
        new NullParentErr(
          dropZone,
          classSelector(constants.classes.sortableList.className),
        ),
      );
      return;
    }

    const error = switchSiblingElements(
      this._currentDraggedItem.sortableList,
      classSelector(constants.classes.sortableList.items.className),
      this._currentDraggedItem.item,
      dropZone,
    );
    if (error !== null) {
      this._errHandler.handle(error);
    }
  };

  private handleDragEnd = () => {
    if (this._currentDraggedItem?.item) {
      this._currentDraggedItem.item.setAttribute("aria-pressed", "false");
    }
    this._currentDraggedItem = null;
  };

  private handleKeyDown = (e: KeyboardEvent) => {
    if (e.key === constants.keyCodes.Space) {
      if (this._currentDraggedItem === null) {
        this.handleDragStart(e);
      } else {
        this.handleDragEnd();
      }
      return;
    }

    if (
      this._currentDraggedItem === null ||
      (e.key !== constants.keyCodes.ArrowDown &&
        e.key !== constants.keyCodes.ArrowUp)
    ) {
      return;
    }

    const nextEl =
      e.key === constants.keyCodes.ArrowUp
        ? parseHtmlElement(this._currentDraggedItem.item.previousElementSibling)
        : parseHtmlElement(this._currentDraggedItem.item.nextElementSibling);
    if (
      !nextEl ||
      !nextEl.classList.contains(constants.classes.sortableList.items.className)
    ) {
      return;
    }

    const error = switchSiblingElements(
      this._currentDraggedItem.sortableList,
      classSelector(constants.classes.sortableList.items.className),
      this._currentDraggedItem.item,
      nextEl,
    );
    if (error !== null) {
      this._errHandler.handle(error);
      return;
    }

    this._currentDraggedItem.item.focus();
  };
}
