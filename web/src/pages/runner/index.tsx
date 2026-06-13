import { useCallback, useEffect, useState } from 'react'
import { Badge } from '@cloudflare/kumo/components/badge'
import { Button } from '@cloudflare/kumo/components/button'
import { Input } from '@cloudflare/kumo/components/input'
import { Select } from '@cloudflare/kumo/components/select'
import { PlusIcon } from '@phosphor-icons/react'
import { DataFormDialog } from '@/components/data-form-dialog'
import { DataTable, type DataTableColumn } from '@/components/data-table'
import { useDataTable } from '@/hooks/use-data-table'
import { projectsService, type Project } from '@/services/projects'
import { runnersService, type Runner } from '@/services/runners'
import { serversService, type ServerNode } from '@/services/servers'
import { useMessage } from '@/contexts/feedback-context'
import { formatDateTime } from '@/utils/date'

interface RunnerFormState {
  projectId: string
  nodeId: string
  name: string
  labels: string
  workDir: string
  registrationToken: string
}

const defaultForm: RunnerFormState = {
  projectId: '',
  nodeId: '',
  name: '',
  labels: 'self-hosted',
  workDir: '',
  registrationToken: '',
}

export default function RunnerPage() {
  const message = useMessage()
  const [createOpen, setCreateOpen] = useState(false)
  const [form, setForm] = useState<RunnerFormState>(defaultForm)
  const [projects, setProjects] = useState<Project[]>([])
  const [nodes, setNodes] = useState<ServerNode[]>([])

  const runnerTable = useDataTable<Runner>({
    fetcher: useCallback(async ({ page, pageSize }) => {
      const response = await runnersService.listRunners({ current: page, pageSize })
      return {
        data: response?.data || [],
        total: response?.pagination.total || 0,
        page: response?.pagination.page,
        pageSize: response?.pagination.pageSize,
      }
    }, []),
    getErrorMessage: useCallback(() => '获取 Runner 列表失败', []),
    onError: useCallback(() => message.error('获取 Runner 列表失败'), [message]),
  })

  useEffect(() => {
    void projectsService.listProjects({ current: 1, pageSize: 100 }).then(response => setProjects(response?.data || []))
    void serversService.listServers({ current: 1, pageSize: 100 }).then(response => setNodes(response?.data || []))
  }, [])

  const handleCreate = async () => {
    await runnersService.createRunner({
      projectId: Number(form.projectId),
      nodeId: Number(form.nodeId),
      name: form.name,
      labels: form.labels || undefined,
      workDir: form.workDir || undefined,
      registrationToken: form.registrationToken,
    })
    setCreateOpen(false)
    setForm(defaultForm)
    runnerTable.reload()
  }

  const enqueue = async (runner: Runner, action: 'start' | 'stop' | 'remove') => {
    if (action === 'start') await runnersService.startRunner(runner.id)
    if (action === 'stop') await runnersService.stopRunner(runner.id)
    if (action === 'remove') await runnersService.removeRunner(runner.id)
    runnerTable.reload()
  }

  const projectName = (id: number) => projects.find(project => project.id === id)?.name || id
  const nodeName = (id: number) => nodes.find(node => node.id === id)?.name || id

  const columns: DataTableColumn<Runner>[] = [
    {
      key: 'name',
      title: '名称',
      render: record => record.name,
    },
    {
      key: 'projectId',
      title: '项目',
      render: record => projectName(record.projectId),
    },
    {
      key: 'nodeId',
      title: '服务器',
      render: record => nodeName(record.nodeId),
    },
    {
      key: 'labels',
      title: '标签',
      render: record => record.labels || '-',
    },
    {
      key: 'status',
      title: '状态',
      render: record => (
        <Badge variant={record.status === 'running' ? 'success' : record.status === 'error' ? 'error' : 'neutral'} appearance="dot">
          {record.status}
        </Badge>
      ),
    },
    {
      key: 'lastError',
      title: '最近错误',
      render: record => record.lastError || '-',
    },
    {
      key: 'createdAt',
      title: '创建时间',
      render: record => formatDateTime(record.createdAt),
    },
    {
      key: 'actions',
      title: '操作',
      render: record => (
        <div className="flex flex-wrap gap-2">
          <Button type="button" size="sm" variant="ghost" onClick={() => enqueue(record, 'start')}>启动</Button>
          <Button type="button" size="sm" variant="ghost" onClick={() => enqueue(record, 'stop')}>停止</Button>
          <Button type="button" size="sm" variant="secondary-destructive" onClick={() => enqueue(record, 'remove')}>移除</Button>
        </div>
      ),
    },
  ]

  return (
    <>
      <DataTable
        title="Runner"
        columns={columns}
        table={runnerTable}
        rowKey={record => record.id}
        toolbar={(
          <Button type="button" size="sm" variant="primary" icon={<PlusIcon size={14} />} onClick={() => setCreateOpen(true)}>
            新建 Runner
          </Button>
        )}
      />

      <DataFormDialog
        open={createOpen}
        title="新建 Runner"
        confirmText="提交"
        onOpenChange={setCreateOpen}
        onSubmit={handleCreate}
      >
        <Select
          label="项目"
          value={form.projectId}
          onValueChange={value => setForm(prev => ({ ...prev, projectId: value || '' }))}
          renderValue={value => projectName(Number(value))}
        >
          {projects.map(project => (
            <Select.Option key={project.id} value={String(project.id)}>{project.name}</Select.Option>
          ))}
        </Select>
        <Select
          label="服务器"
          value={form.nodeId}
          onValueChange={value => setForm(prev => ({ ...prev, nodeId: value || '' }))}
          renderValue={value => nodeName(Number(value))}
        >
          {nodes.map(node => (
            <Select.Option key={node.id} value={String(node.id)}>{node.name || node.hostname}</Select.Option>
          ))}
        </Select>
        <Input
          label="Runner 名称"
          required
          value={form.name}
          onChange={event => setForm(prev => ({ ...prev, name: event.currentTarget.value }))}
        />
        <Input
          label="标签"
          value={form.labels}
          onChange={event => setForm(prev => ({ ...prev, labels: event.currentTarget.value }))}
        />
        <Input
          label="工作目录"
          value={form.workDir}
          placeholder="留空使用 _work"
          onChange={event => setForm(prev => ({ ...prev, workDir: event.currentTarget.value }))}
        />
        <Input
          label="GitHub Registration Token"
          required
          value={form.registrationToken}
          onChange={event => setForm(prev => ({ ...prev, registrationToken: event.currentTarget.value }))}
        />
      </DataFormDialog>
    </>
  )
}
