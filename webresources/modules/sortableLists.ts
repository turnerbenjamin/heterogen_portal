import { DraggedItem } from "./state";
import * as utils from "./utils";
import * as err from "./errorHandling";

type DraggedItemGetter = () => DraggedItem | null;
type DraggedItemSetter = (_v: DraggedItem | null) => void;

export function InitSortableLists(
  draggedItemGetter: DraggedItemGetter,
  draggedItemSetter: DraggedItemSetter,
) {
  // Drag controls
  document.addEventListener("dragstart", (e: DragEvent) =>
    handleDragStart(e, draggedItemSetter),
  );
  document.addEventListener("dragenter", (e: DragEvent) =>
    handleDragEnter(e, draggedItemGetter),
  );
  document.addEventListener("dragend", () =>
    handleDragEnd(draggedItemGetter, draggedItemSetter),
  );

  // Keyboard controls
  document.addEventListener("keydown", handleKeyDown);
}

function handleDragStart(e: DragEvent, setDraggedItem: DraggedItemSetter) {
  const t = utils.ParseHtmlElement(e.target);
  if (!t || !t.classList.contains("sortable-item")) {
    return;
  }

  const parentList = utils.ParseHtmlElement(t.closest(".sortable-list"));
  if (!parentList) {
    return;
  }

  t.classList.add("dragging");

  setDraggedItem({
    sortableList: parentList,
    item: t,
  });
}

function handleDragEnd(
  getDraggedItem: DraggedItemGetter,
  setDraggedItem: DraggedItemSetter,
) {
  const currentDraggedItem = getDraggedItem();
  if (currentDraggedItem?.item) {
    currentDraggedItem.item.classList.remove("dragging");
  }
  setDraggedItem(null);
}

function handleDragEnter(e: DragEvent, getDraggedItem: DraggedItemGetter) {
  const currentDraggedItem = getDraggedItem();
  if (!currentDraggedItem) {
    return;
  }

  const dropZone = utils.ParseHtmlElement(e.target);
  if (
    !dropZone ||
    !dropZone.classList.contains("sortable-item") ||
    currentDraggedItem.item === dropZone
  ) {
    return;
  }

  const dropZoneParentList = utils.ParseHtmlElement(
    dropZone.closest(".sortable-list"),
  );
  if (dropZoneParentList !== currentDraggedItem.sortableList) {
    return;
  }

  const error = utils.SwitchSiblingElements(
    currentDraggedItem.sortableList,
    ".sortable-item",
    currentDraggedItem.item,
    dropZone,
  );
  if (error !== null) {
    err.handle(error);
  }
}

const handleKeyDown = (e: KeyboardEvent) => {
  if (e.key !== "ArrowDown" && e.key !== "ArrowUp") {
    return;
  }

  const target = utils.ParseHtmlElement(e.target);
  const elToMove = utils.ParseHtmlElement(target?.closest(".sortable-item"));
  if (!target || !elToMove) {
    return;
  }

  const parentList = utils.ParseHtmlElement(elToMove.closest(".sortable-list"));
  if (!parentList) {
    return;
  }

  const nextEl =
    e.key === "ArrowUp"
      ? utils.ParseHtmlElement(elToMove.previousElementSibling)
      : utils.ParseHtmlElement(elToMove.nextElementSibling);
  if (!nextEl || !nextEl.classList.contains("sortable-item")) {
    return;
  }

  const error = utils.SwitchSiblingElements(
    parentList,
    ".sortable-item",
    elToMove,
    nextEl,
  );
  if (error !== null) {
    err.handle(error);
    return;
  }

  elToMove.focus();
};
