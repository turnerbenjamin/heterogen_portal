export default {
  mainPageIds: {
    mainTable: "main-table-content",
  },

  layers: {
    app: {
      id: "app",
    },

    sidebar: {
      id: "sidebar",
      content: {
        editColumnsSidebar: {
          editColumnsList: {
            id: "edit-columns-list",
          },
          openAddColumnScreenButton: {
            id: "open-add-column-screen-button",
          },
        },
      },
      events: {
        opened: "openedsidebar",
        closed: "closedsidebar",
      },
    },

    nestedSidebar: {
      id: "nested-sidebar",
      content: {
        addColumnsSidebar: {
          addColumnsOptions: {
            id: "add-columns-options",
          },
        },
      },
      events: {
        opened: "openednestedsidebar",
        closed: "closednestedsidebar",
      },
    },

    alert: {
      id: "alert",
    },

    toast: {
      toastItem: {
        className: "toast",
        animations: {
          showToast: "toast-show",
          closeToast: "toast-hide",
        },
        progressBar: {
          animations: {
            depleteBar: "deplete-toast-progress-bar",
          },
        },
        modifiers: {
          show: "show",
          visible: "visible",
          hide: "hide",
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
  },

  keyCodes: {
    ArrowUp: "ArrowUp",
    ArrowDown: "ArrowDown",
    Space: " ",
    Escape: "Escape",
  },
} as const;
