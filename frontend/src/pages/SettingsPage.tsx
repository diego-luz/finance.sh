import { useTranslation } from 'react-i18next';
import {
  User as UserIcon,
  Palette,
  LogOut,
  ShieldAlert,
  Check,
  Mail,
  Download,
  Star,
  EyeOff,
  Languages,
  ExternalLink,
} from 'lucide-react';
import { PageHeader } from '@/components/PageHeader';
import { LanguageSelect } from '@/components/LanguageSelect';
import {
  Card,
  CardHeader,
  Button,
  Avatar,
  Badge,
  ThemeToggle,
  PrivacyToggle,
} from '@/components/ui';
import { ExportMenu } from '@/components/ExportMenu';
import { TwoFactorCard } from '@/components/settings/TwoFactorCard';
import { ChangePasswordCard } from '@/components/settings/ChangePasswordCard';
import { ImportCard } from '@/components/settings/ImportCard';
import { OrganizationCard } from '@/components/settings/OrganizationCard';
import { CreateOrgCard } from '@/components/settings/CreateOrgCard';
import { PrivacyDataCard } from '@/components/settings/PrivacyDataCard';
import { SessionsCard } from '@/components/settings/SessionsCard';
import { useAuthStore } from '@/stores/authStore';
import { useThemeStore } from '@/stores/themeStore';
import { usePrivacyStore } from '@/stores/privacyStore';
import { useLogout, useResendVerification } from '@/hooks';

const GITHUB_URL = 'https://github.com/finance-sh/finance-sh';

/**
 * Static informational card: finance.sh is fully open-source (AGPL-3.0),
 * self-hosted, with no tiers or paid editions.
 */
function EditionCard() {
  const { t } = useTranslation();
  return (
    <Card>
      <CardHeader
        eyebrow={t('settings.edition.eyebrow')}
        title={t('settings.edition.title')}
      />
      <div className="space-y-4">
        <div>
          <Badge variant="success" dot>
            {t('settings.edition.license')}
          </Badge>
        </div>
        <p className="text-sm text-gray-600 dark:text-gray-300">
          {t('settings.edition.description')}
        </p>
        <a
          href={GITHUB_URL}
          target="_blank"
          rel="noreferrer"
          className="inline-flex items-center gap-1.5 rounded-lg border border-gray-200 px-3 py-1.5 text-sm font-medium text-gray-700 transition hover:border-primary hover:text-primary dark:border-ink-border dark:text-gray-200 dark:hover:border-primary dark:hover:text-primary"
        >
          <Star className="h-4 w-4" /> {t('settings.edition.github')}
          <ExternalLink className="h-3.5 w-3.5 opacity-60" />
        </a>
      </div>
    </Card>
  );
}

