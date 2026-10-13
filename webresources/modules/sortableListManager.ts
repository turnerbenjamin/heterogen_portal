import constants from "./constants";
import { NullParentErr } from "./errors";
import {
  classSelector,
  parseHtmlElement,
  switchSiblingElements,
} from "./utils";
import { Hook, WindowManager } from "./windowManager";

export interface DraggedItem {
  sortableList: HTMLElement;
  item: HTMLElement;
}

export class SortableListManager {
  private readonly _windowManager: WindowManager;
  private _currentDraggedItem: DraggedItem | null = null;

  public constructor(windowManager: WindowManager) {
    this._windowManager = windowManager;

    // Drag controls
    document.addEventListener("dragstart", this.handleDragStart);
    document.addEventListener("dragenter", this.handleDragEnter);
    document.addEventListener("dragend", this.handleDragEnd);

    // Keyboard controls
    document.addEventListener("keydown", this.handleKeyDown);
    this._windowManager.registerHook(Hook.layerChange, this.handleDragEnd);
  }

  private handleDragStart = (e: Event) => {
    const targetRes = parseHtmlElement(e.target);
    if (!targetRes.ok) {
      return;
    }
    const target = targetRes.value;

    if (
      !target ||
      !target.classList.contains(constants.classes.sortableList.items.className)
    ) {
      return;
    }

    const parentListRes = parseHtmlElement(
      target.closest(classSelector(constants.classes.sortableList.className)),
    );
    if (!parentListRes.ok) {
      this._windowManager.displayError(
        new NullParentErr(
          target,
          classSelector(constants.classes.sortableList.className),
        ),
      );
      return;
    }

    target.setAttribute("aria-pressed", "true");

    this._currentDraggedItem = {
      sortableList: parentListRes.value,
      item: target,
    };
  };

  private handleDragEnter = (e: Event) => {
    if (!this._currentDraggedItem) {
      return;
    }

    const dropzoneRes = parseHtmlElement(e.target);
    if (
      !dropzoneRes.ok ||
      !dropzoneRes.value.classList.contains(
        constants.classes.sortableList.items.className,
      ) ||
      this._currentDraggedItem.item === dropzoneRes.value
    ) {
      return;
    }
    const dropzone = dropzoneRes.value;

    const dropZoneParentList = parseHtmlElement(
      dropzone.closest(classSelector(constants.classes.sortableList.className)),
    );
    if (!dropZoneParentList) {
      this._windowManager.displayError(
        new NullParentErr(
          dropzone,
          classSelector(constants.classes.sortableList.className),
        ),
      );
      return;
    }

    const error = switchSiblingElements(
      this._currentDraggedItem.sortableList,
      classSelector(constants.classes.sortableList.items.className),
      this._currentDraggedItem.item,
      dropzone,
    );
    if (error !== null) {
      this._windowManager.displayError(error);
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
      e.preventDefault(); // stop default scroll down behaviour

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

    const nextElRes =
      e.key === constants.keyCodes.ArrowUp
        ? parseHtmlElement(this._currentDraggedItem.item.previousElementSibling)
        : parseHtmlElement(this._currentDraggedItem.item.nextElementSibling);
    if (
      !nextElRes.ok ||
      !nextElRes.value.classList.contains(
        constants.classes.sortableList.items.className,
      )
    ) {
      return;
    }

    const error = switchSiblingElements(
      this._currentDraggedItem.sortableList,
      classSelector(constants.classes.sortableList.items.className),
      this._currentDraggedItem.item,
      nextElRes.value,
    );
    if (error !== null) {
      this._windowManager.displayError(error);
      return;
    }

    this._currentDraggedItem.item.focus();
  };
}
