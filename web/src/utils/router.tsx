/* eslint-disable react-refresh/only-export-components */
import { Route, Router, useLocation } from 'preact-iso'
import MainLayout from '@/components/main-layout'
import PageErrorBoundary from '@/components/page-error-boundary'
import GuardRoute, { AuthGuard } from '@/components/guard-route'
import { useEffect, type ComponentType, type ReactNode } from 'react'

function PageErrorBoundaryWrapper({ children, pageName }: { children: ReactNode; pageName: string }) {
  return (
    <PageErrorBoundary pageName={pageName}>
      {children}
    </PageErrorBoundary>
  )
}

function RootRedirect() {
  const location = useLocation()

  useEffect(() => {
    location.route('/server', true)
  }, [location])

  return null
}

interface BaseMeta {
  title: string
  icon?: ReactNode
  auth?: string
  order?: number
}

export interface PageMeta extends BaseMeta {
  cache?: boolean
  showInSidebar?: boolean
  showInBreadcrumb?: boolean
  disableErrorBoundary?: boolean
}

export type DirMeta = BaseMeta

interface BaseRouteConfig {
  path: string
  children?: RouteMenuConfig[]
}

export interface RouteMenuConfig extends BaseRouteConfig {
  component?: ComponentType
  meta: PageMeta | DirMeta
}

export interface IsoRouteConfig {
  path?: string
  component: ComponentType
  default?: boolean
}

export type NavRoute = {
  path: string
  title?: string
  icon?: ReactNode
  auth?: string
  children?: NavRoute[]
}

interface DefineRouterReturn {
  AppRoutes: ComponentType
  getMetaByPath: (path: string) => RouteMenuConfig['meta'] | undefined
  getBreadcrumbs: (path: string) => { title: string; path: string }[]
  navRoute: NavRoute
}

export function defineRouter(
  configs: RouteMenuConfig[],
  additionalRoutes: IsoRouteConfig[] = []
): DefineRouterReturn {
  const pathToMetaMap = new Map<string, RouteMenuConfig['meta']>()
  const pathToConfigMap = new Map<string, RouteMenuConfig>()

  const processConfigs = (configList: RouteMenuConfig[], parentPath = ''): {
    routes: IsoRouteConfig[]
    navRoutes: NavRoute[]
  } => {
    const routes: IsoRouteConfig[] = []
    const navRoutes: NavRoute[] = []

    configList
      .sort((a, b) => (a.meta.order || 0) - (b.meta.order || 0))
      .forEach(config => {
        const fullPath = parentPath ? `${parentPath}${config.path}` : config.path

        pathToMetaMap.set(fullPath, config.meta)
        pathToConfigMap.set(fullPath, config)

        if (config.component) {
          const PageComponent = config.component
          const PageRoute = () => {
            const pageContent = <PageComponent />
            const guardedContent = config.meta.auth === 'admin' ? (
              <GuardRoute config={{ requireAuth: true, requiredRoles: ['admin'] }}>
                {pageContent}
              </GuardRoute>
            ) : pageContent

            return (
              (config.meta as PageMeta).disableErrorBoundary ? guardedContent : (
                <PageErrorBoundaryWrapper pageName={config.meta.title}>
                  {guardedContent}
                </PageErrorBoundaryWrapper>
              )
            )
          }

          routes.push({
            path: fullPath,
            component: PageRoute,
          })
        }

        const navRoute: NavRoute = {
          path: fullPath,
          title: config.meta.title,
          icon: config.meta.icon,
          auth: config.meta.auth,
        }

        if (config.children && config.children.length > 0) {
          const childResults = processConfigs(config.children, fullPath)
          routes.push(...childResults.routes)
          navRoute.children = childResults.navRoutes
        }

        const isDirectory = config.children && config.children.length > 0
        const shouldShowInSidebar = isDirectory || (config.meta as PageMeta).showInSidebar !== false
        if (shouldShowInSidebar) {
          navRoutes.push(navRoute)
        }
      })

    return { routes, navRoutes }
  }

  const { routes, navRoutes } = processConfigs(configs)

  const navRoute: NavRoute = {
    path: '/',
    children: navRoutes,
  }

  const publicRoutes = additionalRoutes.filter(routeConfig => !routeConfig.default)
  const defaultRoute = additionalRoutes.find(routeConfig => routeConfig.default)

  function ProtectedAppRoutes() {
    return (
      <AuthGuard>
        <MainLayout>
          <Router>
            <Route path="/" component={RootRedirect} />
            {routes.map((routeConfig) => (
              <Route
                key={routeConfig.path}
                path={routeConfig.path!}
                component={routeConfig.component}
              />
            ))}
            {defaultRoute ? <Route key="protected-default" default component={defaultRoute.component} /> : <></>}
          </Router>
        </MainLayout>
      </AuthGuard>
    )
  }

  function AppRoutes() {
    return (
      <Router>
        {publicRoutes.map((routeConfig) => (
          <Route
            key={routeConfig.path}
            path={routeConfig.path!}
            component={routeConfig.component}
          />
        ))}
        <Route path="/" component={ProtectedAppRoutes} />
        <Route path="/*" component={ProtectedAppRoutes} />
        {defaultRoute ? <Route key="default" default component={defaultRoute.component} /> : <></>}
      </Router>
    )
  }

  const getMetaByPath = (path: string) => pathToMetaMap.get(path)

  const getBreadcrumbs = (path: string): { title: string; path: string }[] => {
    const breadcrumbs: { title: string; path: string }[] = []
    const pathSegments = path.split('/').filter(Boolean)

    let currentPath = ''
    for (const segment of pathSegments) {
      currentPath += `/${segment}`
      const config = pathToConfigMap.get(currentPath)

      if (config) {
        const isDirectory = config.children && config.children.length > 0
        const shouldShowInBreadcrumb = isDirectory || (config.meta as PageMeta).showInBreadcrumb !== false
        if (shouldShowInBreadcrumb) {
          breadcrumbs.push({
            title: config.meta.title,
            path: currentPath,
          })
        }
      }
    }

    return breadcrumbs
  }

  return {
    AppRoutes,
    getMetaByPath,
    getBreadcrumbs,
    navRoute
  }
}
