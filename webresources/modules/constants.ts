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
          applyButton: {
            id: "apply-column-edits-button",
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
