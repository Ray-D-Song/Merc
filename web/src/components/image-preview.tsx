import { useEffect, useState } from 'react'
import { Banner } from '@cloudflare/kumo/components/banner'
import { Loader } from '@cloudflare/kumo/components/loader'
import GLOBAL_CONFIG from '@/config'

interface ImagePreviewProps {
  fileId: number
  fileName: string
}

export default function ImagePreview({ fileId, fileName }: ImagePreviewProps) {
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [imageUrl, setImageUrl] = useState<string | null>(null)

  useEffect(() => {
    let objectUrl: string | null = null

    const loadImage = async () => {
      setLoading(true)
      setError(null)

      try {
        const response = await fetch(`${GLOBAL_CONFIG.BASE_URL}/file_attachments/${fileId}/download`, {
          headers: {
            'Authorization': `Bearer ${localStorage.getItem('token') || ''}`
          }
        })

        if (!response.ok) {
          throw new Error('Failed to fetch image')
        }

        const blob = await response.blob()

        if (blob.size > 50 * 1024 * 1024) {
          throw new Error('Image too large for preview. Maximum supported size is 50MB.')
        }

        if (!blob.type.startsWith('image/')) {
          throw new Error('File is not a valid image')
        }

        objectUrl = URL.createObjectURL(blob)
        setImageUrl(objectUrl)
      } catch (err) {
        console.error('Image preview failed:', err)
        let errorMessage = 'Image preview failed'

        if (err instanceof Error) {
          if (err.message.includes('too large')) {
            errorMessage = 'Image file too large for browser preview. Please try a file smaller than 50MB.'
          } else if (err.message.includes('Failed to fetch')) {
            errorMessage = 'Image download failed, please try again later'
          } else if (err.message.includes('not a valid image')) {
            errorMessage = 'Invalid file format, not a valid image file'
          }
        }

        setError(errorMessage)
      } finally {
        setLoading(false)
      }
    }

    loadImage()

    return () => {
      if (objectUrl) URL.revokeObjectURL(objectUrl)
    }
  }, [fileId])

  if (loading) {
    return (
      <div className="flex min-h-96 items-center justify-center gap-3 text-kumo-subtle">
        <Loader size="lg" aria-label="Loading image" />
        <span>Loading image...</span>
      </div>
    )
  }

  if (error) {
    return <Banner variant="error" title="Preview Failed" description={error} />
  }

  if (!imageUrl) {
    return <Banner title="Image Unavailable" description="Unable to load image content" />
  }

  return (
    <div className="flex min-h-96 items-center justify-center bg-kumo-tint p-4">
      <img
        src={imageUrl}
        alt={fileName}
        className="max-h-[calc(100vh-12rem)] max-w-full object-contain"
      />
    </div>
  )
}
