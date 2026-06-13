import { Button } from '@cloudflare/kumo/components/button'
import { ArrowLeftIcon } from '@phosphor-icons/react'
import { type ComponentType } from 'react'
import { useNavigate } from '@/hooks/use-navigate'

export default function SubPageWrapper<T extends JSX.IntrinsicAttributes = object>(Component: ComponentType<T>) {
  return function WrappedComponent(props: T) {
    const navigate = useNavigate()

    return (
      <div className="space-y-4">
        <Button
          type="button"
          variant="ghost"
          icon={<ArrowLeftIcon size={16} />}
          onClick={() => navigate(-1)}
        >
          Back
        </Button>
        <Component {...props} />
      </div>
    )
  }
}
