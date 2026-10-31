import { AlertManager } from "./alertManager";
import constants from "./constants";
import { UnexpectedErr } from "./errors";
import {
  HtmxConfigRequestEvent,
  HtmxEvent,
  onHtmxConfigRequestEvent,
} from "./htmx";
import { Fail, Result, Success } from "./result";
import {
  classSelector,
  elString,
  getChildByClass,
  getChildBySelector,
  getClosestByClass,
  getCurrentTarget,
  getElementById,
  getSelectedOption,
  getTarget,
  getTargetSelect,
  getTemplateById,
  getTemplateContentByClass,
  getTemplateContentBySelector,
  htmlElementAs,
  parseHtmlElement,
  setBoolAttribute,
} from "./utils";

interface ColumnType {
  inputType: string;
  dataType: string;
}

interface FilterRuleEl {
  element: HTMLElement;

  columnSelector: HTMLSelectElement;
  selectedColumn: HTMLOptionElement | null;

  operatorSelector: HTMLSelectElement;
  selectedOperator: HTMLOptionElement | null;

  valueContainer: HTMLElement;
  valueInput: HTMLInputElement | HTMLSelectElement | null;
}

export class TableFilterManager {
  private _alertManager: AlertManager;

  public constructor(alertManager: AlertManager) {
    this._alertManager = alertManager;

    document.addEventListener("change", this.handleRuleColumnOrOperatorChange);
    onHtmxConfigRequestEvent(this.handleHtmxConfigRequestEvent);
  }

  private handleHtmxConfigRequestEvent = (
    e: HtmxEvent<HtmxConfigRequestEvent>,
  ) => {
    const targetRes = parseHtmlElement(e.detail.ctx.sourceElement);
    if (!targetRes.ok) {
      e.preventDefault();
      this._alertManager.renderError(targetRes.error);
      return;
    }
    const target = targetRes.value;

    if (target.id !== constants.filterColumns.applyFiltersBtnId) {
      return;
    }

    const queryRes = this.buildFilterQueryString();
    if (!queryRes.ok) {
      e.preventDefault();
      this._alertManager.renderError(queryRes.error);
      return;
    }

    e.detail.ctx.request.action += `?filter=${queryRes.value}`;
  };

  public hgHandlerDeleteRule = (e: MouseEvent) => {
    const err = this.handleDeleteRule(e);
    if (err !== null) {
      this._alertManager.renderError(err);
    }
  };

  public hgHandlerDeleteRuleGroup = (e: MouseEvent) => {
    const err = this.handleDeleteRuleGroup(e);
    if (err !== null) {
      this._alertManager.renderError(err);
    }
  };

  private handleDeleteRule = (e: MouseEvent): Error | null => {
    const targetRes = getCurrentTarget(e);
    if (!targetRes.ok) {
      return targetRes.error;
    }

    const parentRuleRes = getClosestByClass(
      targetRes.value,
      constants.classes.filterRule.className,
    );
    if (!parentRuleRes.ok) {
      return parentRuleRes.error;
    }

    parentRuleRes.value.remove();
    return null;
  };

  private handleDeleteRuleGroup = (e: MouseEvent): Error | null => {
    const targetRes = getCurrentTarget(e);
    if (!targetRes.ok) {
      return targetRes.error;
    }

    const parentGroupRes = getClosestByClass(
      targetRes.value,
      constants.classes.filterGroup.className,
    );
    if (!parentGroupRes.ok) {
      return parentGroupRes.error;
    }

    parentGroupRes.value.remove();
    return null;
  };

  public hgHandlerNewRule = (e: MouseEvent) => {
    const r = this.newFilterRule();
    if (!r.ok) {
      this._alertManager.renderError(r.error);
      return;
    }

    const err = this.addRuleItem(e, r.value);
    if (err !== null) {
      this._alertManager.renderError(err);
    }
  };

