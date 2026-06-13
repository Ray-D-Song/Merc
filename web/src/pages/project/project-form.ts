import type { CreateProjectRequest } from '@/services/projects'

export type ProjectFormState = CreateProjectRequest & {
  status?: number
}

export const defaultProjectForm: ProjectFormState = {
  repositoryUrl: '',
  name: '',
  description: '',
}