export function SettingsPage() {
  const { t } = useTranslation();
  const user = useAuthStore((s) => s.user);
  const theme = useThemeStore((s) => s.theme);
  const valuesHidden = usePrivacyStore((s) => s.hidden);
  const logout = useLogout();
  const resendVerification = useResendVerification();

  return (
    <div>
      <PageHeader
        eyebrow={t('settings.eyebrow')}
        title={t('settings.title')}
        description={t('settings.description')}
      />

      <div className="grid grid-cols-1 gap-4 lg:grid-cols-2">
        {/* Profile */}
        <Card>
          <CardHeader
            eyebrow={t('settings.profile.eyebrow')}
            title={t('settings.profile.title')}
            subtitle={t('settings.profile.subtitle')}
          />
          <div className="flex items-center gap-4">
            <Avatar name={user?.name} src={user?.avatar_url} size="lg" />
            <div className="min-w-0">
              <p className="font-heading text-lg font-semibold text-gray-900 dark:text-gray-100">
                {user?.name}
              </p>
              <p className="flex items-center gap-1.5 text-sm text-gray-500 dark:text-gray-400">
                <Mail className="h-3.5 w-3.5" /> {user?.email}
              </p>
            </div>
          </div>
          <div className="mt-4 flex flex-wrap items-center gap-2">
            {user?.email_verified ? (
              <Badge variant="success" dot>
                <Check className="h-3 w-3" /> {t('settings.profile.emailVerified')}
              </Badge>
            ) : (
              <>
                <Badge variant="warning" dot>
                  {t('settings.profile.emailNotVerified')}
                </Badge>
                <Button
                  variant="ghost"
                  size="sm"
                  loading={resendVerification.isPending}
                  onClick={() => user?.email && resendVerification.mutate(user.email)}
                >
                  {t('settings.profile.resendVerification')}
                </Button>
              </>
            )}
          </div>
          <div className="mt-4">
            <Button variant="outline" size="sm">
              <UserIcon className="h-4 w-4" /> {t('settings.profile.editProfile')}
            </Button>
          </div>
        </Card>

        {/* Organization (editable for owner/admin) */}
        <OrganizationCard />

        {/* Create an additional organization (Casa + Microempresa) */}
        <CreateOrgCard />

        {/* Appearance */}
        <Card>
          <CardHeader
            eyebrow={t('settings.appearance.eyebrow')}
            title={t('settings.appearance.title')}
            subtitle={t('settings.appearance.subtitle')}
          />
          <div className="space-y-3">
            <div className="flex items-center justify-between rounded-lg border border-gray-100 px-4 py-3 dark:border-ink-border">
              <div className="flex items-center gap-3">
                <span className="flex h-9 w-9 items-center justify-center rounded-lg bg-primary/10 text-primary">
                  <Palette className="h-[18px] w-[18px]" />
                </span>
                <div>
                  <p className="text-sm font-medium text-gray-900 dark:text-gray-100">
                    {theme === 'dark'
                      ? t('settings.appearance.darkMode')
                      : t('settings.appearance.lightMode')}
                  </p>
                  <p className="text-xs text-gray-400">{t('settings.appearance.themeSavedHint')}</p>
                </div>
              </div>
              <ThemeToggle />
            </div>
            <div className="flex items-center justify-between rounded-lg border border-gray-100 px-4 py-3 dark:border-ink-border">
              <div className="flex items-center gap-3">
                <span className="flex h-9 w-9 items-center justify-center rounded-lg bg-primary/10 text-primary">
                  <EyeOff className="h-[18px] w-[18px]" />
                </span>
                <div>
                  <p className="text-sm font-medium text-gray-900 dark:text-gray-100">
                    {valuesHidden
                      ? t('settings.appearance.valuesHidden')
                      : t('settings.appearance.valuesVisible')}
                  </p>
                  <p className="text-xs text-gray-400">
                    {t('settings.appearance.valuesHint')}
                  </p>
                </div>
              </div>
              <PrivacyToggle />
            </div>
            <div className="flex items-center justify-between gap-4 rounded-lg border border-gray-100 px-4 py-3 dark:border-ink-border">
              <div className="flex items-center gap-3">
                <span className="flex h-9 w-9 items-center justify-center rounded-lg bg-primary/10 text-primary">
                  <Languages className="h-[18px] w-[18px]" />
                </span>
                <div>
                  <p className="text-sm font-medium text-gray-900 dark:text-gray-100">
                    {t('language.label')}
                  </p>
                  <p className="text-xs text-gray-400">{t('language.hint')}</p>
                </div>
              </div>
              <div className="w-40 shrink-0">
                <LanguageSelect />
              </div>
            </div>
          </div>
        </Card>

        {/* Edition (replaces the legacy plan / subscription card) */}
        <EditionCard />
      </div>

      {/* Two-factor authentication */}
      <div className="mt-4">
        <TwoFactorCard />
      </div>

      {/* Self-service password change */}
      <ChangePasswordCard />

      {/* Active sessions */}
      <SessionsCard />

      {/* Data export */}
      <Card className="mt-4">
        <CardHeader
          eyebrow={t('settings.export.eyebrow')}
          title={t('settings.export.title')}
          subtitle={t('settings.export.subtitle')}
          action={<Download className="h-5 w-5 text-gray-400" />}
        />
        <ExportMenu />
      </Card>

      {/* Data import (restore an export into a new org) */}
      <ImportCard />

      {/* Privacy & data (LGPD) */}
      <PrivacyDataCard />

      {/* Danger zone */}
      <Card className="mt-4 border-red-200 dark:border-red-500/30">
        <CardHeader
          eyebrow={t('settings.danger.eyebrow')}
          title={t('settings.danger.title')}
          subtitle={t('settings.danger.subtitle')}
          action={<ShieldAlert className="h-5 w-5 text-red-500" />}
        />
        <Button variant="danger" onClick={() => logout.mutate()} loading={logout.isPending}>
          <LogOut className="h-4 w-4" /> {t('settings.danger.logout')}
        </Button>
      </Card>
    </div>
  );
}