  public hgHandlerNewRuleGroup = (e: MouseEvent) => {
    const r = this.newFilterGroup();
    if (!r.ok) {
      this._alertManager.renderError(r.error);
      return;
    }

    const err = this.addRuleItem(e, r.value);
    if (err !== null) {
      this._alertManager.renderError(err);
    }
  };

  private addRuleItem = (e: MouseEvent, newItem: HTMLElement): Error | null => {
    // Parse target
    const t = getTarget(e);
    if (!t.ok) {
      return t.error;
    }

    // Get parent group
    const g = getClosestByClass(
      t.value,
      constants.classes.filterGroup.className,
    );
    if (!g.ok) {
      return g.error;
    }

    // Get items collection from parent group
    const gItems = getChildByClass(
      g.value,
      constants.classes.filterGroup.children.items.className,
    );
    if (!gItems.ok) {
      return gItems.error;
    }

    // Access parent group depth
    const depth = parseInt(
      g.value.getAttribute(constants.classes.filterGroup.attributes.depth) ??
        "",
    );
    if (isNaN(depth)) {
      return new UnexpectedErr(
        `unable to parse depth from group: ${elString(g.value)}`,
      );
    }

    // If new item is group handle depth
    if (newItem.classList.contains(constants.classes.filterGroup.className)) {
      const newDepth = depth + 1;
      newItem.setAttribute(
        constants.classes.filterGroup.attributes.depth,
        newDepth.toString(),
      );

      setBoolAttribute(
        newItem,
        constants.classes.filterGroup.attributes.depthEven,
        newDepth % 2 === 0,
      );

      if (depth >= 3) {
        const newGroupButton = getChildByClass(
          newItem,
          constants.classes.filterGroup.children.controls.children
            .newRuleGroupButton.className,
        );
        if (!newGroupButton.ok) {
          return newGroupButton.error;
        }
        newGroupButton.value.classList.add(constants.globalModifiers.hidden);
      }
    }

    gItems.value.appendChild(newItem);
    return null;
  };

  private buildFilterQueryString = (): Result<string> => {
    const sidebarRes = getElementById(
      constants.layers.sidebar.content.editFiltersSidebar.id,
    );
    if (!sidebarRes.ok) {
      return sidebarRes;
    }

    const topLevelFilterGroupRes = getChildBySelector(
      sidebarRes.value,
      `${classSelector(constants.classes.filterGroup.className)}[data-depth="1"]`,
    );
    if (!topLevelFilterGroupRes.ok) {
      return topLevelFilterGroupRes;
    }

    return this.buildGroupFilter(topLevelFilterGroupRes.value);
  };

  private buildGroupFilter(filterGroupEl: HTMLElement): Result<string> {
    const groupOperatorElRes = getChildByClass(
      filterGroupEl,
      constants.classes.filterGroup.children.operator.className,
    );
    if (!groupOperatorElRes.ok) {
      return groupOperatorElRes;
    }

    const operatorValueRes = this.parseOperatorValue(groupOperatorElRes.value);
    if (!operatorValueRes.ok) {
      return operatorValueRes;
    }

    if (operatorValueRes.value === null) {
      return Fail(
        new UnexpectedErr(
          `operator is null - This should be a user error not an app error`,
        ),
      );
    }

    const filterGroupItems = getChildByClass(
      filterGroupEl,
      constants.classes.filterGroup.children.items.className,
    );
    if (!filterGroupItems.ok) {
      return filterGroupItems;
    }

    let qstr = "";
    const joinRuleStr = ` ${operatorValueRes.value} `;

    let itemCount = 0;
    for (const e of filterGroupItems.value.children) {
      const c = parseHtmlElement(e);
      if (!c.ok) {
        continue;
      }

      if (c.value.classList.contains(constants.classes.filterGroup.className)) {
        const groupFilterRes = this.buildGroupFilter(c.value);
        if (!groupFilterRes.ok) {
          return groupFilterRes;
        }
        if (itemCount > 0) {
          qstr += joinRuleStr;
        }
        qstr += `(${groupFilterRes.value})`;

        itemCount++;
        continue;
      }

      if (c.value.classList.contains(constants.classes.filterRule.className)) {
        const ruleRes = this.parseFilterRule(c.value);
        if (!ruleRes.ok) {
          return ruleRes;
        }

        const filterRuleRes = this.buildRuleFilter(ruleRes.value);
        if (!filterRuleRes.ok) {
          return filterRuleRes;
        }

        if (itemCount > 0) {
          qstr += joinRuleStr;
        }
        qstr += filterRuleRes.value;

        itemCount++;
        continue;
      }
    }

    return Success(qstr);
  }

