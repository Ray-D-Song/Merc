import { useCallback, useEffect, useRef } from 'react'
import { $getRoot, $getSelection, type EditorState, FORMAT_TEXT_COMMAND, UNDO_COMMAND, REDO_COMMAND, $isRangeSelection } from 'lexical'
import { LexicalComposer } from '@lexical/react/LexicalComposer'
import { ContentEditable } from '@lexical/react/LexicalContentEditable'
import { HistoryPlugin } from '@lexical/react/LexicalHistoryPlugin'
import { OnChangePlugin } from '@lexical/react/LexicalOnChangePlugin'
import { useLexicalComposerContext } from '@lexical/react/LexicalComposerContext'
import { LexicalErrorBoundary } from '@lexical/react/LexicalErrorBoundary'
import { RichTextPlugin } from '@lexical/react/LexicalRichTextPlugin'
import { ListPlugin } from '@lexical/react/LexicalListPlugin'
import { LinkPlugin } from '@lexical/react/LexicalLinkPlugin'
import { $createTextNode } from 'lexical'
import { $createHeadingNode, $createQuoteNode, HeadingNode, QuoteNode } from '@lexical/rich-text'
import { $setBlocksType } from '@lexical/selection'
import { $createListNode, $createListItemNode, ListNode, ListItemNode } from '@lexical/list'
import { LinkNode } from '@lexical/link'
import { $generateHtmlFromNodes, $generateNodesFromDOM } from '@lexical/html'
import { Button } from '@cloudflare/kumo/components/button'
import {
  ArrowCounterClockwiseIcon,
  ArrowClockwiseIcon,
  ListBulletsIcon,
  ListNumbersIcon,
  LinkIcon,
  QuotesIcon,
  TextBIcon,
  TextHOneIcon,
  TextHThreeIcon,
  TextHTwoIcon,
  TextItalicIcon,
  TextStrikethroughIcon,
  TextUnderlineIcon,
} from '@phosphor-icons/react'

interface LexicalEditorProps {
  value: string
  onChange: (newVal: string) => void
  placeholder?: string
}

const theme = {
  text: {
    bold: 'font-bold',
    italic: 'italic',
    underline: 'underline',
    strikethrough: 'line-through',
  },
  heading: {
    h1: 'text-2xl font-bold mb-2',
    h2: 'text-xl font-bold mb-2',
    h3: 'text-lg font-bold mb-2',
  },
  quote: 'border-l-4 border-gray-300 dark:border-gray-600 pl-4 italic bg-gray-50 dark:bg-gray-800 py-2',
  list: {
    ul: 'list-disc ml-4',
    ol: 'list-decimal ml-4',
    listitem: 'mb-1',
  },
  link: 'text-blue-500 dark:text-blue-400 underline cursor-pointer hover:text-blue-700 dark:hover:text-blue-300',
}

