/**
 * Robust copy-to-clipboard. The modern `navigator.clipboard` API only works in
 * "secure contexts" (HTTPS or localhost/127.0.0.1). On a LAN-IP over plain HTTP
 * it silently rejects. So we try the modern API first, then fall back to the
 * legacy `document.execCommand('copy')` via a hidden textarea — which works
 * everywhere (deprecated but still supported by every browser).
 *
 * Returns true on success.
 */
export async function copyToClipboard(text: string): Promise<boolean> {
  if (!text) return false;

  // Path 1: modern, secure-context only.
  try {
    if (navigator.clipboard && window.isSecureContext) {
      await navigator.clipboard.writeText(text);
      return true;
    }
  } catch {
    // fall through to legacy
  }

  // Path 2: legacy textarea + execCommand('copy'). Works in plain HTTP.
  try {
    const ta = document.createElement('textarea');
    ta.value = text;
    // Keep it off-screen and non-interactive but still selectable.
    ta.setAttribute('readonly', '');
    ta.style.position = 'fixed';
    ta.style.top = '0';
    ta.style.left = '0';
    ta.style.opacity = '0';
    ta.style.pointerEvents = 'none';
    document.body.appendChild(ta);
    ta.focus();
    ta.select();
    ta.setSelectionRange(0, text.length);
    const ok = document.execCommand('copy');
    document.body.removeChild(ta);
    return ok;
  } catch {
    return false;
  }
}