  private buildRuleFilter(rule: FilterRuleEl): Result<string> {
    if (rule.selectedColumn === null) {
      return Fail(
        new UnexpectedErr(
          `selected column is null - THIS SHOULD BE A USER ERROR`,
        ),
      );
    }
    const selectedCol = rule.selectedColumn.value;

    const selectedColTypeRes = this.getColumnType(rule.selectedColumn);
    if (!selectedColTypeRes.ok) {
      return selectedColTypeRes;
    }
    const selectedColType = selectedColTypeRes.value;
    if (selectedColType === null) {
      return Fail(
        new UnexpectedErr(
          `unable to build rule filter: cannot read column type ` +
            `for ${elString(rule.selectedColumn)}`,
        ),
      );
    }

    if (rule.selectedOperator === null) {
      return Fail(
        new UnexpectedErr(
          `rule operator is null - THIS SHOULD BE A USER ERROR`,
        ),
      );
    }
    const selectedOperator = rule.selectedOperator.value;

    // Handle empty and not empty - special cases
    if (selectedOperator === constants.comparisonOperators.empty) {
      return Success(
        `${selectedCol} ${constants.comparisonOperators.equal} null`,
      );
    }

    if (selectedOperator === constants.comparisonOperators.notEmpty) {
      return Success(
        `${selectedCol} ${constants.comparisonOperators.notEqual} null`,
      );
    }

    if (rule.valueInput === null) {
      return Fail(
        new UnexpectedErr(
          `unable to build rule filter: expected value input but found null`,
        ),
      );
    }

    const valueStringRes = this.parseValueQueryString(
      selectedColType,
      rule.valueInput,
    );
    if (!valueStringRes.ok) {
      return valueStringRes;
    }
    const selectedValue = valueStringRes.value;

    return Success(`${selectedCol} ${selectedOperator} ${selectedValue}`);
  }

  private parseValueQueryString(
    colType: ColumnType,
    input: HTMLInputElement | HTMLSelectElement,
  ): Result<string> {
    const v = input.value;
    switch (colType.inputType) {
      case constants.filterColumns.columnTypes.string:
      case constants.filterColumns.columnTypes.identifier:
      case constants.filterColumns.columnTypes.lookup:
        return Success(`"${v}"`);
      case constants.filterColumns.columnTypes.number:
      case constants.filterColumns.columnTypes.enum:
        return Success(v);
      case constants.filterColumns.columnTypes.date:
        return Success(`datetime(${v}T00:00:00Z)`);
      case constants.filterColumns.columnTypes.point:
        return Success(`point(${v})`);

      default:
        return Fail(
          new UnexpectedErr(
            `unable to parse value query string: type not recognised: ${colType}`,
          ),
        );
    }
  }

  private parseOperatorValue(el: HTMLElement): Result<"and" | "or" | null> {
    const selectRes = htmlElementAs(el, HTMLSelectElement);
    if (!selectRes.ok) {
      return selectRes;
    }

    const operatorValue = selectRes.value.value;
    if (!operatorValue) {
      return Success(null);
    }

    switch (operatorValue) {
      case "and":
      case "or":
        return Success(operatorValue);
      default:
        return Fail(
          new UnexpectedErr(
            `unable to parse operator value: unexpected value: ${operatorValue}`,
          ),
        );
    }
  }

