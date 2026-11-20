import qrcode from './vendor/qrcode.mjs';
import { NotConnectedError, getSettings, isShortenable, recentLinks, shorten } from './lib/trunc.js';

const app = document.getElementById('app');
const SVG_NS = 'http://www.w3.org/2000/svg';

function show(templateId) {
  const node = document.getElementById(templateId).content.cloneNode(true);
  app.replaceChildren(node);
  return app;
}

function showMessage(text, plain = false) {
  show('tpl-message');
  const p = app.querySelector('.message');
  p.textContent = text;
  p.classList.toggle('plain', plain);
}

/** The active tab of the last focused browser window (works as a popup and in tests). */
async function activeTab() {
  const win = await chrome.windows.getLastFocused({ windowTypes: ['normal'] }).catch(() => null);
  const [tab] = await chrome.tabs.query(win ? { active: true, windowId: win.id } : { active: true, currentWindow: true });
  return tab;
}

function qrModules(text) {
  const qr = qrcode(0, 'M');
  qr.addData(text);
  qr.make();
  return qr;
}

/** Draws the QR code as one SVG path with a 4-module quiet zone. */
function drawQr(svg, text) {
  const qr = qrModules(text);
  const n = qr.getModuleCount();
  let d = '';
  for (let y = 0; y < n; y++) {
    for (let x = 0; x < n; x++) if (qr.isDark(y, x)) d += `M${x} ${y}h1v1h-1z`;
  }
  svg.setAttribute('viewBox', `-4 -4 ${n + 8} ${n + 8}`);
  svg.setAttribute('shape-rendering', 'crispEdges');
  const bg = document.createElementNS(SVG_NS, 'rect');
  bg.setAttribute('x', '-4');
  bg.setAttribute('y', '-4');
  bg.setAttribute('width', String(n + 8));
  bg.setAttribute('height', String(n + 8));
  bg.setAttribute('fill', '#fff');
  const path = document.createElementNS(SVG_NS, 'path');
  path.setAttribute('d', d);
  path.setAttribute('fill', '#000');
  svg.replaceChildren(bg, path);
}

function downloadQr(text, code) {
  const qr = qrModules(text);
  const n = qr.getModuleCount();
  const scale = Math.floor(1024 / (n + 8));
  const canvas = document.createElement('canvas');
  canvas.width = canvas.height = (n + 8) * scale;
  const ctx = canvas.getContext('2d');
  ctx.fillStyle = '#fff';
  ctx.fillRect(0, 0, canvas.width, canvas.height);
  ctx.fillStyle = '#000';
  for (let y = 0; y < n; y++) for (let x = 0; x < n; x++) if (qr.isDark(y, x)) ctx.fillRect((x + 4) * scale, (y + 4) * scale, scale, scale);
  const a = document.createElement('a');
  a.href = canvas.toDataURL('image/png');
  a.download = `trunc-${code || 'qr'}.png`;
  a.click();
}

async function copy(text, button) {
  try {
    await navigator.clipboard.writeText(text);
    if (button) {
      button.textContent = 'Copied';
      setTimeout(() => (button.textContent = 'Copy'), 1500);
    }
    return true;
  } catch {
    return false;
  }
}

function renderResult(result) {
  show('tpl-result');
  const input = app.querySelector('#short-url');
  input.value = result.shortUrl;
  app.querySelector('.original').textContent = result.originalUrl;
  app.querySelector('.original').title = result.originalUrl;
  drawQr(app.querySelector('.qr'), result.shortUrl);
  const copyBtn = app.querySelector('[data-action="copy"]');
  copyBtn.addEventListener('click', () => copy(result.shortUrl, copyBtn));
  app.querySelector('[data-action="download-qr"]').addEventListener('click', () => downloadQr(result.shortUrl, result.shortCode));
  input.addEventListener('focus', () => input.select());
  // Copy straight away, as most people open the popup to grab the link.
  copy(result.shortUrl, copyBtn);
}

function renderConnect() {
  show('tpl-connect');
  const err = app.querySelector('.error');
  const signin = app.querySelector('[data-action="signin"]');
  signin.addEventListener('click', async () => {
    signin.disabled = true;
    signin.textContent = 'Opening sign-in…';
    err.hidden = true;
    const res = await chrome.runtime.sendMessage({ type: 'signIn' });
    if (res?.ok) {
      init();
    } else {
      signin.disabled = false;
      signin.textContent = 'Sign in with Trunc';
      err.textContent = res?.error || 'Sign-in did not finish.';
      err.hidden = false;
    }
  });
  app.querySelector('[data-action="apikey"]').addEventListener('click', () => chrome.runtime.openOptionsPage());
}

async function renderRecent() {
  const section = document.getElementById('recent');
  try {
    const links = await recentLinks(5);
    const list = section.querySelector('ul');
    list.replaceChildren(
      ...links.map((l) => {
        const li = document.createElement('li');
        const info = document.createElement('div');
        info.className = 'link';
        const short = document.createElement('span');
        short.className = 'short';
        short.textContent = l.shortUrl.replace(/^https?:\/\//, '');
        const dest = document.createElement('span');
        dest.className = 'dest';
        dest.textContent = l.originalUrl;
        dest.title = l.originalUrl;
        info.append(short, dest);
        const clicks = document.createElement('span');
        clicks.className = 'clicks';
        clicks.textContent = `${l.clicks} ${l.clicks === 1 ? 'click' : 'clicks'}`;
        const btn = document.createElement('button');
        btn.type = 'button';
        btn.className = 'secondary';
        btn.textContent = 'Copy';
        btn.setAttribute('aria-label', `Copy ${l.shortUrl}`);
        btn.addEventListener('click', () => copy(l.shortUrl, btn));
        li.append(info, clicks, btn);
        return li;
      })
    );
    section.hidden = links.length === 0;
  } catch {
    section.hidden = true;
  }
}

async function init() {
  const { apiKey, baseUrl } = await getSettings();
  document.getElementById('open-dashboard').href = `${baseUrl}/dashboard`;
  if (!apiKey) {
    renderConnect();
    return;
  }
  const tab = await activeTab();
  if (!tab || !isShortenable(tab.url)) {
    showMessage('Open a web page to shorten it. Your recent links are below.', true);
  } else {
    try {
      renderResult(await shorten(tab.url));
    } catch (e) {
      if (e instanceof NotConnectedError) return renderConnect();
      showMessage(e.message);
    }
  }
  renderRecent();
}

init();
