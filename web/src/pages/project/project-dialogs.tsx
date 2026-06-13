import { Input, Textarea } from '@cloudflare/kumo/components/input'
import { Select } from '@cloudflare/kumo/components/select'
import { useTranslation } from 'react-i18next'
import { AppDialog } from '@/components/app-dialog'
import { DataFormDialog } from '@/components/data-form-dialog'
import type { Project } from '@/services/projects'
import type { ProjectFormState } from './project-form'
import type { Dispatch, SetStateAction } from 'react'

interface ProjectDialogsProps {
  createModalOpen: boolean
  editingProject: Project | null
  deleteProject: Project | null
  formState: ProjectFormState
  setCreateModalOpen: (open: boolean) => void
  setEditingProject: (project: Project | null) => void
  setDeleteProject: (project: Project | null) => void
  setFormState: Dispatch<SetStateAction<ProjectFormState>>
  onCreate: () => void
  onEdit: () => void
  onDelete: () => void
}

export function ProjectDialogs({
  createModalOpen,
  editingProject,
  deleteProject,
  formState,
  setCreateModalOpen,
  setEditingProject,
  setDeleteProject,
  setFormState,
  onCreate,
  onEdit,
  onDelete,
}: ProjectDialogsProps) {
  const { t } = useTranslation()

  const renderProjectForm = (showStatus: boolean) => (
    <>
      <Input
        label={t('projects.form.name')}
        required
        maxLength={255}
        value={formState.name}
        placeholder={t('projects.form.namePlaceholder')}
        onChange={(event) => setFormState(prev => ({ ...prev, name: event.currentTarget.value }))}
      />
      <Textarea
        label={t('projects.form.description')}
        rows={3}
        value={formState.description}
        placeholder={t('projects.form.descriptionPlaceholder')}
        onChange={(event) => setFormState(prev => ({ ...prev, description: event.currentTarget.value }))}
      />
      {showStatus && (
        <Select
          label={t('projects.form.status')}
          value={String(formState.status ?? 1)}
          onValueChange={(value) => setFormState(prev => ({ ...prev, status: Number(value) }))}
          renderValue={(value) => String(value) === '1' ? t('projects.status.enabled') : t('projects.status.disabled')}
        >
          <Select.Option value="1">{t('projects.status.enabled')}</Select.Option>
          <Select.Option value="0">{t('projects.status.disabled')}</Select.Option>
        </Select>
      )}
    </>
  )

  return (
    <>
      <DataFormDialog
        open={createModalOpen}
        title={t('projects.createProject')}
        confirmText={t('projects.createProject')}
        onOpenChange={setCreateModalOpen}
        onSubmit={onCreate}
      >
        {renderProjectForm(false)}
      </DataFormDialog>

      <DataFormDialog
        open={!!editingProject}
        title={t('projects.editProject')}
        confirmText={t('projects.actions.submit')}
        onOpenChange={(open) => !open && setEditingProject(null)}
        onSubmit={onEdit}
      >
        {renderProjectForm(true)}
      </DataFormDialog>

      <AppDialog
        open={!!deleteProject}
        role="alertdialog"
        title={t('projects.confirm.deleteTitle')}
        description={t('projects.confirm.deleteDescription')}
        confirmText={t('projects.actions.delete')}
        confirmVariant="destructive"
        onOpenChange={(open) => !open && setDeleteProject(null)}
        onConfirm={onDelete}
      />
    </>
  )
}
