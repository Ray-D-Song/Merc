import { http } from '@/utils/fetcher'
import type { Optional } from '@/utils/types'

export interface Project {
  id: number
  name: string
  provider: string
  repositoryUrl: string
  owner: string
  repo: string
  description: string
  status: number
  createrId: number
  createrUserName: string
  lastUpdaterId: number
  lastUpdaterUserName: string
  createdAt: string
  updatedAt: string
}

export interface CreateProjectRequest {
  repositoryUrl: string
  name?: string
  description?: string
}

export interface UpdateProjectRequest {
  id: number
  repositoryUrl?: string
  name: string
  description?: string
  status?: number
}

interface PaginationMeta {
  page: number
  pageSize: number
  total: number
  totalPages: number
}

export interface ProjectListResponse {
  data: Project[]
  pagination: PaginationMeta
}

export interface ProjectListParams {
  current?: number
  pageSize?: number
}

export const projectsService = {
  listProjects: (params?: ProjectListParams): Promise<Optional<ProjectListResponse>> =>
    http.post('/v1/project/list', {
      page: params?.current ?? 1,
      page_size: params?.pageSize ?? 20
    }),

  getProject: (id: number): Promise<Optional<Project>> =>
    http.get(`/v1/project/${id}`),

  createProject: (data: CreateProjectRequest): Promise<Optional<Project>> =>
    http.post('/v1/project/create', data, {}, { showSuccessMessage: true }),

  updateProject: (data: UpdateProjectRequest): Promise<Optional<Project>> =>
    http.put('/v1/project/update', data, {}, { showSuccessMessage: true }),

  deleteProject: (id: number): Promise<Optional<{ message: string }>> =>
    http.delete(`/v1/project/${id}`, {}, { showSuccessMessage: true }),
}
