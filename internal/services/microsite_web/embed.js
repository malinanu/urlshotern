(function () {
  var ok = /^https:\/\/(www\.youtube-nocookie\.com\/embed\/|player\.vimeo\.com\/video\/|open\.spotify\.com\/embed\/|www\.openstreetmap\.org\/export\/embed\.html\?)/;
  document.addEventListener("click", function (e) {
    var a = e.target && e.target.closest ? e.target.closest("a.embed-facade[data-embed]") : null;
    if (!a) return;
    var src = a.getAttribute("data-embed");
    if (!ok.test(src)) return;
    e.preventDefault();
    var f = document.createElement("iframe");
    f.src = src;
    f.title = a.getAttribute("data-title") || "Embedded media";
    f.setAttribute("allow", "autoplay; encrypted-media; picture-in-picture; fullscreen; clipboard-write");
    f.setAttribute("loading", "lazy");
    f.setAttribute("referrerpolicy", "strict-origin-when-cross-origin");
    f.className = "embed-frame";
    var h = a.getAttribute("data-h");
    if (h) f.style.height = h + "px"; else f.style.aspectRatio = a.getAttribute("data-ratio") || "16/9";
    a.replaceWith(f);
  });
})();
