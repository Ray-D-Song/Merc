import { useLocation } from 'preact-iso'

export function useNavigate() {
  const location = useLocation()

  return (target: string | number, replace = false) => {
    if (typeof target === 'number') {
      if (target < 0) {
        location.back()
      } else if (target > 0) {
        location.forward()
      }
      return
    }

    location.route(target, replace)
  }
}
