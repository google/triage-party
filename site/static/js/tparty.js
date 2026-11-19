/**
 * Copyright 2020 Google Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

// Encodes a DataTables order, such as [[3, "desc"], [1, "asc"]], as "3.desc-1.asc"
function encodeOrder(order) {
  return order.map(function (o) { return o[0] + "." + o[1]; }).join("-");
}

// Decodes an encoded order, dropping any part that does not refer to one of the columns
function decodeOrder(s, columns) {
  var order = [];
  s.split("-").forEach(function (part) {
    var m = /^(\d+)\.(asc|desc)$/.exec(part);
    if (m && parseInt(m[1], 10) < columns) {
      order.push([parseInt(m[1], 10), m[2]]);
    }
  });
  return order;
}

// Creates a DataTable whose sort order and filter are persisted in the page URL,
// so that reloading or sharing the page keeps them.
function persistentDataTable(id, options) {
  var sortKey = "sort." + id;
  var filterKey = "filter." + id;
  var el = $("#" + id);
  var defaultOrder = encodeOrder(options.order || []);

  var params = new URLSearchParams(window.location.search);
  if (params.has(sortKey)) {
    var order = decodeOrder(params.get(sortKey), el.find("thead tr").first().children().length);
    if (order.length > 0) {
      options.order = order;
    }
  }
  if (params.get(filterKey)) {
    options.search = { search: params.get(filterKey) };
  }

  var table = el.DataTable(options);

  table.on("order.dt search.dt", function () {
    var params = new URLSearchParams(window.location.search);

    var order = encodeOrder(table.order());
    if (order === defaultOrder) {
      params.delete(sortKey);
    } else {
      params.set(sortKey, order);
    }

    var filter = table.search();
    if (filter === "") {
      params.delete(filterKey);
    } else {
      params.set(filterKey, filter);
    }

    var query = params.toString();
    var url = window.location.pathname + (query ? "?" + query : "") + window.location.hash;
    if (url !== window.location.pathname + window.location.search + window.location.hash) {
      window.history.replaceState(null, "", url);
    }
  });

  return table;
}
