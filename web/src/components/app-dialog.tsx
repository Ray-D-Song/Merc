import { Dialog } from '@cloudflare/kumo/components/dialog'
import { Button } from '@cloudflare/kumo/components/button'
import type { ButtonProps } from '@cloudflare/kumo/components/button'
import type { ReactNode } from 'react'
import { useTranslation } from 'react-i18next'

export interface AppDialogProps {
  open: boolean
  title: ReactNode
  description?: ReactNode
  children?: ReactNode
  confirmText?: ReactNode
  cancelText?: ReactNode
  confirmVariant?: 'primary' | 'destructive'
  loading?: boolean
  size?: 'sm' | 'base' | 'lg' | 'xl'
  role?: 'dialog' | 'alertdialog'
  onOpenChange: (open: boolean) => void
  onConfirm?: () => void
}

export function AppDialog({
  open,
  title,
  description,
  children,
  confirmText,
  cancelText,
  confirmVariant = 'primary',
  loading,
  size = 'base',
  role = 'dialog',
  onOpenChange,
  onConfirm,
}: AppDialogProps) {
  const { t } = useTranslation()
  const resolvedCancelText = cancelText ?? t('common.cancel')

  return (
    <Dialog.Root open={open} onOpenChange={onOpenChange} role={role}>
      <Dialog size={size} className="w-[calc(100vw-2rem)] max-w-2xl p-6">
        <div className="space-y-5">
          <div className="space-y-1">
            <Dialog.Title>{title}</Dialog.Title>
            {description && (
              <Dialog.Description>{description}</Dialog.Description>
            )}
          </div>
          {children}
          {(confirmText || onConfirm) && (
            <div className="flex justify-end gap-2">
              <Dialog.Close
                render={(props) => {
                  const closeProps = props as unknown as ButtonProps

                  return (
                    <Button {...closeProps} type="button" variant="secondary">
                      {resolvedCancelText}
                    </Button>
                  )
                }}
              />
              <Button
                type="button"
                variant={confirmVariant}
                loading={loading}
                disabled={loading}
                onClick={onConfirm}
              >
                {confirmText}
              </Button>
            </div>
          )}
        </div>
      </Dialog>
    </Dialog.Root>
  )
}
