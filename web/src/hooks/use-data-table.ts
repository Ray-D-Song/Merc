import { useCallback, useEffect, useState, type ReactNode } from 'react'
import { Errors } from '@/utils/error'

export interface DataTableQuery {
  page: number
  pageSize: number
}

export interface DataTableFetchResult<T> {
  data: T[]
  total: number
  page?: number
  pageSize?: number
}

interface UseDataTableOptions<T> {
  fetcher: (query: DataTableQuery) => Promise<DataTableFetchResult<T> | null | undefined>
  initialPage?: number
  initialPageSize?: number
  pageSizeOptions?: number[]
  getErrorMessage?: (error: unknown) => ReactNode
  onError?: (error: unknown) => void
}

export interface DataTableController<T> {
  data: T[]
  total: number
  page: number
  pageSize: number
  pageSizeOptions: number[]
  loading: boolean
  error: ReactNode
  setPage: (page: number) => void
  setPageSize: (pageSize: number) => void
  reload: () => void
  resetPage: () => void
}

function isAuthNavigationError(error: unknown) {
  return error instanceof Errors.Unauthorized || error instanceof Errors.Forbidden
}

export function useDataTable<T>({
  fetcher,
  initialPage = 1,
  initialPageSize = 20,
  pageSizeOptions = [10, 20, 50],
  getErrorMessage,
  onError,
}: UseDataTableOptions<T>): DataTableController<T> {
  const [data, setData] = useState<T[]>([])
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(initialPage)
  const [pageSize, setPageSizeState] = useState(initialPageSize)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<ReactNode>(null)
  const [reloadKey, setReloadKey] = useState(0)

  const setPageSize = useCallback((nextPageSize: number) => {
    setPageSizeState(nextPageSize)
    setPage(1)
  }, [])

  const reload = useCallback(() => {
    setReloadKey(key => key + 1)
  }, [])

  const resetPage = useCallback(() => {
    setPage(1)
  }, [])

  useEffect(() => {
    let disposed = false

    async function loadData() {
      try {
        setLoading(true)
        setError(null)
        const result = await fetcher({ page, pageSize })

        if (disposed) return
        setData(result?.data ?? [])
        setTotal(result?.total ?? 0)

        if (result?.page && result.page !== page) {
          setPage(result.page)
        }
        if (result?.pageSize && result.pageSize !== pageSize) {
          setPageSizeState(result.pageSize)
        }
      } catch (err) {
        if (disposed) return
        console.error('Failed to load table data:', err)
        if (isAuthNavigationError(err)) return
        setError(getErrorMessage ? getErrorMessage(err) : 'Failed to load data')
        onError?.(err)
      } finally {
        if (!disposed) {
          setLoading(false)
        }
      }
    }

    loadData()

    return () => {
      disposed = true
    }
  }, [fetcher, getErrorMessage, onError, page, pageSize, reloadKey])

  return {
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
    resetPage,
  }
}
