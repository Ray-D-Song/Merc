import { useEffect, useState } from 'react'
import { Banner } from '@cloudflare/kumo/components/banner'
import { Button } from '@cloudflare/kumo/components/button'
import { Loader } from '@cloudflare/kumo/components/loader'
import { Table } from '@cloudflare/kumo/components/table'
import * as XLSX from 'xlsx'
import GLOBAL_CONFIG from '@/config'

interface ExcelPreviewProps {
  fileId: number
}

interface CellValue {
  [key: number]: string | number | boolean
  key: number
}

interface SheetColumn {
  title: string
  dataIndex: number
  key: number
}

interface SheetData {
  name: string
  data: CellValue[]
  columns: SheetColumn[]
}

export default function ExcelPreview({ fileId }: ExcelPreviewProps) {
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [sheets, setSheets] = useState<SheetData[]>([])
  const [activeSheetIndex, setActiveSheetIndex] = useState(0)

  useEffect(() => {
    const loadExcelFile = async () => {
      setLoading(true)
      setError(null)

      try {
        const response = await fetch(`${GLOBAL_CONFIG.BASE_URL}/file_attachments/${fileId}/download`, {
          headers: {
            'Authorization': `Bearer ${localStorage.getItem('token') || ''}`
          }
        })

        if (!response.ok) {
          throw new Error('Failed to fetch file')
        }

        const arrayBuffer = await response.arrayBuffer()

        if (arrayBuffer.byteLength > 10 * 1024 * 1024) {
          throw new Error('File too large for preview. Maximum supported size is 10MB.')
        }

        const workbook = XLSX.read(arrayBuffer, { type: 'array' })

        if (!workbook.SheetNames || workbook.SheetNames.length === 0) {
          throw new Error('No sheets found in Excel file')
        }

        const sheetsData: SheetData[] = workbook.SheetNames.map((sheetName) => {
          const worksheet = workbook.Sheets[sheetName]

          if (!worksheet) {
            return { name: sheetName, data: [], columns: [] }
          }

          const jsonData = XLSX.utils.sheet_to_json(worksheet, {
            header: 1,
            defval: '',
            blankrows: false
          }) as (string | number | boolean)[][]

          if (jsonData.length === 0) {
            return { name: sheetName, data: [], columns: [] }
          }

          const headers = jsonData[0] || []
          const maxCols = Math.max(...jsonData.map(row => row.length))

          const columns: SheetColumn[] = Array.from({ length: maxCols }, (_, index) => ({
            title: String(headers[index] || `Column ${index + 1}`),
            dataIndex: index,
            key: index,
          }))

          const dataRows: CellValue[] = jsonData.slice(1).map((row, rowIndex) => {
            const rowData: CellValue = { key: rowIndex }
            for (let i = 0; i < maxCols; i++) {
              rowData[i] = row[i] || ''
            }
            return rowData
          })

          return { name: sheetName, data: dataRows, columns }
        })

        setSheets(sheetsData)
        setActiveSheetIndex(0)
      } catch (err) {
        console.error('Excel Preview Failed:', err)
        let errorMessage = 'Excel Preview Failed'

        if (err instanceof Error) {
          if (err.message.includes('File too large')) {
            errorMessage = 'File too large for browser preview. Please try a file smaller than 10MB.'
          } else if (err.message.includes('Failed to fetch')) {
            errorMessage = 'File download failed, please try again later'
          } else if (err.message.includes('not supported or corrupted')) {
            errorMessage = 'Excel file format not supported or corrupted'
          }
        }

        setError(errorMessage)
      } finally {
        setLoading(false)
      }
    }

    loadExcelFile()
  }, [fileId])

  if (loading) {
    return (
      <div className="flex min-h-96 items-center justify-center gap-3 text-kumo-subtle">
        <Loader size="lg" aria-label="Loading Excel file" />
        <span>Loading Excel file...</span>
      </div>
    )
  }

  if (error) {
    return <Banner variant="error" title="Preview Failed" description={error} />
  }

  if (sheets.length === 0) {
    return <Banner title="File is empty" description="This Excel file contains no data" />
  }

  const currentSheet = sheets[activeSheetIndex]

  return (
    <div className="space-y-4 p-4">
      {sheets.length > 1 && (
        <div className="flex flex-wrap gap-2">
          {sheets.map((sheet, index) => (
            <Button
              key={sheet.name}
              type="button"
              size="sm"
              variant={index === activeSheetIndex ? 'primary' : 'secondary'}
              onClick={() => setActiveSheetIndex(index)}
            >
              {sheet.name}
            </Button>
          ))}
        </div>
      )}

      <div className="max-h-[600px] overflow-auto rounded-lg border border-kumo-line">
        <Table className="min-w-max">
          <Table.Header sticky>
            <Table.Row>
              {currentSheet.columns.map(column => (
                <Table.Head key={column.key}>{column.title}</Table.Head>
              ))}
            </Table.Row>
          </Table.Header>
          <Table.Body>
            {currentSheet.data.slice(0, 500).map(row => (
              <Table.Row key={row.key}>
                {currentSheet.columns.map(column => (
                  <Table.Cell key={column.key}>
                    {String(row[column.dataIndex] || '')}
                  </Table.Cell>
                ))}
              </Table.Row>
            ))}
          </Table.Body>
        </Table>
      </div>
      {currentSheet.data.length > 500 && (
        <p className="text-sm text-kumo-subtle">Showing first 500 rows of {currentSheet.data.length}.</p>
      )}
    </div>
  )
}
