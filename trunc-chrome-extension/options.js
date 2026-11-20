import { DEFAULT_BASE_URL, disconnect, getSettings, normalizeBaseUrl, savePastedKey } from './lib/trunc.js';

const $ = (id) => document.getElementById(id);

function showError(message) {
  const e = $('account-error');
  e.textContent = message || '';
  e.hidden = !message;
}

async function render() {
  const { apiKey, baseUrl, connectedVia } = await getSettings();
  const host = new URL(baseUrl).host;
  $('server-url').value = baseUrl;
  $('keys-link').href = `${baseUrl}/api-keys`;
  $('connected').hidden = !apiKey;
  $('disconnected').hidden = !!apiKey;
  $('status').className = apiKey ? 'status ok' : 'status';
  $('status').textContent = apiKey
    ? `Connected to ${host}${connectedVia === 'signin' ? ' with Sign in with Trunc' : ' with an API key'}.`
    : `Not connected to ${host}.`;
}

$('signin').addEventListener('click', async () => {
  showError('');
  $('signin').disabled = true;
  const res = await chrome.runtime.sendMessage({ type: 'signIn' });
  $('signin').disabled = false;
  if (!res?.ok) showError(res?.error || 'Sign-in did not finish.');
  render();
});

$('key-form').addEventListener('submit', async (e) => {
  e.preventDefault();
  showError('');
  try {
    await savePastedKey($('api-key').value);
    $('api-key').value = '';
  } catch (err) {
    showError(err.message);
  }
  render();
});

$('disconnect').addEventListener('click', async () => {
  await disconnect();
  render();
});

$('server-form').addEventListener('submit', async (e) => {
  e.preventDefault();
  const origin = normalizeBaseUrl($('server-url').value || DEFAULT_BASE_URL);
  const msg = $('server-msg');
  if (!origin) {
    msg.textContent = 'Use an https:// address (or http://localhost for testing).';
    return;
  }
  if (origin !== DEFAULT_BASE_URL) {
    // The extension only reaches trunc.site by default; ask for this server too.
    const granted = await chrome.permissions.request({ origins: [`${origin}/*`] }).catch(() => false);
    if (!granted) {
      msg.textContent = 'Chrome needs your permission to reach that server.';
      return;
    }
  }
  const { baseUrl } = await getSettings();
  if (origin !== baseUrl) await disconnect(); // a key belongs to one server
  await chrome.storage.local.set({ baseUrl: origin });
  msg.textContent = 'Saved.';
  render();
});

render();
