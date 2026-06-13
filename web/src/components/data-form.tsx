import { forwardRef, type ComponentPropsWithoutRef, type ReactNode } from 'react'

interface DataFormProps extends Omit<ComponentPropsWithoutRef<'form'>, 'onSubmit' | 'children'> {
  children: ReactNode
  onSubmit: () => void | Promise<void>
}

export const DataForm = forwardRef<HTMLFormElement, DataFormProps>(function DataForm({
  children,
  className = 'space-y-4',
  onSubmit,
  ...props
}, ref) {
  return (
    <form
      {...props}
      ref={ref}
      className={className}
      onSubmit={(event) => {
        event.preventDefault()
        void onSubmit()
      }}
    >
      {children}
    </form>
  )
})
