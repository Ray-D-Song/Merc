/// <reference types="vite/client" />

export interface FeedbackApi {
  success: (message: string) => void
  error: (message: string) => void
  info: (message: string) => void
  warning: (message: string) => void
}

declare global {
  interface Window {
    feedback?: {
      message: FeedbackApi
      notification: FeedbackApi
    }
  }
}
