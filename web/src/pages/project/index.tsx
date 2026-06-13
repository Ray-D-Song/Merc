import { useCallback, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { Badge } from '@cloudflare/kumo/components/badge'
import { Button } from '@cloudflare/kumo/components/button'
import { PlusIcon } from '@phosphor-icons/react'
import { DataTable, type DataTableColumn } from '@/components/data-table'
import { useDataTable } from '@/hooks/use-data-table'
import { projectsService, type Project, type CreateProjectRequest, type UpdateProjectRequest } from '@/services/projects'
import { useAuth } from '@/hooks/use-auth'
import { useMessage } from '@/contexts/feedback-context'
import { formatDateTime } from '@/utils/date'
import { ProjectDialogs } from './project-dialogs'
import { defaultProjectForm, type ProjectFormState } from './project-form'

function toCreateRequest(form: ProjectFormState): CreateProjectRequest {
  return {
    repositoryUrl: form.repositoryUrl,
    name: form.name || undefined,
    description: form.description || undefined,
  }
}

function toUpdateRequest(form: ProjectFormState, id: number): UpdateProjectRequest {
  return {
    id,
    repositoryUrl: form.repositoryUrl || undefined,
    name: form.name || '',
    description: form.description || undefined,
    status: form.status,
  }
}

export default function ProjectPage() {
  const { t } = useTranslation()
  const message = useMessage()
  const { isAdmin } = useAuth()
  const [createModalOpen, setCreateModalOpen] = useState(false)
  const [editingProject, setEditingProject] = useState<Project | null>(null)
  const [deleteProject, setDeleteProject] = useState<Project | null>(null)
  const [formState, setFormState] = useState<ProjectFormState>(defaultProjectForm)

  const projectTable = useDataTable<Project>({
    fetcher: useCallback(async ({ page, pageSize }) => {
      const response = await projectsService.listProjects({
        current: page,
        pageSize,
      })

      return {
        data: response?.data || [],
        total: response?.pagination.total || 0,
        page: response?.pagination.page,
        pageSize: response?.pagination.pageSize,
      }
    }, []),
    getErrorMessage: useCallback(() => t('projects.messages.loadFailed'), [t]),
    onError: useCallback(() => {
      message.error(t('projects.messages.loadFailed'))
    }, [message, t]),
  })

  const openCreateModal = () => {
    setFormState(defaultProjectForm)
    setCreateModalOpen(true)
  }

  const openEditModal = (record: Project) => {
    setEditingProject(record)
    setFormState({
      repositoryUrl: record.repositoryUrl,
      name: record.name,
      description: record.description || '',
      status: record.status,
    })
  }

  const handleCreate = async () => {
    try {
      await projectsService.createProject(toCreateRequest(formState))
      setCreateModalOpen(false)
      setFormState(defaultProjectForm)
      projectTable.reload()
    } catch (err) {
      console.error('Failed to create project:', err)
      message.error(t('projects.messages.createFailed'))
    }
  }

  const handleEdit = async () => {
    if (!editingProject) return
    try {
      await projectsService.updateProject(toUpdateRequest(formState, editingProject.id))
      setEditingProject(null)
      setFormState(defaultProjectForm)
      projectTable.reload()
    } catch (err) {
      console.error('Failed to update project:', err)
      message.error(t('projects.messages.updateFailed'))
    }
  }

  const handleDelete = async () => {
    if (!deleteProject) return
    try {
      await projectsService.deleteProject(deleteProject.id)
      setDeleteProject(null)
      projectTable.reload()
    } catch (err) {
      console.error('Failed to delete project:', err)
      message.error(t('projects.messages.deleteFailed'))
    }
  }

  const columns: DataTableColumn<Project>[] = [
    {
      key: 'id',
      title: t('projects.columns.id'),
      width: '80px',
      render: (record) => record.id,
    },
    {
      key: 'name',
      title: t('projects.columns.name'),
      render: (record) => record.name,
    },
    {
      key: 'repositoryUrl',
      title: 'GitHub 仓库',
      render: (record) => record.repositoryUrl || '-',
    },
    {
      key: 'status',
      title: t('projects.columns.status'),
      render: (record) => (
        <Badge variant={record.status === 1 ? 'success' : 'error'} appearance="dot">
          {record.status === 1 ? t('projects.status.enabled') : t('projects.status.disabled')}
        </Badge>
      ),
    },
    {
      key: 'createrUserName',
      title: t('projects.columns.createrUserName'),
      render: (record) => record.createrUserName || '-',
    },
    {
      key: 'lastUpdaterUserName',
      title: t('projects.columns.lastUpdaterUserName'),
      render: (record) => record.lastUpdaterUserName || '-',
    },
    {
      key: 'createdAt',
      title: t('projects.columns.createdAt'),
      render: (record) => formatDateTime(record.createdAt),
    },
    {
      key: 'updatedAt',
      title: t('projects.columns.updatedAt'),
      render: (record) => formatDateTime(record.updatedAt),
    },
    {
      key: 'actions',
      title: t('projects.columns.actions'),
      render: (record) => (
        <div className="flex flex-wrap gap-2">
          {isAdmin() && (
            <>
              <Button type="button" size="sm" variant="ghost" onClick={() => openEditModal(record)}>
                {t('projects.actions.edit')}
              </Button>
              <Button type="button" size="sm" variant="secondary-destructive" onClick={() => setDeleteProject(record)}>
                {t('projects.actions.delete')}
              </Button>
            </>
          )}
        </div>
      ),
    },
  ]

  return (
    <>
      <DataTable
        title={t('projects.title')}
        columns={columns}
        table={projectTable}
        rowKey={(record) => record.id}
        toolbar={isAdmin() && (
          <Button type="button" size="sm" variant="primary" icon={<PlusIcon size={14} />} onClick={openCreateModal}>
            {t('projects.createProject')}
          </Button>
        )}
      />

      <ProjectDialogs
        createModalOpen={createModalOpen}
        editingProject={editingProject}
        deleteProject={deleteProject}
        formState={formState}
        setCreateModalOpen={setCreateModalOpen}
        setEditingProject={setEditingProject}
        setDeleteProject={setDeleteProject}
        setFormState={setFormState}
        onCreate={handleCreate}
        onEdit={handleEdit}
        onDelete={handleDelete}
      />
    </>
  )
}
