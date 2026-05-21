'use client';

import { useMemo } from 'react';
import { useAuthStore } from '@/stores/auth-store';
import type { NavItem, NavGroup } from '@/types';

export function useFilteredNavItems(items: NavItem[]) {
  const user = useAuthStore((s) => s.user);

  const filteredItems = useMemo(() => {
    return items
      .filter((item) => {
        if (!item.access) {
          return true;
        }

        if (item.access.role) {
          if (user?.role !== item.access.role) {
            return false;
          }
        }

        return true;
      })
      .map((item) => {
        if (item.items && item.items.length > 0) {
          const filteredChildren = item.items.filter((childItem) => {
            if (!childItem.access) {
              return true;
            }
            if (childItem.access.role) {
              if (user?.role !== childItem.access.role) {
                return false;
              }
            }
            return true;
          });

          return {
            ...item,
            items: filteredChildren,
          };
        }

        return item;
      });
  }, [items, user]);

  return filteredItems;
}

export function useFilteredNavGroups(groups: NavGroup[]) {
  const allItems = useMemo(
    () => groups.flatMap((g) => g.items),
    [groups]
  );
  const filteredItems = useFilteredNavItems(allItems);

  return useMemo(() => {
    const filteredSet = new Set(filteredItems.map((item) => item.title));
    return groups
      .map((group) => ({
        ...group,
        items: filteredItems.filter((item) =>
          group.items.some(
            (gi) => gi.title === item.title && filteredSet.has(gi.title)
          )
        ),
      }))
      .filter((group) => group.items.length > 0);
  }, [groups, filteredItems]);
}
