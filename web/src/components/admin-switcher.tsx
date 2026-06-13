import { useAuth } from '@/hooks/use-auth'
import { type ComponentType } from 'react'

export default function AdminSwitcher<T extends JSX.IntrinsicAttributes = object>(AdminComp: ComponentType<T>, ElseComp: ComponentType<T>) {
  return function SwitcherComponent(props: T) {
    const { isAdmin } = useAuth()
    console.log('isadmin', isAdmin())

    if (isAdmin()) {
      return <AdminComp {...props} />
    }

    return <ElseComp {...props} />
  }
}