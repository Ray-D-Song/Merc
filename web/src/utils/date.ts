import dayjs from 'dayjs'

export function formatDateTime(value?: string | null) {
  if (!value) return '-'

  const date = dayjs(value)
  return date.isValid() ? date.format('YYYY-MM-DD HH:mm:ss') : value
}
