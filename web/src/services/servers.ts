import { http } from '@/utils/fetcher'
import type { Optional } from '@/utils/types'

export interface ServerNode {
  id: number
  nodeKey: string
  name: string
  hostname: string
  roles: string
  os: string
  arch: string
  version: string
  status: string
  resourceJson: string
  lastHeartbeat?: string
  createdAt: string
  updatedAt: string
}

interface PaginationMeta {
  page: number
  pageSize: number
  total: number
  totalPages: number
}

export interface ServerNodeListResponse {
  data: ServerNode[]
  pagination: PaginationMeta
}

export interface ServerNodeListParams {
  current?: number
  pageSize?: number
  status?: string
  os?: string
}

export interface CreateAgentTokenResponse {
  id: number
  name: string
  token: string
  expiresAt?: string
  createdAt: string
}

export const serversService = {
  listServers: (params?: ServerNodeListParams): Promise<Optional<ServerNodeListResponse>> =>
    http.post('/v1/server/list', {
      page: params?.current ?? 1,
      page_size: params?.pageSize ?? 20,
      status: params?.status || undefined,
      os: params?.os || undefined,
    }),

  createAgentToken: (name: string): Promise<Optional<CreateAgentTokenResponse>> =>
    http.post('/v1/server/token/create', { name }, {}, { showSuccessMessage: true }),
}