  private handleRuleColumnOrOperatorChange = (e: Event) => {
    // get target select or return
    const targetRes = getTargetSelect(e);
    if (!targetRes.ok) {
      return;
    }

    // Guard target has a parent fitler rule
    const parentFilterRule = getClosestByClass(
      targetRes.value,
      constants.classes.filterRule.className,
    );
    if (!parentFilterRule.ok) {
      return;
    }

    // parse filter fule
    const filterRuleRes = this.parseFilterRule(targetRes.value);
    if (!filterRuleRes.ok) {
      this._alertManager.renderError(filterRuleRes.error);
      return;
    }
    const filterRule = filterRuleRes.value;

    // get selected column type
    const columnTypeRes = this.getColumnType(filterRule.selectedColumn);
    if (!columnTypeRes.ok) {
      this._alertManager.renderError(columnTypeRes.error);
      return;
    }

    // update operator options
    for (const operator of filterRule.operatorSelector.options) {
      const supportsType =
        columnTypeRes.value !== null &&
        operator.hasAttribute(`data-supports-${columnTypeRes.value.inputType}`);

      if (
        !supportsType &&
        filterRule.operatorSelector.value === operator.value
      ) {
        filterRule.operatorSelector.value = "";
        filterRule.selectedOperator = null;
      }

      if (!operator.value || supportsType) {
        operator.classList.remove("hidden");
      } else {
        operator.classList.add("hidden");
      }
    }
    filterRule.operatorSelector.title =
      filterRule.selectedOperator?.textContent ?? "";
    filterRule.operatorSelector.disabled = !columnTypeRes.value;

    // update value input
    let columnDataType: string | null = null;
    if (
      columnTypeRes.value !== null &&
      !!filterRule.selectedOperator?.value &&
      filterRule.selectedOperator.value !==
        constants.comparisonOperators.empty &&
      filterRule.selectedOperator.value !==
        constants.comparisonOperators.notEmpty
    ) {
      columnDataType = columnTypeRes.value.dataType;
    }

    // remove existing value input if necessary
    if (columnDataType === null && filterRule.valueInput !== null) {
      filterRule.valueInput.remove();
    }

    // update existing value input if necessary
    if (
      columnDataType !== null &&
      (filterRule.valueInput === null ||
        filterRule.valueInput.dataset.type !== columnTypeRes.value?.dataType)
    ) {
      if (filterRule.valueInput !== null) {
        filterRule.valueInput.remove();
      }
      const newValueRes = this.newFilterInput(columnDataType);
      if (!newValueRes.ok) {
        this._alertManager.renderError(newValueRes.error);
        return;
      }
      filterRule.valueContainer.appendChild(newValueRes.value);
    }

    filterRule.columnSelector.title = filterRule.selectedColumn?.text ?? "";
  };

