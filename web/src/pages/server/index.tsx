import { useCallback, useEffect, useState } from 'react'
import { Badge } from '@cloudflare/kumo/components/badge'
import { Button } from '@cloudflare/kumo/components/button'
import { CopyIcon, PlusIcon } from '@phosphor-icons/react'
import { DataTable, type DataTableColumn } from '@/components/data-table'
import { useDataTable } from '@/hooks/use-data-table'
import { serversService, type ServerNode } from '@/services/servers'
import { useMessage } from '@/contexts/feedback-context'
import { formatDateTime } from '@/utils/date'

export default function ServerPage() {
  const message = useMessage()
  const [joinCommand, setJoinCommand] = useState('')

  const serverTable = useDataTable<ServerNode>({
    fetcher: useCallback(async ({ page, pageSize }) => {
      const response = await serversService.listServers({ current: page, pageSize })
      return {
        data: response?.data || [],
        total: response?.pagination.total || 0,
        page: response?.pagination.page,
        pageSize: response?.pagination.pageSize,
      }
    }, []),
    getErrorMessage: useCallback(() => '获取服务器列表失败', []),
    onError: useCallback(() => message.error('获取服务器列表失败'), [message]),
  })
  const { reload: reloadServers } = serverTable

  useEffect(() => {
    const timer = window.setInterval(() => {
      reloadServers()
    }, 5000)

    return () => window.clearInterval(timer)
  }, [reloadServers])

  const createJoinCommand = async () => {
    const response = await serversService.createAgentToken(`agent-${Date.now()}`)
    if (!response?.token) return
    const origin = window.location.origin
    setJoinCommand(`merc agent --server-url ${origin} --token ${response.token}`)
  }

  const copyJoinCommand = async () => {
    if (!joinCommand) return
    await navigator.clipboard.writeText(joinCommand)
    message.success('已复制加入命令')
  }

  const columns: DataTableColumn<ServerNode>[] = [
    {
      key: 'name',
      title: '名称',
      render: record => record.name || record.hostname,
    },
    {
      key: 'roles',
      title: '角色',
      render: record => record.roles,
    },
    {
      key: 'os',
      title: '系统',
      render: record => `${record.os}/${record.arch}`,
    },
    {
      key: 'status',
      title: '状态',
      render: record => (
        <Badge variant={record.status === 'online' ? 'success' : 'error'} appearance="dot">
          {record.status}
        </Badge>
      ),
    },
    {
      key: 'lastHeartbeat',
      title: '最后心跳',
      render: record => record.lastHeartbeat ? formatDateTime(record.lastHeartbeat) : '-',
    },
    {
      key: 'version',
      title: '版本',
      render: record => record.version || '-',
    },
  ]

  return (
    <div className="space-y-3">
      <DataTable
        title="服务器"
        columns={columns}
        table={serverTable}
        rowKey={record => record.id}
        toolbar={(
          <Button type="button" size="sm" variant="primary" icon={<PlusIcon size={14} />} onClick={createJoinCommand}>
            生成加入命令
          </Button>
        )}
      />
      {joinCommand && (
        <div className="flex items-center gap-2 rounded border border-kumo-line bg-kumo-base p-3">
          <code className="min-w-0 flex-1 overflow-x-auto text-sm text-kumo-default">{joinCommand}</code>
          <Button type="button" size="sm" shape="square" icon={<CopyIcon size={16} />} aria-label="复制" onClick={copyJoinCommand} />
        </div>
      )}
    </div>
  )
}
