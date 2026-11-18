/* Digital card downloads: PNG, JPG and QR code. Loaded only on card pages. */
(function () {
  'use strict';
  var W = 600, H = 1050;
  var bar = document.querySelector('.vc-actions');
  var card = document.querySelector('.vc[data-vc]');
  if (!bar || !card) return;
  bar.hidden = false;
  var slug = bar.getAttribute('data-slug') || 'card';

  var msg = document.createElement('p');
  msg.className = 'vc-msg';
  msg.setAttribute('role', 'alert');
  bar.appendChild(msg);

  function loadLib() {
    return new Promise(function (resolve, reject) {
      if (window.htmlToImage) return resolve();
      var s = document.createElement('script');
      s.src = '/m-static/html-to-image.js';
      s.onload = resolve;
      s.onerror = function () { reject(new Error('Could not load the image tool')); };
      document.head.appendChild(s);
    });
  }

  function settle(root) {
    var waits = [];
    if (document.fonts && document.fonts.ready) waits.push(document.fonts.ready);
    root.querySelectorAll('img').forEach(function (img) {
      if (!img.complete) {
        waits.push(new Promise(function (r) { img.onload = img.onerror = r; }));
      }
    });
    return Promise.all(waits);
  }

  function save(url, name) {
    var a = document.createElement('a');
    a.href = url;
    a.download = name;
    document.body.appendChild(a);
    a.click();
    a.remove();
  }

  // The card is copied off-screen at exactly 600px wide so its em-based sizing resolves at design size.
  function cardImage(format) {
    return loadLib().then(function () {
      var host = document.createElement('div');
      host.setAttribute('aria-hidden', 'true');
      host.style.cssText = 'position:fixed;left:-10000px;top:0;width:' + W + 'px;pointer-events:none';
      var clone = card.cloneNode(true);
      clone.style.width = W + 'px';
      clone.style.maxWidth = 'none';
      host.appendChild(clone);
      document.body.appendChild(host);
      var opts = { width: W, height: H, pixelRatio: 2, cacheBust: false, quality: 0.95 };
      if (format === 'jpg') opts.backgroundColor = '#ffffff';
      var render = format === 'png' ? window.htmlToImage.toPng : window.htmlToImage.toJpeg;
      var safari = /^((?!chrome|android).)*safari/i.test(navigator.userAgent);
      return settle(clone)
        .then(function () { return safari ? render(clone, opts) : null; })
        .then(function () { return render(clone, opts); })
        .then(function (url) { host.remove(); return url; }, function (e) { host.remove(); throw e; });
    });
  }

  // The QR code is drawn from the card's own SVG, dark modules on white, at 1024px.
  function qrImage() {
    var svg = card.querySelector('.vc-qr svg');
    var path = svg && svg.querySelector('path');
    if (!svg || !path) return Promise.reject(new Error('No QR code'));
    var vb = (svg.getAttribute('viewBox') || '').split(' ').map(Number);
    var xml = '<svg xmlns="http://www.w3.org/2000/svg" viewBox="' + vb.join(' ') + '" shape-rendering="crispEdges">' +
      '<rect x="' + vb[0] + '" y="' + vb[1] + '" width="' + vb[2] + '" height="' + vb[3] + '" fill="#ffffff"/>' +
      '<path fill="#000000" d="' + path.getAttribute('d') + '"/></svg>';
    return new Promise(function (resolve, reject) {
      var img = new Image();
      img.onload = function () {
        var c = document.createElement('canvas');
        c.width = c.height = 1024;
        var g = c.getContext('2d');
        g.imageSmoothingEnabled = false;
        g.drawImage(img, 0, 0, 1024, 1024);
        resolve(c.toDataURL('image/png'));
      };
      img.onerror = function () { reject(new Error('Could not draw the QR code')); };
      img.src = 'data:image/svg+xml;charset=utf-8,' + encodeURIComponent(xml);
    });
  }

  bar.addEventListener('click', function (e) {
    var btn = e.target.closest('button[data-dl]');
    if (!btn || btn.disabled) return;
    var kind = btn.getAttribute('data-dl');
    var label = btn.textContent;
    btn.disabled = true;
    btn.textContent = 'Preparing…';
    msg.textContent = '';
    var job = kind === 'qr' ? qrImage().then(function (u) { save(u, slug + '-qr.png'); })
      : cardImage(kind).then(function (u) { save(u, slug + '.' + (kind === 'png' ? 'png' : 'jpg')); });
    job.catch(function () {
      msg.textContent = 'Could not create the image. If the card uses an image from another site, upload it instead.';
    }).then(function () {
      btn.disabled = false;
      btn.textContent = label;
    });
  });
})();
