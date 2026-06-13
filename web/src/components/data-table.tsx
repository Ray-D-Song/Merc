import { Button } from '@cloudflare/kumo/components/button'
import { Empty } from '@cloudflare/kumo/components/empty'
import { LayerCard } from '@cloudflare/kumo/components/layer-card'
import { Pagination } from '@cloudflare/kumo/components/pagination'
import { Table } from '@cloudflare/kumo/components/table'
import type { ReactNode } from 'react'
import { useTranslation } from 'react-i18next'
import type { DataTableController } from '@/hooks/use-data-table'

export interface DataTableColumn<T> {
  key: string
  title: ReactNode
  width?: string
  className?: string
  render: (record: T) => ReactNode
}

interface DataTableProps<T> {
  title: ReactNode
  columns: DataTableColumn<T>[]
  table: DataTableController<T>
  rowKey: (record: T) => string | number
  toolbar?: ReactNode
  filters?: ReactNode
  emptyTitle?: string
  emptyDescription?: string
}

export function DataTable<T>({
  title,
  columns,
  table,
  rowKey,
  toolbar,
  filters,
  emptyTitle = 'No data',
  emptyDescription = 'There are no rows to show.',
}: DataTableProps<T>) {
  const { t } = useTranslation()
  const {
    data,
    total,
    page,
    pageSize,
    pageSizeOptions,
    loading,
    error,
    setPage,
    setPageSize,
    reload,
  } = table
  const showTable = !error && (loading || data.length > 0)
  const showRefreshOverlay = loading && data.length > 0

  return (
    <section>
      <LayerCard className="overflow-hidden">
        <div className="flex flex-col gap-2 border-b border-kumo-line bg-kumo-base px-3 py-3 sm:flex-row sm:items-center sm:justify-between">
          <div className="min-w-0">
            <h1 className="text-lg font-semibold text-kumo-default">{title}</h1>
          </div>
          {toolbar && <div className="flex shrink-0 flex-wrap gap-2">{toolbar}</div>}
        </div>

        {filters && (
          <div className="flex flex-wrap items-end gap-2 border-b border-kumo-line bg-kumo-tint/40 px-3 py-2">
            {filters}
          </div>
        )}

        {!loading && error && (
          <div className="flex min-h-52 flex-col items-center justify-center gap-3 p-6 text-center">
            <p className="text-kumo-danger">{error}</p>
            <Button type="button" onClick={reload}>
              Retry
            </Button>
          </div>
        )}

        {!loading && !error && data.length === 0 && (
          <div className="min-h-52 p-6">
            <Empty title={emptyTitle} description={emptyDescription} />
          </div>
        )}

        {showTable && (
          <div className="relative overflow-x-auto">
            {showRefreshOverlay && (
              <div className="pointer-events-none absolute inset-x-0 top-0 z-10 h-0.5 overflow-hidden bg-kumo-tint">
                <div className="h-full w-1/3 animate-pulse bg-kumo-brand" />
              </div>
            )}
            <Table className="min-w-full">
              <Table.Header variant="compact">
                <Table.Row>
                  {columns.map(column => (
                    <Table.Head
                      key={column.key}
                      className={`px-3 py-2 text-sm ${column.className ?? ''}`}
                      style={column.width ? { width: column.width } : undefined}
                    >
                      {column.title}
                    </Table.Head>
                  ))}
                </Table.Row>
              </Table.Header>
              <Table.Body>
                {data.length > 0
                  ? data.map(record => (
                      <Table.Row key={rowKey(record)} className={showRefreshOverlay ? 'opacity-70' : undefined}>
                        {columns.map(column => (
                          <Table.Cell key={column.key} className={`px-3 py-2 text-sm ${column.className ?? ''}`}>
                            {column.render(record)}
                          </Table.Cell>
                        ))}
                      </Table.Row>
                    ))
                  : Array.from({ length: Math.min(5, pageSize) }).map((_, rowIndex) => (
                      <Table.Row key={`loading-${rowIndex}`}>
                        {columns.map(column => (
                          <Table.Cell key={`${column.key}-${rowIndex}`} className={`px-3 py-2 ${column.className ?? ''}`}>
                            <div className="h-3 w-20 animate-pulse rounded bg-kumo-tint" />
                          </Table.Cell>
                        ))}
                      </Table.Row>
                    ))}
              </Table.Body>
            </Table>
          </div>
        )}

        <div className="border-t border-kumo-line bg-kumo-base px-3 py-2">
          <Pagination
            page={page}
            setPage={setPage}
            perPage={pageSize}
            totalCount={total}
            labels={{
              navigation: t('pagination.navigation'),
              firstPage: t('pagination.firstPage'),
              previousPage: t('pagination.previousPage'),
              nextPage: t('pagination.nextPage'),
              lastPage: t('pagination.lastPage'),
              pageNumber: t('pagination.pageNumber'),
              pageSize: t('pagination.pageSize'),
            }}
          >
            {total === 0 ? (
              <span className="text-sm text-kumo-subtle">{emptyTitle}</span>
            ) : (
              <>
                <Pagination.Info>
                  {({ pageShowingRange, totalCount }) => (
                    t('pagination.info', { range: pageShowingRange, total: totalCount ?? 0 })
                  )}
                </Pagination.Info>
                <Pagination.Separator />
              </>
            )}
            <Pagination.PageSize
              value={pageSize}
              onChange={setPageSize}
              options={pageSizeOptions}
              label={t('pagination.perPage')}
            />
            <Pagination.Controls />
          </Pagination>
        </div>
      </LayerCard>
    </section>
  )
}
