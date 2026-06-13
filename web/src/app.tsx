import { lazy, LocationProvider } from 'preact-iso'
import { Toasty } from '@cloudflare/kumo/components/toast'
import { RouterProvider } from './contexts/router-context'
import { defineRouter } from './utils/router'
import { FeedbackProvider, appToastManager } from './contexts/feedback-context'
import GlobalErrorBoundary from './components/global-error-boundary'
import AdminSwitcher from './components/admin-switcher'
import { FolderIcon, HouseIcon, UsersIcon } from '@phosphor-icons/react'

const Login = lazy(() => import('./pages/login'))
const User = lazy(() => import('./pages/admin/user'))
const Project = lazy(() => import('./pages/project'))
const Forbidden = lazy(() => import('./pages/403'))
const NotFound = lazy(() => import('./pages/404'))
const AdminDashboard = lazy(() => import('./pages/admin/dashboard'))
const MemberDashboard = lazy(() => import('./pages/member/dashboard'))

const { AppRoutes, getBreadcrumbs, getMetaByPath, navRoute } = defineRouter([
  {
    path: '/home',
    component: AdminSwitcher(AdminDashboard, MemberDashboard),
    meta: {
      title: '首页',
      icon: <HouseIcon size={18} />
    }
  },
  {
    path: '/admin/user',
    component: User,
    meta: {
      title: '用户管理',
      icon: <UsersIcon size={18} />,
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
