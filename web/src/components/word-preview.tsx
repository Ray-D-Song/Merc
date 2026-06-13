import { useState, useEffect, useRef } from 'react'
import { Banner } from '@cloudflare/kumo/components/banner'
import { Loader } from '@cloudflare/kumo/components/loader'
import { renderAsync } from 'docx-preview'
import GLOBAL_CONFIG from '@/config'

interface WordPreviewProps {
  fileId: number
  fileName: string
}

export default function WordPreview({ fileId }: WordPreviewProps) {
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const containerRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    const loadWordDocument = async () => {
      setLoading(true)
      setError(null)

      if (!containerRef.current) {
        setLoading(false)
        return
      }

      try {
        // Fetch file
        const response = await fetch(`${GLOBAL_CONFIG.BASE_URL}/file_attachments/${fileId}/download`, {
          headers: {
            'Authorization': `Bearer ${localStorage.getItem('token') || ''}`
          }
        })

        if (!response.ok) {
          throw new Error(`Failed to fetch document: ${response.status} ${response.statusText}`)
        }

        const arrayBuffer = await response.arrayBuffer()

        // Check file size (20MB)
        if (arrayBuffer.byteLength > 20 * 1024 * 1024) {
          throw new Error('Document too large for preview. Maximum supported size is 20MB.')
        }

        // Clear container
        containerRef.current.innerHTML = ''

        // Add timeout handling
        const renderPromise = renderAsync(arrayBuffer, containerRef.current, undefined, {
          className: 'docx-preview',
          inWrapper: false,
          ignoreWidth: false,
          ignoreHeight: false,
          ignoreFonts: false,
          breakPages: true,
          ignoreLastRenderedPageBreak: true,
          experimental: false,
          trimXmlDeclaration: true,
          useBase64URL: false,
          debug: false
        })

        const timeoutPromise = new Promise((_, reject) => {
          setTimeout(() => {
            reject(new Error('Word document rendering timeout (30s)'))
          }, 30000)
        })

        // Use Promise.race to implement timeout mechanism
        await Promise.race([renderPromise, timeoutPromise])

        // Check rendering result
        if (containerRef.current && containerRef.current.innerHTML.length === 0) {
          throw new Error('Word document rendered with empty content, document format may not be supported')
        }

        setLoading(false)

      } catch (error) {
        console.error('Word document preview failed:', error)
        setLoading(false)

        let errorMessage = 'Word document preview failed'

        if (error instanceof Error) {
          if (error.message.includes('too large')) {
            errorMessage = 'Word document too large for browser preview. Please try a file smaller than 20MB.'
          } else if (error.message.includes('Failed to fetch')) {
            errorMessage = 'Document download failed, please try again later'
          } else if (error.message.includes('not supported') || error.message.includes('corrupted')) {
            errorMessage = 'Word document format not supported or corrupted'
          } else {
            errorMessage = `Word document preview failed: ${error.message}`
          }
        }

        setError(errorMessage)
      }
    }

    loadWordDocument()
  }, [fileId])

  return (
    <div className="relative h-full overflow-auto bg-white p-4">
      {loading && (
        <div className="absolute inset-0 z-10 flex items-center justify-center gap-3 bg-white/90 text-kumo-subtle">
          <Loader size="lg" aria-label="Loading Word document" />
          <span>Loading Word document...</span>
        </div>
      )}

      {error && (
        <div className="absolute inset-0 z-10 flex items-center justify-center bg-white p-5">
          <Banner variant="error" title="Preview Failed" description={error} />
        </div>
      )}

      <div
        ref={containerRef}
        style={{
          maxWidth: '21cm', // A4 width
          margin: '0 auto',
          minHeight: '400px',
          background: '#fff',
          boxShadow: '0 0 10px rgba(0,0,0,0.1)',
          padding: '2cm',
          visibility: loading || error ? 'hidden' : 'visible'
        }}
      />
      <style>
        {`
          .docx-preview {
            font-family: 'Times New Roman', serif;
            line-height: 1.6;
            color: #000;
          }

          .docx-preview p {
            margin: 0 0 12px 0;
          }

          .docx-preview h1,
          .docx-preview h2,
          .docx-preview h3,
          .docx-preview h4,
          .docx-preview h5,
          .docx-preview h6 {
            margin: 16px 0 8px 0;
            font-weight: bold;
          }

          .docx-preview table {
            border-collapse: collapse;
            width: 100%;
            margin: 12px 0;
          }

          .docx-preview table td,
          .docx-preview table th {
            border: 1px solid #ddd;
            padding: 8px;
            text-align: left;
          }

          .docx-preview table th {
            background-color: #f2f2f2;
            font-weight: bold;
          }

          .docx-preview img {
            max-width: 100%;
            height: auto;
            margin: 12px 0;
          }

          .docx-preview ul,
          .docx-preview ol {
            margin: 12px 0;
            padding-left: 30px;
          }

          .docx-preview li {
            margin: 4px 0;
          }
        `}
      </style>
    </div>
  )
}
