interface FileIconProps {
  fileName: string
  contentType?: string
  size?: number
  className?: string
}

// Get VSCode icon class name for file extension
const getFileIcon = (fileName: string, contentType?: string): string => {
  const extension = fileName.split('.').pop()?.toLowerCase() || ''

  // Determine by MIME type
  if (contentType) {
    if (contentType.startsWith('image/')) {
      return 'i-vscode-icons:file-type-image'
    }

    if (contentType.startsWith('video/')) {
      if (contentType.includes('mp4')) return 'i-vscode-icons:file-type-video'
      if (contentType.includes('avi')) return 'i-vscode-icons:file-type-video'
      if (contentType.includes('mov')) return 'i-vscode-icons:file-type-video'
      return 'i-vscode-icons:file-type-video'
    }

    if (contentType.includes('pdf')) return 'i-vscode-icons:file-type-pdf'
    if (contentType.includes('zip')) return 'i-vscode-icons:file-type-zip'
    if (contentType.includes('rar')) return 'i-vscode-icons:file-type-zip2'
  }

  // Determine by file extension
  const iconMap: Record<string, string> = {
    // Images
    png: 'i-vscode-icons:file-type-image',
    jpg: 'i-vscode-icons:file-type-image',
    jpeg: 'i-vscode-icons:file-type-image',
    gif: 'i-vscode-icons:file-type-image',
    svg: 'i-vscode-icons:file-type-image',
    webp: 'i-vscode-icons:file-type-image',
    bmp: 'i-vscode-icons:file-type-image',
    ico: 'i-vscode-icons:file-type-image',

    // Videos
    mp4: 'i-vscode-icons:file-type-video',
    avi: 'i-vscode-icons:file-type-video',
    mov: 'i-vscode-icons:file-type-video',
    wmv: 'i-vscode-icons:file-type-video',
    flv: 'i-vscode-icons:file-type-video',
    webm: 'i-vscode-icons:file-type-video',

    // Documents
    pdf: 'i-vscode-icons:file-type-pdf',
    doc: 'i-vscode-icons:file-type-word',
    docx: 'i-vscode-icons:file-type-word',
    xls: 'i-vscode-icons:file-type-excel',
    xlsx: 'i-vscode-icons:file-type-excel',
    ppt: 'i-vscode-icons:file-type-powerpoint',
    pptx: 'i-vscode-icons:file-type-powerpoint',
    txt: 'i-vscode-icons:file-type-text',

    // Archives
    zip: 'i-vscode-icons:file-type-zip',
    rar: 'i-vscode-icons:file-type-zip2',
    '7z': 'i-vscode-icons:file-type-zip2',
    tar: 'i-vscode-icons:file-type-zip2',
    gz: 'i-vscode-icons:file-type-zip2',

    // Code files
    js: 'i-vscode-icons:file-type-js',
    ts: 'i-vscode-icons:file-type-typescript',
    jsx: 'i-vscode-icons:file-type-reactjs',
    tsx: 'i-vscode-icons:file-type-reactts',
    vue: 'i-vscode-icons:file-type-vue',
    css: 'i-vscode-icons:file-type-css',
    scss: 'i-vscode-icons:file-type-scss',
    less: 'i-vscode-icons:file-type-less',
    html: 'i-vscode-icons:file-type-html',
    json: 'i-vscode-icons:file-type-json',
    xml: 'i-vscode-icons:file-type-xml',
    yaml: 'i-vscode-icons:file-type-yaml',
    yml: 'i-vscode-icons:file-type-yaml',

    // Other common files
    md: 'i-vscode-icons:file-type-markdown',
    py: 'i-vscode-icons:file-type-python',
    java: 'i-vscode-icons:file-type-java',
    php: 'i-vscode-icons:file-type-php',
    rb: 'i-vscode-icons:file-type-ruby',
    go: 'i-vscode-icons:file-type-go',
    rs: 'i-vscode-icons:file-type-rust',
    c: 'i-vscode-icons:file-type-c',
    cpp: 'i-vscode-icons:file-type-cpp',
    h: 'i-vscode-icons:file-type-h',
    sh: 'i-vscode-icons:file-type-shell',
    sql: 'i-vscode-icons:file-type-sql'
  }

  return iconMap[extension] || 'i-vscode-icons:default-file'
}

export default function FileIcon({
  fileName,
  contentType,
  size = 16,
  className = ''
}: FileIconProps) {
  const iconClass = getFileIcon(fileName, contentType)

  return (
    <div
      className={`${iconClass} ${className}`}
      style={{ width: size, height: size, display: 'inline-block' }}
    />
  )
}