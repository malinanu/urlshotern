// Copies text for the background worker (which has no DOM).
chrome.runtime.onMessage.addListener((msg) => {
  if (msg?.target !== 'offscreen' || msg.type !== 'copy') return;
  const area = document.getElementById('buffer');
  area.value = msg.text;
  area.select();
  document.execCommand('copy');
});
