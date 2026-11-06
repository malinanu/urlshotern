(function () {
  var els = document.querySelectorAll("[data-countdown]");
  if (!els.length) return;
  function pad(n) { return n < 10 ? "0" + n : "" + n; }
  function tick() {
    var now = Date.now();
    els.forEach(function (el) {
      var t = Date.parse(el.getAttribute("data-countdown"));
      if (isNaN(t)) return;
      var s = Math.max(0, Math.floor((t - now) / 1000));
      var v = el.querySelector(".cd-value");
      if (!v) return;
      if (s === 0) { v.textContent = el.getAttribute("data-done") || ""; v.className = "cd-done"; return; }
      var set = function (u, n) { var x = el.querySelector('[data-u="' + u + '"]'); if (x) x.textContent = n; };
      set("d", Math.floor(s / 86400)); set("h", pad(Math.floor(s / 3600) % 24));
      set("m", pad(Math.floor(s / 60) % 60)); set("s", pad(s % 60));
      el.classList.add("cd-live");
    });
  }
  tick();
  setInterval(tick, 1000);
})();
