export default {
  mainPages: {
    mainTablePage: {
      id: "main-table-content",
      mainTable: {
        id: "main-table",
      },
    },
  },

  filterColumns: {
    componentId: "edit-filters-sidebar",
    templates: {
      filterRule: "filter-rule_template",
      filterGroup: "filter-group_template",
      filterInputs: "filter-input_templates",
    },

    columnTypes: {
      string: "string",
      number: "number",
      date: "date",
      enum: "enum",
      identifier: "identifier",
      lookup: "lookup",
      point: "point",
    },

    applyFiltersBtnId: "apply-filters-button",
  },

  editColumns: {
    componentId: "edit-columns-sidebar",
    templates: {
      sortableItem: "sortable-item_template",
    },
  },

  layers: {
    app: {
      id: "app",
    },

    sidebar: {
      id: "sidebar",
      content: {
        labelElAttribute: "sidebar-label",
        editColumnsSidebar: {
          id: "edit-columns-sidebar",

          editColumnsList: {
            id: "edit-columns-list",
          },
          openAddColumnScreenButton: {
            id: "open-add-column-screen-button",
          },
          applyButton: {
            id: "apply-column-edits-button",
          },
        },
        editFiltersSidebar: {
          id: "edit-filters-sidebar",
        },
      },
    },

    nestedSidebar: {
      id: "nested-sidebar",
      content: {
        addColumnsSidebar: {
          id: "add-columns-sidebar",
          addColumnsOptions: {
            id: "add-columns-options",
          },
        },
      },
    },

    alert: {
      id: "alert",
      containerId: "alert-layer",
      title: {
        id: "alert-title",
      },
      message: {
        id: "alert-message",
      },
      detail: {
        id: "alert-detail",
      },
      confirmButton: {
        id: "alert-accept",
      },
      declineButton: {
        id: "alert-decline",
      },
      modifiers: {
        error: "error",
      },
      attributes: {
        serverAlert: "server-alert",
      },
    },

    toast: {
      id: "toast-container",
      toastItem: {
        className: "toast",
        animations: {
          showToast: "toast-show",
          closeToast: "toast-hide",
        },
        progressBar: {
          className: "progress-bar",
          animations: {
            depleteBar: "deplete-toast-progress-bar",
          },
        },
        modifiers: {
          show: "show",
          visible: "visible",
          hide: "hide",
          notification: "notification",
          warning: "warning",
          error: "error",
          success: "success",
        },
      },
    },

    modifierClasses: {
      show: "show",
      close: "close",
    },

    modifierAttributes: {
      inert: "inert",
    },

    contentShape: {
      initialFocusAttribute: "initial-focus",
    },
  },

  sidebarCommon: {
    animations: {
      showSidebar: "show-sidebar",
      closeSidebar: "close-sidebar",
    },
    contentGroupClass: "sidebar-content-group-container",
  },

  globalModifiers: {
    hidden: "hidden",
  },

  classes: {
    sortableList: {
      className: "sortable-list",
      items: {
        className: "sortable-item",
        attributes: {
          columnPath: "data-column-path",
        },
        textContainer: {
          className: "sortable-item-text",
        },
      },
    },

    optionGroup: {
      className: "option-group",
      items: {
        className: "option-group-item",
        attributes: {
          columnPath: "data-column-path",
        },
      },
    },

    filterGroup: {
      className: "filter-group",
      attributes: {
        depth: "data-depth",
        depthEven: "data-depth-even",
      },
      children: {
        operator: {
          className: "filter-group-operator",
        },
        items: {
          className: "filter-group-items",
        },
        controls: {
          className: "filter-group-controls",
          children: {
            newRuleGroupButton: {
              className: "filter-group-controls_new-rule-button",
            },
          },
        },
      },
    },

    filterRule: {
      className: "filter-rule",
      children: {
        columnSelector: {
          className: "filter-rule-display-name",
        },
        operatorSelector: {
          className: "filter-rule-operator",
        },
        valueInputWrapper: {
          className: "filter-rule-value",
        },
      },
    },
  },

  comparisonOperators: {
    empty: "empty",
    notEmpty: "notempty",
    equal: "eq",
    notEqual: "ne",
    greaterThan: "gt",
    greaterThanOrEqual: "ge",
    lessThan: "lt",
    lessThanOrEqual: "le",
    contains: "contains",
    startsWith: "startswith",
    endsWith: "endswith",
  },

  keyCodes: {
    ArrowUp: "ArrowUp",
    ArrowDown: "ArrowDown",
    Space: " ",
    Escape: "Escape",
  },
} as const;
