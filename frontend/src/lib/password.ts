/**
 * Cryptographically-strong password suggester. Mixes lowercase, uppercase,
 * digits and a curated symbol set, guarantees at least one of each class, then
 * shuffles. Uses `crypto.getRandomValues` (Web Crypto) — never `Math.random`.
 */
const LOWER = 'abcdefghijkmnopqrstuvwxyz'; // no l
const UPPER = 'ABCDEFGHJKLMNPQRSTUVWXYZ'; // no I, O
const DIGITS = '23456789'; // no 0, 1
const SYMBOLS = '!@#$%&*?-_=+';

function pick(alphabet: string, n: number): string {
  const out = new Array<string>(n);
  const buf = new Uint32Array(n);
  crypto.getRandomValues(buf);
  for (let i = 0; i < n; i++) {
    out[i] = alphabet[buf[i] % alphabet.length];
  }
  return out.join('');
}

function shuffle(s: string): string {
  const arr = s.split('');
  const buf = new Uint32Array(arr.length);
  crypto.getRandomValues(buf);
  for (let i = arr.length - 1; i > 0; i--) {
    const j = buf[i] % (i + 1);
    [arr[i], arr[j]] = [arr[j], arr[i]];
  }
  return arr.join('');
}

/** Returns a strong password of the given length (default 16). */
export function suggestPassword(length = 16): string {
  const minLen = Math.max(length, 12);
  const base =
    pick(LOWER, 1) + pick(UPPER, 1) + pick(DIGITS, 1) + pick(SYMBOLS, 1) +
    pick(LOWER + UPPER + DIGITS + SYMBOLS, minLen - 4);
  return shuffle(base);
}
