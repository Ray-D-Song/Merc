import { useRef, type ReactNode } from 'react'
import { AppDialog, type AppDialogProps } from './app-dialog'
import { DataForm } from './data-form'

interface DataFormDialogProps extends Omit<AppDialogProps, 'children' | 'onConfirm'> {
  children: ReactNode
  formClassName?: string
  formId?: string
  onSubmit: () => void | Promise<void>
}

export function DataFormDialog({
  children,
  formClassName,
  formId,
  onSubmit,
  ...dialogProps
}: DataFormDialogProps) {
  const formRef = useRef<HTMLFormElement>(null)

  return (
    <AppDialog
      {...dialogProps}
      onConfirm={() => {
        formRef.current?.requestSubmit()
      }}
    >
      <DataForm id={formId} ref={formRef} className={formClassName} onSubmit={onSubmit}>
        {children}
      </DataForm>
    </AppDialog>
  )
}