function ToolbarPlugin() {
  const [editor] = useLexicalComposerContext()

  const formatText = useCallback((format: 'bold' | 'italic' | 'underline' | 'strikethrough') => {
    editor.dispatchCommand(FORMAT_TEXT_COMMAND, format)
  }, [editor])

  const formatHeading = useCallback((headingSize: 'h1' | 'h2' | 'h3') => {
    editor.update(() => {
      const selection = $getSelection()
      if ($isRangeSelection(selection)) {
        $setBlocksType(selection, () => $createHeadingNode(headingSize))
      }
    })
  }, [editor])

  const formatQuote = useCallback(() => {
    editor.update(() => {
      const selection = $getSelection()
      if ($isRangeSelection(selection)) {
        $setBlocksType(selection, () => $createQuoteNode())
      }
    })
  }, [editor])

  const formatList = useCallback((listType: 'bullet' | 'number') => {
    editor.update(() => {
      const selection = $getSelection()
      if (selection) {
        const list = $createListNode(listType)
        const listItem = $createListItemNode()
        list.append(listItem)
        selection.insertNodes([list])
      }
    })
  }, [editor])

  const insertLink = useCallback(() => {
    const url = prompt('Please enter link URL:')
    if (url) {
      editor.update(() => {
        const selection = $getSelection()
        if (selection) {
          const textNode = $createTextNode(url)
          selection.insertNodes([textNode])
        }
      })
    }
  }, [editor])

  const undo = useCallback(() => {
    editor.dispatchCommand(UNDO_COMMAND, undefined)
  }, [editor])

  const redo = useCallback(() => {
    editor.dispatchCommand(REDO_COMMAND, undefined)
  }, [editor])

  return (
    <div className="flex flex-wrap gap-0.5 px-3 py-2 border-b border-b-solid border-gray-200 dark:border-gray-700">
      <Button
        onClick={() => formatText('bold')}
        variant="ghost"
        size="sm"
        shape="square"
        icon={<TextBIcon size={14} />}
        aria-label="Bold"
      />
      <Button
        onClick={() => formatText('italic')}
        variant="ghost"
        size="sm"
        shape="square"
        icon={<TextItalicIcon size={14} />}
        aria-label="Italic"
      />
      <Button
        onClick={() => formatText('underline')}
        variant="ghost"
        size="sm"
        shape="square"
        icon={<TextUnderlineIcon size={14} />}
        aria-label="Underline"
      />
      <Button
        onClick={() => formatText('strikethrough')}
        variant="ghost"
        size="sm"
        shape="square"
        icon={<TextStrikethroughIcon size={14} />}
        aria-label="Strikethrough"
      />

      <div className="w-px h-4 bg-gray-300 dark:bg-gray-600 mx-2 self-center" />

      <Button
        onClick={() => formatHeading('h1')}
        variant="ghost"
        size="sm"
        shape="square"
        icon={<TextHOneIcon size={14} />}
        aria-label="Heading 1"
      />
      <Button
        onClick={() => formatHeading('h2')}
        variant="ghost"
        size="sm"
        shape="square"
        icon={<TextHTwoIcon size={14} />}
        aria-label="Heading 2"
      />
      <Button
        onClick={() => formatHeading('h3')}
        variant="ghost"
        size="sm"
        shape="square"
        icon={<TextHThreeIcon size={14} />}
        aria-label="Heading 3"
      />

      <div className="w-px h-4 bg-gray-300 dark:bg-gray-600 mx-2 self-center" />

      <Button
        onClick={() => formatQuote()}
        variant="ghost"
        size="sm"
        shape="square"
        icon={<QuotesIcon size={14} />}
        aria-label="Quote"
      />
      <Button
        onClick={() => formatList('bullet')}
        variant="ghost"
        size="sm"
        shape="square"
        icon={<ListBulletsIcon size={14} />}
        aria-label="Bullet list"
      />
      <Button
        onClick={() => formatList('number')}
        variant="ghost"
        size="sm"
        shape="square"
        icon={<ListNumbersIcon size={14} />}
        aria-label="Numbered list"
      />
      <Button
        onClick={insertLink}
        variant="ghost"
        size="sm"
        shape="square"
        icon={<LinkIcon size={14} />}
        aria-label="Insert link"
      />

      <div className="w-px h-4 bg-gray-300 dark:bg-gray-600 mx-2 self-center" />

      <Button
        onClick={undo}
        variant="ghost"
        size="sm"
        shape="square"
        icon={<ArrowCounterClockwiseIcon size={14} />}
        aria-label="Undo"
      />
      <Button
        onClick={redo}
        variant="ghost"
        size="sm"
        shape="square"
        icon={<ArrowClockwiseIcon size={14} />}
        aria-label="Redo"
      />
    </div>
  )
}

function ValueUpdatePlugin({ value, onChange }: { value: string; onChange: (html: string) => void }) {
  const [editor] = useLexicalComposerContext()
  const isInternalUpdate = useRef(false)
  const lastValue = useRef('')
  const isFirstRender = useRef(true)

  useEffect(() => {
    if ((value !== lastValue.current && !isInternalUpdate.current) || isFirstRender.current) {
      lastValue.current = value
      isFirstRender.current = false

      editor.update(() => {
        const root = $getRoot()
        root.clear()

        if (value) {
          const parser = new DOMParser()
          const dom = parser.parseFromString(value, 'text/html')
          const nodes = $generateNodesFromDOM(editor, dom)
          root.append(...nodes)
        }
      })
    }
    isInternalUpdate.current = false
  }, [editor, value])

  const handleChange = useCallback((editorState: EditorState) => {
    isInternalUpdate.current = true
    editorState.read(() => {
      const htmlString = $generateHtmlFromNodes(editor, null)
      lastValue.current = htmlString
      onChange(htmlString)
    })
  }, [editor, onChange])

  return <OnChangePlugin onChange={handleChange} ignoreSelectionChange />
}

const initialConfig = {
  namespace: 'LexicalEditor',
  theme,
  nodes: [HeadingNode, QuoteNode, ListNode, ListItemNode, LinkNode],
  onError: (error: Error) => {
    console.error('Lexical error:', error)
  },
}

export default function LexicalEditor({
  value,
  onChange,
}: LexicalEditorProps) {
  return (
    <div
      className="border border-gray-200 dark:border-gray-700 rounded-lg overflow-hidden shadow-sm bg-white dark:bg-[#1F1F1F]"
      data-editor-container
    >
      <LexicalComposer initialConfig={initialConfig}>
        <ToolbarPlugin />
        <div className="relative">
          <RichTextPlugin
            contentEditable={
              <ContentEditable
                className="min-h-32 px-4 focus:outline-none dark:text-white"
                style={{ caretColor: '#000' }}
                spellcheck={false}
                autoCorrect="off"
                autoCapitalize="off"
              />
            }
            ErrorBoundary={LexicalErrorBoundary}
          />
          <HistoryPlugin />
          <ListPlugin />
          <LinkPlugin />
          <ValueUpdatePlugin value={value} onChange={onChange} />
        </div>
      </LexicalComposer>
    </div>
  )
}
