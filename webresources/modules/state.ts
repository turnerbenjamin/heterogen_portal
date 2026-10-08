export interface DraggedItem {
  sortableList: HTMLElement;
  item: HTMLElement;
}

export class AppState {
  private draggedItem: DraggedItem | null = null;

  public getDraggedItem = () => {
    return this.draggedItem;
  };

  public setDraggedItem = (v: DraggedItem | null) => {
    this.draggedItem = v;
  };
}
