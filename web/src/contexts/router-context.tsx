/* eslint-disable react-refresh/only-export-components */
import { createContext, useContext } from 'react'
import type { RouteMenuConfig, NavRoute } from '@/utils/router'

interface RouterContextValue {
  getBreadcrumbs: (path: string) => { title: string; path: string }[]
  getMetaByPath: (path: string) => RouteMenuConfig['meta'] | undefined
  navRoute: NavRoute
}

export const RouterContext = createContext<RouterContextValue | undefined>(undefined)

export const RouterProvider = RouterContext.Provider

export function useRouterContext() {
  const context = useContext(RouterContext)
  if (!context) {
    throw new Error('useRouterContext must be used within RouterProvider')
  }
  return context
}
