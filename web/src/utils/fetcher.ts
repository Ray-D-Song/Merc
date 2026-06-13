import GLOBAL_CONFIG from '@/config'
import type { Optional } from './types'
import { Errors } from './error'

export interface ApiError {
  message: string
  statusCode?: number
}

export interface FetchOptions extends RequestInit {
  headers?: Record<string, string>
}

export interface FetcherOpts {
  maxRetries?: number
  retryDelay?: number
  retryCondition?: (error: Error, attempt: number) => boolean

  hideErrorMessage?: boolean
  showSuccessMessage?: boolean
  errorMessageOverride?: string

  timeout?: number
  silent?: boolean // Silent failure, exceptions are not thrown
  abortController?: AbortController
}

function isAuthNavigationError(error: unknown) {
  return error instanceof Errors.Unauthorized || error instanceof Errors.Forbidden
}

function redirectForAuthError(error: unknown) {
  const target = error instanceof Errors.Forbidden ? '/403' : '/login'

  if (window.location.pathname === target) return

  window.history.replaceState(null, '', target)
  window.dispatchEvent(new PopStateEvent('popstate'))
}

export async function fetcher<T>(
  endpoint: string,
  fetchOptions: FetchOptions = {},
  options: FetcherOpts = {}
): Promise<Optional<T>> {
  const url = `${GLOBAL_CONFIG.BASE_URL}${endpoint}`
  const maxRetries = options.maxRetries ?? 0
  const retryDelay = options.retryDelay ?? 1000
  const retryCondition = options.retryCondition ?? ((_error: Error, attempt: number) => 
    attempt < maxRetries
  )

  const headers: Record<string, string> = {
    ...fetchOptions.headers,
  }

  // Only set Content-Type if not already set and body is not FormData
  if (!headers['Content-Type'] && !(fetchOptions.body instanceof FormData)) {
    headers['Content-Type'] = 'application/json'
  }

  // Setup timeout and abort controller
  const abortController = options.abortController || new AbortController()
  let timeoutId: ReturnType<typeof setTimeout> | undefined

  if (options.timeout) {
    timeoutId = setTimeout(() => {
      abortController.abort()
    }, options.timeout)
  }

  const config: RequestInit = {
    ...fetchOptions,
    headers,
    credentials: 'include', // Include cookies in requests
    signal: abortController.signal,
  }

  let lastError: Error
  let attempt = 0

  while (attempt <= maxRetries) {
    try {
      const response = await fetch(url, config)

      if (timeoutId) {
        clearTimeout(timeoutId)
      }

      if (!response.ok) {
        // Try to read error message from response body (backend returns a simple error object)
        let errorMessage = response.statusText || 'Network response was not ok'

        try {
          const errorData = await response.json()
          // Backend returns { "error": "error message" }
          if (errorData && typeof errorData === 'object' && 'error' in errorData) {
            errorMessage = errorData.error
          }
        } catch {
          // If JSON parsing fails, use statusText
        }

        // Create appropriate error based on status code
        let fetchError: Error
        switch (response.status) {
          case 400:
            fetchError = new Errors.InvalidParameter(errorMessage)
            break
          case 401:
            fetchError = new Errors.Unauthorized(errorMessage)
            break
          case 403:
            fetchError = new Errors.Forbidden(errorMessage)
            break
          case 404:
            fetchError = new Errors.ResourceNotFound(errorMessage)
            break
          case 409:
            fetchError = new Errors.ResourceAlreadyExists(errorMessage)
            break
          case 422:
            fetchError = new Errors.ValidationFailed(errorMessage)
            break
          case 500:
            fetchError = new Errors.InternalServerError(errorMessage)
            break
          default:
            fetchError = new Errors.Unknown(errorMessage)
        }

        // Check if we should retry
        if (attempt < maxRetries && retryCondition(fetchError, attempt)) {
          lastError = fetchError
          attempt++
          await new Promise(resolve => setTimeout(resolve, retryDelay))
          continue
        }

        throw fetchError
      }

      const resp = await response.json()
      const result = GLOBAL_CONFIG.RESPONSE_HANDLER<T>(resp)

      // Show success message if explicitly enabled and resp has a message
      if (window.feedback && options.showSuccessMessage && resp && typeof resp === 'object' && 'message' in resp) {
        window.feedback.message.success((resp as { message: string }).message)
      }

      return result
    } catch (error) {
      if (timeoutId) {
        clearTimeout(timeoutId)
      }

      lastError = error instanceof Error ? error : new Errors.Unknown('Unknown error')

      // Handle abort/timeout
      if (abortController.signal.aborted) {
        if (options.silent) {
          console.error('Request aborted')
          return null
        }
        const timeoutError = options.timeout
          ? new Errors.RequestTimeout('Request timeout')
          : new Errors.NetworkError('Request aborted')
        if (window.feedback && !options.hideErrorMessage) {
          window.feedback.message.error(options.errorMessageOverride || timeoutError.message)
        }
        throw timeoutError
      }

      // Check if we should retry
      if (attempt < maxRetries && retryCondition(lastError, attempt)) {
        attempt++
        await new Promise(resolve => setTimeout(resolve, retryDelay))
        continue
      }

      // The network errors are all caught in the process
      if (options.silent) {
        console.error(lastError)
        return null
      }

      console.error(lastError)

      if (isAuthNavigationError(lastError)) {
        redirectForAuthError(lastError)
        throw lastError
      }

      // Show error message using global feedback if available
      if (window.feedback && !options.hideErrorMessage) {
        window.feedback.message.error(options.errorMessageOverride || (lastError as Error).message)
      }

      throw lastError
    }
  }

  // This should never be reached, but just in case
  throw lastError!
}

// Convenience methods
export const http = {
  get: <T>(endpoint: string, fetchOptions?: FetchOptions, options?: FetcherOpts) =>
    fetcher<T>(endpoint, { ...fetchOptions, method: 'GET' }, options),

  post: <T>(endpoint: string, data?: unknown, fetchOptions?: FetchOptions, options?: FetcherOpts) =>
    fetcher<T>(endpoint, { ...fetchOptions, method: 'POST', body: JSON.stringify(data) }, options),

  postFormData: <T>(endpoint: string, formData: FormData, fetchOptions?: FetchOptions, options?: FetcherOpts) => {
    const headers = { ...fetchOptions?.headers }
    // Remove Content-Type to let browser set it with boundary for FormData
    delete headers['Content-Type']
    return fetcher<T>(endpoint, {
      ...fetchOptions,
      method: 'POST',
      body: formData,
      headers
    }, options)
  },

  put: <T>(endpoint: string, data?: unknown, fetchOptions?: FetchOptions, options?: FetcherOpts) =>
    fetcher<T>(endpoint, { ...fetchOptions, method: 'PUT', body: JSON.stringify(data) }, options),

  patch: <T>(endpoint: string, data?: unknown, fetchOptions?: FetchOptions, options?: FetcherOpts) =>
    fetcher<T>(endpoint, { ...fetchOptions, method: 'PATCH', body: JSON.stringify(data) }, options),

  delete: <T>(endpoint: string, fetchOptions?: FetchOptions, options?: FetcherOpts) =>
    fetcher<T>(endpoint, { ...fetchOptions, method: 'DELETE' }, options),
}
