import { create } from 'zustand';
import { persist } from 'zustand/middleware';

interface SidebarState {
  /** Desktop sidebar collapsed to an icons-only rail. Persisted. */
  collapsed: boolean;
  toggle: () => void;
  setCollapsed: (v: boolean) => void;
}

export const useSidebarStore = create<SidebarState>()(
  persist(
    (set) => ({
      collapsed: false,
      toggle: () => set((s) => ({ collapsed: !s.collapsed })),
      setCollapsed: (v) => set({ collapsed: v }),
    }),
    { name: 'finance-sh-sidebar' },
  ),
);
