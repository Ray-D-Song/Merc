/* eslint-disable react-refresh/only-export-components */
import { createContext, useContext, useEffect, useMemo, type ReactNode } from 'react'
import { createKumoToastManager } from '@cloudflare/kumo/components/toast'
import type { FeedbackApi } from '@/vite-env'

export const appToastManager = createKumoToastManager()

interface FeedbackContextType {
  message: FeedbackApi
  notification: FeedbackApi
}

const createFeedbackApi = (): FeedbackApi => {
  const show = (variant: 'success' | 'error' | 'info' | 'warning', message: string) => {
    appToastManager.add({
      title: message,
      variant,
    })
  }

  return {
    success: (message) => show('success', message),
    error: (message) => show('error', message),
    info: (message) => show('info', message),
    warning: (message) => show('warning', message),
  }
}

const messageApi = createFeedbackApi()
const notificationApi = createFeedbackApi()

const FeedbackContext = createContext<FeedbackContextType | undefined>(undefined)

export function FeedbackProvider({ children }: { children: ReactNode }) {
  const contextValue: FeedbackContextType = useMemo(() => ({
    message: messageApi,
    notification: notificationApi,
  }), [])

  useEffect(() => {
    window.feedback = contextValue

    return () => {
      window.feedback = undefined
    }
  }, [contextValue])

  return (
    <FeedbackContext.Provider value={contextValue}>
      {children}
    </FeedbackContext.Provider>
  )
}

export const useFeedback = (): FeedbackContextType => {
  const context = useContext(FeedbackContext)

  if (context === undefined) {
    throw new Error('useFeedback must be used within a FeedbackProvider')
  }

  return context
}

export const useMessage = (): FeedbackApi => {
  const { message } = useFeedback()
  return message
}

export const useNotification = (): FeedbackApi => {
  const { notification } = useFeedback()
  return notification
}
