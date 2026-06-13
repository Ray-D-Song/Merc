import { http } from '@/utils/fetcher'
import type { Optional } from '@/utils/types'

export interface User {
  id: number
  username: string
  realname: string
  email: string
  role: 'admin' | 'user'
  isAdmin: boolean
  isEnabled: boolean
  createdAt: string
  updatedAt: string
}

export interface CreateUserRequest {
  username: string
  realname: string
  email: string
  password: string
  role: 'admin' | 'user'
}

export interface UpdateUserRequest {
  realname?: string
  email?: string
  password?: string
  role?: 'admin' | 'user'
}

export interface UsersListResponse {
  data: User[]
  pagination: {
    page: number
    pageSize: number
    total: number
    totalPages: number
  }
}

export interface UsersListParams {
  current?: number
  pageSize?: number
  role?: 'admin' | 'user'
  isEnabled?: boolean
  username?: string
  realname?: string
  email?: string
}

export const usersService = {
  // Get paginated user list
  getUsers: (params?: UsersListParams): Promise<Optional<UsersListResponse>> => {
    const searchParams = new URLSearchParams()
    if (params?.current) searchParams.append('page', params.current.toString())
    if (params?.pageSize) searchParams.append('page_size', params.pageSize.toString())
    if (params?.username) searchParams.append('username', params.username)
    if (params?.realname) searchParams.append('realname', params.realname)
    if (params?.email) searchParams.append('email', params.email)
    if (params?.role) searchParams.append('role', params.role)
    if (params?.isEnabled !== undefined) searchParams.append('isEnabled', params.isEnabled.toString())

    const queryString = searchParams.toString()
    return http.get(`/v1/user/list${queryString ? `?${queryString}` : ''}`)
  },

  // Create a new user
  createUser: (data: CreateUserRequest): Promise<Optional<User>> =>
    http.post('/v1/user/create', data, {}, { showSuccessMessage: true }),

  // Update user information
  updateUser: (id: number, data: UpdateUserRequest): Promise<Optional<User>> =>
    http.patch(`/v1/user/update/${id}`, data, {}, { showSuccessMessage: true }),

  // Disable a user
  disableUser: (userId: number): Promise<Optional<{ message: string }>> =>
    http.post('/v1/user/disable', { userId }, {}, { showSuccessMessage: true }),

  // Enable a user
  enableUser: (userId: number): Promise<Optional<{ message: string }>> =>
    http.post('/v1/user/enable', { userId }, {}, { showSuccessMessage: true }),
}