  private parseFilterRule(el: HTMLElement): Result<FilterRuleEl> {
    // access rule element
    const filterRuleRes = getClosestByClass(
      el,
      constants.classes.filterRule.className,
    );
    if (!filterRuleRes.ok) {
      return filterRuleRes;
    }

    // access column selector
    const columnSelectorElRes = getChildByClass(
      filterRuleRes.value,
      constants.classes.filterRule.children.columnSelector.className,
    );
    if (!columnSelectorElRes.ok) {
      return columnSelectorElRes;
    }

    const columnSelectorElTypedRes = htmlElementAs(
      columnSelectorElRes.value,
      HTMLSelectElement,
    );
    if (!columnSelectorElTypedRes.ok) {
      return columnSelectorElTypedRes;
    }

    const selectedColumnRes = getSelectedOption(columnSelectorElTypedRes.value);
    if (!selectedColumnRes.ok) {
      return selectedColumnRes;
    }

    // access operator selector
    const operatorSelectorElRes = getChildByClass(
      filterRuleRes.value,
      constants.classes.filterRule.children.operatorSelector.className,
    );
    if (!operatorSelectorElRes.ok) {
      return operatorSelectorElRes;
    }

    const operatorSelectorElTypedRes = htmlElementAs(
      operatorSelectorElRes.value,
      HTMLSelectElement,
    );
    if (!operatorSelectorElTypedRes.ok) {
      return operatorSelectorElTypedRes;
    }

    const selectedOperatorRes = getSelectedOption(
      operatorSelectorElTypedRes.value,
    );
    if (!selectedOperatorRes.ok) {
      return selectedOperatorRes;
    }

    // get value input wrapper
    const valueInputWrapperRes = getChildByClass(
      filterRuleRes.value,
      constants.classes.filterRule.children.valueInputWrapper.className,
    );
    if (!valueInputWrapperRes.ok) {
      return valueInputWrapperRes;
    }

    const inputEl = valueInputWrapperRes.value.querySelector("input,select");
    if (inputEl !== null) {
      if (
        !(inputEl instanceof HTMLInputElement) &&
        !(inputEl instanceof HTMLSelectElement)
      ) {
        return Fail(
          new UnexpectedErr(
            "expected value input to be an input or select element",
          ),
        );
      }
    }

    return Success({
      element: filterRuleRes.value,

      columnSelector: columnSelectorElTypedRes.value,
      selectedColumn: selectedColumnRes.value,

      operatorSelector: operatorSelectorElTypedRes.value,
      selectedOperator: selectedOperatorRes.value,

      valueContainer: valueInputWrapperRes.value,
      valueInput: inputEl,
    });
  }

  private getColumnType = (
    el: HTMLOptionElement | null,
  ): Result<ColumnType | null> => {
    if (el === null || !el.value) {
      return Success(null);
    }

    const inputType = el.dataset.inputType;
    if (!inputType) {
      return Fail(
        new UnexpectedErr(
          `expected ${elString(el)} to have data-input-type attribute`,
        ),
      );
    }

    const dataType = el.dataset.dataType;
    if (!dataType) {
      return Fail(
        new UnexpectedErr(
          `expected ${elString(el)} to have data-data-type attribute`,
        ),
      );
    }

    return Success({
      inputType: inputType,
      dataType: dataType,
    });
  };

  private newFilterRule = (): Result<HTMLElement> => {
    const templRes = getTemplateById(
      constants.filterColumns.templates.filterRule,
    );
    if (!templRes.ok) {
      return templRes;
    }

    const prototype = getTemplateContentByClass(
      templRes.value,
      constants.classes.filterRule.className,
    );
    if (!prototype.ok) {
      return prototype;
    }

    const clone = prototype.value.cloneNode(true) as HTMLElement;
    return Success(clone);
  };

  private newFilterGroup = (): Result<HTMLElement> => {
    const templRes = getTemplateById(
      constants.filterColumns.templates.filterGroup,
    );
    if (!templRes.ok) {
      return templRes;
    }

    const prototype = getTemplateContentByClass(
      templRes.value,
      constants.classes.filterGroup.className,
    );
    if (!prototype.ok) {
      return prototype;
    }

    const clone = prototype.value.cloneNode(true) as HTMLElement;
    return Success(clone);
  };

  private newFilterInput = (type: string): Result<HTMLElement> => {
    const templRes = getTemplateById(
      constants.filterColumns.templates.filterInputs,
    );
    if (!templRes.ok) {
      return templRes;
    }

    const prototype = getTemplateContentBySelector(
      templRes.value,
      `[data-type=${type}]`,
    );
    if (!prototype.ok) {
      return prototype;
    }

    const clone = prototype.value.cloneNode(true) as HTMLElement;
    return Success(clone);
  };
}
