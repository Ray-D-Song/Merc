import { http } from '../utils/fetcher'

export interface LoginCredentials {
  username: string
  password: string
}

export interface RegisterCredentials {
  username: string
  password: string
}

export interface AuthResponse {
  id: number
  username: string
  realname: string
  email: string
  role: string
  isAdmin: boolean
}

// For non-hook usage (direct service calls)
export const authService = {
  login: (credentials: LoginCredentials) =>
    http.post<AuthResponse>('/v1/auth/login', credentials, {}, { hideErrorMessage: true }),

  register: (credentials: RegisterCredentials) =>
    http.post<AuthResponse>('/v1/auth/register', credentials, {}, { hideErrorMessage: true }),

  logout: () =>
    http.post<void>('/v1/auth/logout', undefined, {}, { showSuccessMessage: true }),

  getCurrentUser: () =>
    http.get<AuthResponse>('/v1/auth/me', {}, { hideErrorMessage: true }),
}
