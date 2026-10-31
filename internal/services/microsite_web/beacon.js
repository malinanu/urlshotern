(function () {
  var s = document.currentScript;
  var ep = s && s.getAttribute("data-endpoint");
  if (!ep) return;
  document.addEventListener("click", function (e) {
    var el = e.target && e.target.closest ? e.target.closest("[data-track]") : null;
    if (!el) return;
    var body = JSON.stringify({
      type: el.getAttribute("data-track"),
      block_id: el.getAttribute("data-block") || "",
      item_id: el.getAttribute("data-item") || ""
    });
    try {
      if (navigator.sendBeacon) {
        navigator.sendBeacon(ep, new Blob([body], { type: "text/plain" }));
      } else {
        fetch(ep, { method: "POST", body: body, keepalive: true, headers: { "Content-Type": "text/plain" } });
      }
    } catch (_) {}
  }, true);
})();
