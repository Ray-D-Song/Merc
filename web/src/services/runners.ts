import { http } from '@/utils/fetcher'
import type { Optional } from '@/utils/types'

export interface Runner {
  id: number
  projectId: number
  nodeId: number
  name: string
  labels: string
  workDir: string
  installDir: string
  os: string
  arch: string
  status: string
  processId: number
  lastError: string
  lastSeenAt?: string
  createdAt: string
  updatedAt: string
}

export interface CreateRunnerRequest {
  projectId: number
  nodeId: number
  name: string
  labels?: string
  workDir?: string
  registrationToken: string
}

interface PaginationMeta {
  page: number
  pageSize: number
  total: number
  totalPages: number
}

export interface RunnerListResponse {
  data: Runner[]
  pagination: PaginationMeta
}

export interface RunnerListParams {
  current?: number
  pageSize?: number
  projectId?: number
  nodeId?: number
  status?: string
}

export const runnersService = {
  listRunners: (params?: RunnerListParams): Promise<Optional<RunnerListResponse>> =>
    http.post('/v1/runner/list', {
      page: params?.current ?? 1,
      page_size: params?.pageSize ?? 20,
      projectId: params?.projectId || undefined,
      nodeId: params?.nodeId || undefined,
      status: params?.status || undefined,
    }),

  createRunner: (data: CreateRunnerRequest): Promise<Optional<Runner>> =>
    http.post('/v1/runner/create', data, {}, { showSuccessMessage: true }),

  startRunner: (id: number) =>
    http.post(`/v1/runner/${id}/start`, {}, {}, { showSuccessMessage: true }),

  stopRunner: (id: number) =>
    http.post(`/v1/runner/${id}/stop`, {}, {}, { showSuccessMessage: true }),

  removeRunner: (id: number) =>
    http.post(`/v1/runner/${id}/remove`, {}, {}, { showSuccessMessage: true }),
}
