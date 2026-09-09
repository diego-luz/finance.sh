import type { FieldValues, Path, UseFormReturn } from 'react-hook-form';
import { ApiRequestError } from '@/lib/axios';

/** The slice of a react-hook-form instance that {@link applyApiFieldErrors} needs. */
export type FieldErrorTarget<T extends FieldValues> = Pick<
  UseFormReturn<T>,
  'setError' | 'getValues'
>;

/**
 * Attaches the backend's per-field validation messages to the matching inputs.
 *
 * The API returns `error.fields` keyed by the payload's JSON names
 * (`closing_day`, `account_id`, ...), which is what forms register their inputs
 * under. Keys the form does not have are skipped on purpose: setting an error on
 * an unregistered name leaves an error object no input renders and the user
 * cannot clear, which would silently block resubmission.
 *
 * Returns true when at least one message was attached, so callers can tell a
 * field-level rejection from a generic one.
 */
export function applyApiFieldErrors<T extends FieldValues>(
  error: unknown,
  form: FieldErrorTarget<T>,
): boolean {
  if (!(error instanceof ApiRequestError) || !error.fields) return false;

  const known = new Set(Object.keys(form.getValues()));
  let applied = false;

  for (const [field, message] of Object.entries(error.fields)) {
    if (!message || !known.has(field)) continue;
    form.setError(field as Path<T>, { type: 'server', message });
    applied = true;
  }

  return applied;
}
