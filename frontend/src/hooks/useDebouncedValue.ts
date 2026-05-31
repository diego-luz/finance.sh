import { useEffect, useState } from 'react';

/**
 * Returns a debounced copy of `value` that only updates after `delay` ms have
 * passed without further changes. Useful to throttle network calls tied to
 * fast-changing inputs (e.g. a text field driving an autocomplete request).
 */
export function useDebouncedValue<T>(value: T, delay = 400): T {
  const [debounced, setDebounced] = useState(value);

  useEffect(() => {
    const id = window.setTimeout(() => setDebounced(value), delay);
    return () => window.clearTimeout(id);
  }, [value, delay]);

  return debounced;
}
