// Common pagination types
export interface PageRequestBox {
  page: number
  size: number
}

export interface PageResponseBox<T> {
  list: T[]
  total: number
}

// Common response types
export interface ResponseSuccessBox<T = unknown> {
  status: "success"
  message?: string
  data: T
}

export interface ErrorBox {
  code: string
  details?: string
}

export interface ResponseFailBox {
  status: string
  message?: string
  error: ErrorBox
}

export type ResponseBox<T = unknown> = ResponseSuccessBox<T> | ResponseFailBox

export type Optional<T> = T | null

export type Either<T, U = unknown> = T | U