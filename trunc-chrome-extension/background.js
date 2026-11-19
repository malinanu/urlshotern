// Background service worker: right-click "Shorten with Trunc", and Sign in
// with Trunc (run here because the popup closes when the sign-in window opens).
import { NotConnectedError, isShortenable, shorten, signInWithTrunc } from './lib/trunc.js';

const MENU_ID = 'trunc-shorten';

chrome.runtime.onInstalled.addListener((details) => {
  chrome.contextMenus.create({ id: MENU_ID, title: 'Shorten with Trunc', contexts: ['page', 'link'] });
  if (details.reason === 'install') chrome.runtime.openOptionsPage();
});

async function flashBadge(text, color) {
  await chrome.action.setBadgeBackgroundColor({ color });
  await chrome.action.setBadgeText({ text });
  setTimeout(() => chrome.action.setBadgeText({ text: '' }), 2500);
}

// Service workers have no DOM, so copying goes through an offscreen document.
async function copyText(text) {
  const has = await chrome.offscreen.hasDocument?.();
  if (!has) {
    await chrome.offscreen.createDocument({
      url: 'offscreen.html',
      reasons: [chrome.offscreen.Reason.CLIPBOARD],
      justification: 'Copy the new short link to the clipboard.',
    });
  }
  await chrome.runtime.sendMessage({ target: 'offscreen', type: 'copy', text });
}

chrome.contextMenus.onClicked.addListener(async (info, tab) => {
  if (info.menuItemId !== MENU_ID) return;
  const url = info.linkUrl || info.pageUrl || tab?.url || '';
  if (!isShortenable(url)) {
    await flashBadge('!', '#dc2626');
    return;
  }
  try {
    const { shortUrl } = await shorten(url);
    await copyText(shortUrl);
    await chrome.storage.session.set({ lastResult: { url, shortUrl, at: Date.now() } });
    await flashBadge('✓', '#16a34a');
  } catch (e) {
    if (e instanceof NotConnectedError) {
      chrome.runtime.openOptionsPage();
      return;
    }
    await chrome.storage.session.set({ lastError: { url, message: e.message, at: Date.now() } });
    await flashBadge('!', '#dc2626');
  }
});

chrome.runtime.onMessage.addListener((msg, _sender, sendResponse) => {
  if (msg?.type === 'signIn') {
    signInWithTrunc()
      .then(() => sendResponse({ ok: true }))
      .catch((e) => sendResponse({ ok: false, error: e.message }));
    return true; // keep the channel open for the async answer
  }
  return false;
});
