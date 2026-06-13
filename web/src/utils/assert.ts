export function isMobile(): boolean {
  // Check if we're in a browser environment
  if (typeof window === 'undefined') {
    return false
  }

  // Check screen width - typically mobile devices are <= 768px
  const screenWidth = window.innerWidth || document.documentElement.clientWidth || document.body.clientWidth
  
  // Also check for touch capability as an additional indicator
  const hasTouchScreen = 'ontouchstart' in window || navigator.maxTouchPoints > 0
  
  // Consider it mobile if screen width is <= 768px or if it's a touch device with width <= 1024px
  return screenWidth <= 768 || (hasTouchScreen && screenWidth <= 1024)
}
