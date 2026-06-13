import { lazy, LocationProvider } from 'preact-iso'
import { Toasty } from '@cloudflare/kumo/components/toast'
import { RouterProvider } from './contexts/router-context'
import { defineRouter } from './utils/router'
import { FeedbackProvider, appToastManager } from './contexts/feedback-context'
import GlobalErrorBoundary from './components/global-error-boundary'
import { FolderIcon, HardDrivesIcon, PlayCircleIcon } from '@phosphor-icons/react'

const Login = lazy(() => import('./pages/login'))
const Project = lazy(() => import('./pages/project'))
const Server = lazy(() => import('./pages/server'))
const Runner = lazy(() => import('./pages/runner'))
const Forbidden = lazy(() => import('./pages/403'))
const NotFound = lazy(() => import('./pages/404'))

const { AppRoutes, getBreadcrumbs, getMetaByPath, navRoute } = defineRouter([
  {
    path: '/server',
    component: Server,
    meta: {
      title: '服务器',
      icon: <HardDrivesIcon size={18} />,
      auth: 'admin'
    }
  },
  {
    path: '/project',
    component: Project,
    meta: {
      title: '项目管理',
      icon: <FolderIcon size={18} />,
    }
  },
  {
    path: '/runner',
    component: Runner,
    meta: {
      title: 'Runner',
      icon: <PlayCircleIcon size={18} />,
      auth: 'admin'
    }
  },
], [
  {
    path: '/login',
    component: Login,
  },
  {
    path: '/403',
    component: Forbidden,
  },
  {
    path: '/404',
    component: NotFound,
  },
  {
    default: true,
    component: NotFound,
  }
])

export default function App() {
  return (
    <GlobalErrorBoundary>
      <Toasty toastManager={appToastManager}>
        <LocationProvider>
          <FeedbackProvider>
            <RouterProvider value={{ getBreadcrumbs, getMetaByPath, navRoute }}>
              <AppRoutes />
            </RouterProvider>
          </FeedbackProvider>
        </LocationProvider>
      </Toasty>
    </GlobalErrorBoundary>
  )
}
