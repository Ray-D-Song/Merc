import { useState, useEffect, useCallback, useRef } from 'react'
import { authService } from '@/services/auth'
import { Errors } from '@/utils/error'

export interface User {
  id: number
  username: string
  realname: string
  email: string
  role: string
  isAdmin: boolean
}

interface AuthState {
  user: User | null
  loading: boolean
  error: string | null
}

export const useAuth = () => {
  const [authState, setAuthState] = useState<AuthState>({
    user: null,
    loading: true,
    error: null
  })

  // Track if we're currently fetching user to prevent duplicate calls
  const isFetchingUser = useRef(false)

  const getCurrentUser = useCallback(async () => {
    // Prevent duplicate calls
    if (isFetchingUser.current) {
      return
    }

    try {
      isFetchingUser.current = true
      setAuthState(prev => ({ ...prev, loading: true, error: null }))
      const res = await authService.getCurrentUser()
      if (res) {
        setAuthState({ user: res as User, loading: false, error: null })
      }
    } catch (error) {
      console.error('Failed to get current user:', error)
      if (error instanceof Errors.Unauthorized) {
        setAuthState({
          user: null,
          loading: false,
          error: null
        })
        return
      }
      setAuthState({
        user: null,
        loading: false,
        error: error instanceof Error ? error.message : 'Authentication failed'
      })
    } finally {
      isFetchingUser.current = false
    }
  }, [])

  // Initialize auth state on mount
  useEffect(() => {
    getCurrentUser()
  }, [getCurrentUser])

  const login = useCallback(async (credentials: { username: string; password: string }) => {
    try {
      setAuthState(prev => ({ ...prev, loading: true, error: null }))
      const response = await authService.login(credentials)

      if (!response) return response

      // Backend sets cookie, just update user state
      setAuthState({ user: response as User, loading: false, error: null })

      return response
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : 'Login failed'
      setAuthState({
        user: null,
        loading: false,
        error: errorMessage
      })
      throw error
    }
  }, [])

  const register = useCallback(async (credentials: { username: string; password: string }) => {
    try {
      setAuthState(prev => ({ ...prev, loading: true, error: null }))
      const response = await authService.register(credentials)

      if (!response) return response

      // Backend sets cookie, just update user state
      setAuthState({ user: response as User, loading: false, error: null })

      return response
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : 'Registration failed'
      setAuthState({
        user: null,
        loading: false,
        error: errorMessage
      })
      throw error
    }
  }, [])

  const logout = useCallback(async () => {
    try {
      // Call logout API (which clears the cookie)
      await authService.logout()
    } catch (error) {
      console.error('Logout API call failed:', error)
    } finally {
      // Always clear local state regardless of API call result
      setAuthState({ user: null, loading: false, error: null })
    }
  }, [])

  const clearError = useCallback(() => {
    setAuthState(prev => ({ ...prev, error: null }))
  }, [])

  // Helper functions for authorization
  const hasRole = useCallback((role: string): boolean => {
    return authState.user?.role === role
  }, [authState.user])

  const hasAnyRole = useCallback((roles: string[]): boolean => {
    return authState.user?.role ? roles.includes(authState.user.role) : false
  }, [authState.user])

  // eslint-disable-next-line @typescript-eslint/no-unused-vars
  const hasPermission = useCallback((_permission: string): boolean => {
    // For now, permissions are based on role
    // Admin has all permissions, members have limited permissions
    if (authState.user?.isAdmin) return true
    // Add specific permission logic here based on requirements
    return false
  }, [authState.user])

  const hasAllPermissions = useCallback((permissions: string[]): boolean => {
    return permissions.every(permission => hasPermission(permission))
  }, [hasPermission])

  const hasAnyPermission = useCallback((permissions: string[]): boolean => {
    return permissions.some(permission => hasPermission(permission))
  }, [hasPermission])

  const isAdmin = useCallback(() => {
    return authState.user?.isAdmin ?? false
  }, [authState.user])

  return {
    // State
    user: authState.user,
    loading: authState.loading,
    error: authState.error,
    isAuthenticated: !!authState.user,

    // Actions
    login,
    register,
    logout,
    clearError,
    getCurrentUser,

    // Authorization helpers
    hasRole,
    hasAnyRole,
    hasPermission,
    hasAllPermissions,
    hasAnyPermission,

    isAdmin
  }
}
