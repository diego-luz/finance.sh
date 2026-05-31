import {
  Landmark,
  Wallet,
  TrendingUp,
  CreditCard,
  ShoppingCart,
  Utensils,
  Car,
  Home,
  Heart,
  Gift,
  Plane,
  Briefcase,
  GraduationCap,
  Zap,
  Film,
  Dumbbell,
  Coins,
  PiggyBank,
  Receipt,
  Smartphone,
  type LucideIcon,
} from 'lucide-react';
import type { AccountType } from '@/types';

/** Map account types to a representative icon + label. */
export const accountTypeMeta: Record<AccountType, { icon: LucideIcon; label: string }> = {
  bank: { icon: Landmark, label: 'Conta bancária' },
  wallet: { icon: Wallet, label: 'Carteira' },
  investment: { icon: TrendingUp, label: 'Investimento' },
  credit_card: { icon: CreditCard, label: 'Cartão de crédito' },
};

/** Named icon set for categories/accounts custom icons. */
export const iconMap: Record<string, LucideIcon> = {
  landmark: Landmark,
  wallet: Wallet,
  'trending-up': TrendingUp,
  'credit-card': CreditCard,
  cart: ShoppingCart,
  food: Utensils,
  car: Car,
  home: Home,
  health: Heart,
  gift: Gift,
  travel: Plane,
  work: Briefcase,
  education: GraduationCap,
  utilities: Zap,
  entertainment: Film,
  fitness: Dumbbell,
  coins: Coins,
  savings: PiggyBank,
  receipt: Receipt,
  phone: Smartphone,
};

export const iconOptions = Object.keys(iconMap);

/** Resolve a named icon, falling back to a generic coins icon. */
export function resolveIcon(name?: string): LucideIcon {
  if (name && iconMap[name]) return iconMap[name];
  return Coins;
}
