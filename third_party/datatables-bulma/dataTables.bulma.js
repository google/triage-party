(function(b) {
  "function" === typeof define && define.amd ? define(["jquery", "datatables.net"], function(a) {
    return b(a, window, document)
  }) : "object" === typeof exports ? module.exports = function(a, d) {
    a || (a = window);
    if (!d || !d.fn.dataTable) d = require("datatables.net")(a, d).$;
    return b(d, a, a.document)
  } : b(jQuery, window, document)
})(function(b, a, d) {
  var f = b.fn.dataTable;
  b.extend(!0, f.defaults, {
    dom: "<'columns'<'column is-6'l><'column is-6'f>><'columns'<'column is-12'tr>><'columns'<'column is-5'i><'column is-7'p>>",
    renderer: "bulma"
  });
  b.extend(f.ext.classes, {
    sWrapper: "dataTables_wrapper dt-bulma",
    sFilterInput: "input is-small",
    sLengthSelect: "input is-small",
    sProcessing: "dataTables_processing panel",
    sPageButton: "pagination-link",
    sPagePrevious: "pagination-previous",
    sPageNext: "pagination-next",
    sPageButtonActive: "is-current"
  });
  f.ext.renderer.pageButton.bulma = function(a, h, r, m, j, n) {
    var o = new f.Api(a),
      s = a.oClasses,
      k = a.oLanguage.oPaginate,
      t = a.oLanguage.oAria.paginate || {},
      e, g, p = 0,
      q = function(d, f) {
        var l, h, i, c, m = function(a) {
          a.preventDefault();
          (!b(a.currentTarget).is("[disabled]") && !b(a.currentTarget).is("#table_ellipsis")) && o.page() != a.data.action && o.page(a.data.action).draw("page")
        };
        l = 0;
        for (h = f.length; l < h; l++)
          if (c = f[l], b.isArray(c)) q(d, c);
          else {
            g = e = "";
            var dd = false;
            switch (c) {
              case "ellipsis":
                e = "&#x2026;";
                dd = true;
                break;
              case "first":
                e = k.sFirst;
                dd = c + (0 < j ? false : true);
                break;
              case "previous":
                e = k.sPrevious;
                dd = 0 < j ? false : true;
                break;
              case "next":
                e = k.sNext;
                dd = j < n - 1 ? false : true;
                break;
              case "last":
                e = k.sLast;
                dd = c + (j < n - 1 ? false : true);
                break;
              default:
                e = c + 1, g = j === c ? " is-current" : "";
                dd = false;
            }
            e && (i = b("<li>", {
              id: 0 === r && "string" === typeof c ? a.sTableId + "_" + c : null
            }).append(b("<a>", {
              "class": s.sPageButton + " " + g,
              href: "#",
              "aria-controls": a.sTableId,
              "aria-label": t[c],
              "data-dt-idx": p,
              tabindex: a.iTabIndex,
              disabled: dd
            }).html(e)).appendTo(d), a.oApi._fnBindAction(i, {
              action: c
            }, m), p++)
          }
      },
      i;
    try {
      i = b(h).find(d.activeElement).data("dt-idx")
    } catch (u) {}
    q(b(h).empty().html('<ul class="pagination-list"/>').children("ul"), m);
    i && b(h).find("[data-dt-idx=" + i + "]").focus()
  };
  return f
});
