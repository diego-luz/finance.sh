import { useTranslation } from 'react-i18next';
import { Select } from '@/components/ui';
import { SUPPORTED_LANGUAGES } from '@/i18n';

interface LanguageSelectProps {
  label?: string;
  className?: string;
}

/**
 * Dropdown to switch the UI language. Calling `i18n.changeLanguage` triggers a
 * re-render of every component using `useTranslation`, and the detector caches
 * the choice to localStorage (`finance-sh:lang`) so it persists across reloads.
 */
export function LanguageSelect({ label, className }: LanguageSelectProps) {
  const { i18n, t } = useTranslation();

  // i18n.language may be a regional variant (e.g. "pt"); resolvedLanguage is the
  // one actually selected from supportedLngs, so it matches our option values.
  const current = (i18n.resolvedLanguage ?? i18n.language) as string;

  const options = SUPPORTED_LANGUAGES.map((code) => ({
    value: code,
    label: t(`language.${code}`),
  }));

  const value = SUPPORTED_LANGUAGES.some((code) => code === current) ? current : 'pt-BR';

  return (
    <Select
      label={label}
      aria-label={t('language.label')}
      options={options}
      value={value}
      onChange={(e) => void i18n.changeLanguage(e.target.value)}
      className={className}
    />
  );
}
