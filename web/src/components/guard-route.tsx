import { useEffect, type ReactNode } from 'react'
import { useLocation } from 'preact-iso'
import { Loader } from '@cloudflare/kumo/components/loader'
import { useAuth } from '@/hooks/use-auth'

export interface GuardConfig {
  requireAuth?: boolean
  requiredRoles?: string[]
  requiredPermissions?: string[]
  requireAllPermissions?: boolean
  redirectTo?: string
  unauthorizedRedirectTo?: string
  fallback?: ReactNode
}

interface GuardRouteProps {
  children: ReactNode
  config?: GuardConfig
}

const defaultConfig: GuardConfig = {
  requireAuth: true,
  redirectTo: '/login',
  unauthorizedRedirectTo: '/403',
  requireAllPermissions: true,
  fallback: (
    <div className="flex min-h-52 items-center justify-center">
      <Loader size="lg" aria-label="Checking authentication" />
    </div>
  )
}

function Redirect({ to, replace }: { to: string; replace?: boolean }) {
  const location = useLocation()

  useEffect(() => {
    location.route(to, replace)
  }, [location, replace, to])

  return null
}

export default function GuardRoute({ children, config = {} }: GuardRouteProps) {
  const {
    loading,
    isAuthenticated,
    hasAnyRole,
    hasAllPermissions,
    hasAnyPermission
  } = useAuth()

  const guardConfig = { ...defaultConfig, ...config }

  if (loading) {
    return <>{guardConfig.fallback}</>
  }

  if (guardConfig.requireAuth && !isAuthenticated) {
    return <Redirect to={guardConfig.redirectTo!} replace />
  }

  if (guardConfig.requiredRoles && guardConfig.requiredRoles.length > 0) {
    if (!hasAnyRole(guardConfig.requiredRoles)) {
      return <Redirect to={guardConfig.unauthorizedRedirectTo!} replace />
    }
  }

  if (guardConfig.requiredPermissions && guardConfig.requiredPermissions.length > 0) {
    const hasPermissions = guardConfig.requireAllPermissions
      ? hasAllPermissions(guardConfig.requiredPermissions)
      : hasAnyPermission(guardConfig.requiredPermissions)

    if (!hasPermissions) {
      return <Redirect to={guardConfig.unauthorizedRedirectTo!} replace />
    }
  }

  return <>{children}</>
}

export function AuthGuard({ children }: { children: ReactNode }) {
  return (
    <GuardRoute config={{ requireAuth: true }}>
      {children}
    </GuardRoute>
  )
}

export function AdminGuard({ children }: { children: ReactNode }) {
  return (
    <GuardRoute config={{ requireAuth: true, requiredRoles: ['admin'] }}>
      {children}
    </GuardRoute>
  )
}

export function RoleGuard({ children, roles }: { children: ReactNode; roles: string[] }) {
  return (
    <GuardRoute config={{ requireAuth: true, requiredRoles: roles }}>
      {children}
    </GuardRoute>
  )
}

export function PermissionGuard({
  children,
  permissions,
  requireAll = true
}: {
  children: ReactNode
  permissions: string[]
  requireAll?: boolean
}) {
  return (
    <GuardRoute
      config={{
        requireAuth: true,
        requiredPermissions: permissions,
        requireAllPermissions: requireAll,
      }}
    >
      {children}
    </GuardRoute>
  )
}
