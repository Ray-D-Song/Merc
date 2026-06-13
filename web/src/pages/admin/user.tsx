import { useCallback, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { Badge } from '@cloudflare/kumo/components/badge'
import { Button } from '@cloudflare/kumo/components/button'
import { Input } from '@cloudflare/kumo/components/input'
import { Select } from '@cloudflare/kumo/components/select'
import { SensitiveInput } from '@cloudflare/kumo/components/sensitive-input'
import { PlusIcon } from '@phosphor-icons/react'
import { AppDialog } from '@/components/app-dialog'
import { DataFormDialog } from '@/components/data-form-dialog'
import { DataTable, type DataTableColumn } from '@/components/data-table'
import { useDataTable } from '@/hooks/use-data-table'
import { usersService, type User, type CreateUserRequest, type UsersListParams } from '@/services/users'
import { useAuth } from '@/hooks/use-auth'
import { useMessage } from '@/contexts/feedback-context'
import { formatDateTime } from '@/utils/date'

const defaultCreateForm: CreateUserRequest = {
  username: '',
  realname: '',
  email: '',
  password: '',
  role: 'user',
}

export default function UserPage() {
  const { t } = useTranslation()
  const message = useMessage()
  const { user: currentAuthUser } = useAuth()
  const [filters, setFilters] = useState<UsersListParams>({})
  const [createModalOpen, setCreateModalOpen] = useState(false)
  const [createForm, setCreateForm] = useState<CreateUserRequest>(defaultCreateForm)
  const [confirmUser, setConfirmUser] = useState<User | null>(null)

  const userTable = useDataTable<User>({
    fetcher: useCallback(async ({ page, pageSize }) => {
      const response = await usersService.getUsers({
        current: page,
        pageSize,
        username: filters.username,
        realname: filters.realname,
        email: filters.email,
        role: filters.role,
        isEnabled: filters.isEnabled,
      })

      return {
        data: response?.data || [],
        total: response?.pagination?.total || 0,
        page: response?.pagination?.page,
        pageSize: response?.pagination?.pageSize,
      }
    }, [filters]),
    getErrorMessage: useCallback(() => t('users.messages.loadFailed'), [t]),
    onError: useCallback(() => {
      message.error(t('users.messages.loadFailed'))
    }, [message, t]),
  })

  const handleCreate = async () => {
    try {
      await usersService.createUser(createForm)
      message.success(t('users.messages.createSuccess'))
      setCreateModalOpen(false)
      setCreateForm(defaultCreateForm)
      userTable.reload()
    } catch (err) {
      console.error('Failed to create user:', err)
      message.error(t('users.messages.createFailed'))
    }
  }

  const handleToggleStatus = async () => {
    if (!confirmUser) return

    try {
      if (confirmUser.isEnabled) {
        await usersService.disableUser(confirmUser.id)
        message.success(t('users.messages.disableSuccess'))
      } else {
        await usersService.enableUser(confirmUser.id)
        message.success(t('users.messages.enableSuccess'))
      }
      setConfirmUser(null)
      userTable.reload()
    } catch (err) {
      console.error('Failed to toggle user status:', err)
      message.error(confirmUser.isEnabled ? t('users.messages.disableFailed') : t('users.messages.enableFailed'))
    }
  }

  const columns: DataTableColumn<User>[] = [
    {
      key: 'id',
      title: t('users.columns.id'),
      width: '80px',
      render: (record) => record.id,
    },
    {
      key: 'username',
      title: t('users.columns.username'),
      render: (record) => record.username,
    },
    {
      key: 'realname',
      title: t('users.columns.realname'),
      render: (record) => record.realname,
    },
    {
      key: 'email',
      title: t('users.columns.email'),
      render: (record) => record.email,
    },
    {
      key: 'role',
      title: t('users.columns.role'),
      render: (record) => (
        <Badge variant={record.role === 'admin' ? 'warning' : 'info'}>
          {record.role === 'admin' ? t('users.roles.admin') : t('users.roles.user')}
        </Badge>
      ),
    },
    {
      key: 'status',
      title: t('users.columns.status'),
      render: (record) => (
        <Badge variant={record.isEnabled ? 'success' : 'error'} appearance="dot">
          {record.isEnabled ? t('users.status.enabled') : t('users.status.disabled')}
        </Badge>
      ),
    },
    {
      key: 'createdAt',
      title: t('users.columns.createdAt'),
      render: (record) => formatDateTime(record.createdAt),
    },
    {
      key: 'actions',
      title: t('users.columns.actions'),
      render: (record) => {
        const canToggleStatus = record.id !== currentAuthUser?.id

        return canToggleStatus ? (
          <Button
            type="button"
            size="sm"
            variant={record.isEnabled ? 'secondary-destructive' : 'secondary'}
            onClick={() => setConfirmUser(record)}
          >
            {record.isEnabled ? t('users.actions.disable') : t('users.actions.enable')}
          </Button>
        ) : null
      },
    },
  ]

  return (
    <>
      <DataTable
        title={t('users.title')}
        columns={columns}
        table={userTable}
        rowKey={(record) => record.id}
        toolbar={(
          <Button type="button" size="sm" variant="primary" icon={<PlusIcon size={14} />} onClick={() => setCreateModalOpen(true)}>
            {t('users.createUser')}
          </Button>
        )}
        filters={(
          <>
            <div className="min-w-44 flex-1">
              <Input
                size="sm"
                label={t('users.columns.username')}
                value={filters.username || ''}
                onChange={(event) => {
                  userTable.resetPage()
                  setFilters(prev => ({ ...prev, username: event.currentTarget.value || undefined }))
                }}
              />
            </div>
            <div className="min-w-44 flex-1">
              <Input
                size="sm"
                label={t('users.columns.realname')}
                value={filters.realname || ''}
                onChange={(event) => {
                  userTable.resetPage()
                  setFilters(prev => ({ ...prev, realname: event.currentTarget.value || undefined }))
                }}
              />
            </div>
            <div className="min-w-52 flex-1">
              <Input
                size="sm"
                label={t('users.columns.email')}
                value={filters.email || ''}
                onChange={(event) => {
                  userTable.resetPage()
                  setFilters(prev => ({ ...prev, email: event.currentTarget.value || undefined }))
                }}
              />
            </div>
            <div className="min-w-28">
              <Select
                size="sm"
                label={t('users.columns.role')}
                value={filters.role || 'all'}
                onValueChange={(value) => {
                  userTable.resetPage()
                  setFilters(prev => ({ ...prev, role: value === 'all' ? undefined : value as 'admin' | 'user' }))
                }}
              >
                <Select.Option value="all">All</Select.Option>
                <Select.Option value="admin">{t('users.roles.admin')}</Select.Option>
                <Select.Option value="user">{t('users.roles.user')}</Select.Option>
              </Select>
            </div>
          </>
        )}
      />

      <DataFormDialog
        open={createModalOpen}
        title={t('users.createUser')}
        confirmText={t('users.createUser')}
        onOpenChange={setCreateModalOpen}
        onSubmit={handleCreate}
      >
        <Input
          label={t('users.form.username')}
          required
          minLength={3}
          pattern="[a-zA-Z0-9_]+"
          value={createForm.username}
          placeholder={t('users.form.usernamePlaceholder')}
          onChange={(event) => setCreateForm(prev => ({ ...prev, username: event.currentTarget.value }))}
        />
        <Input
          label={t('users.form.realname')}
          required
          value={createForm.realname}
          placeholder={t('users.form.realnamePlaceholder')}
          onChange={(event) => setCreateForm(prev => ({ ...prev, realname: event.currentTarget.value }))}
        />
        <Input
          label={t('users.form.email')}
          required
          type="email"
          value={createForm.email}
          placeholder={t('users.form.emailPlaceholder')}
          onChange={(event) => setCreateForm(prev => ({ ...prev, email: event.currentTarget.value }))}
        />
        <SensitiveInput
          label={t('users.form.password')}
          required
          minLength={6}
          value={createForm.password}
          placeholder={t('users.form.passwordPlaceholder')}
          onValueChange={(value) => setCreateForm(prev => ({ ...prev, password: value }))}
        />
        <Select
          label={t('users.form.role')}
          required
          value={createForm.role}
          placeholder={t('users.form.rolePlaceholder')}
          onValueChange={(value) => setCreateForm(prev => ({ ...prev, role: value as 'admin' | 'user' }))}
        >
          <Select.Option value="user">{t('users.roles.user')}</Select.Option>
          <Select.Option value="admin">{t('users.roles.admin')}</Select.Option>
        </Select>
      </DataFormDialog>

      <AppDialog
        open={!!confirmUser}
        role="alertdialog"
        title={confirmUser?.isEnabled ? t('users.confirm.disableTitle') : t('users.confirm.enableTitle')}
        description={confirmUser?.isEnabled ? t('users.confirm.disableDescription') : t('users.confirm.enableDescription')}
        confirmText={confirmUser?.isEnabled ? t('users.actions.disable') : t('users.actions.enable')}
        confirmVariant={confirmUser?.isEnabled ? 'destructive' : 'primary'}
        onOpenChange={(open) => !open && setConfirmUser(null)}
        onConfirm={handleToggleStatus}
      />
    </>
  )
}
