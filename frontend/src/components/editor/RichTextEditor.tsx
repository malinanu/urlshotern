'use client';

import { useState, useEffect, useRef } from 'react';
import {
  Bold,
  Italic,
  Underline,
  Link,
  Image,
  List,
  ListOrdered,
  Quote,
  Code,
  Type,
  AlignLeft,
  AlignCenter,
  AlignRight,
  Undo,
  Redo,
  Eye,
  EyeOff,
  Maximize2,
  Minimize2,
} from 'lucide-react';

interface RichTextEditorProps {
  content: string;
  onChange: (content: string) => void;
  placeholder?: string;
  height?: string;
  showPreview?: boolean;
  allowImages?: boolean;
  maxLength?: number;
  className?: string;
}

export default function RichTextEditor({
  content,
  onChange,
  placeholder = "Start writing...",
  height = "400px",
  showPreview = true,
  allowImages = true,
  maxLength,
  className = "",
}: RichTextEditorProps) {
  const [isPreview, setIsPreview] = useState(false);
  const [isFullscreen, setIsFullscreen] = useState(false);
  const [selectedText, setSelectedText] = useState('');
  const [wordCount, setWordCount] = useState(0);
  const editorRef = useRef<HTMLTextAreaElement>(null);
  const previewRef = useRef<HTMLDivElement>(null);

  // Update word count when content changes
  useEffect(() => {
    const text = content.replace(/<[^>]*>/g, '').trim();
    const words = text.split(/\s+/).filter(word => word.length > 0);
    setWordCount(words.length);
  }, [content]);

  // Handle text selection
  const handleSelection = () => {
    if (editorRef.current) {
      const start = editorRef.current.selectionStart;
      const end = editorRef.current.selectionEnd;
      const text = content.substring(start, end);
      setSelectedText(text);
    }
  };

  // Insert text at cursor position
  const insertText = (before: string, after: string = '') => {
    if (!editorRef.current) return;

    const start = editorRef.current.selectionStart;
    const end = editorRef.current.selectionEnd;
    const text = content;
    const selectedText = text.substring(start, end);
    
    let newText;
    if (selectedText) {
      // Wrap selected text
      newText = text.substring(0, start) + before + selectedText + after + text.substring(end);
    } else {
      // Insert at cursor
      newText = text.substring(0, start) + before + after + text.substring(start);
    }

    onChange(newText);

    // Restore cursor position
    setTimeout(() => {
      if (editorRef.current) {
        const newPosition = start + before.length + (selectedText ? selectedText.length : 0);
        editorRef.current.setSelectionRange(newPosition, newPosition);
        editorRef.current.focus();
      }
    }, 0);
  };

  // Format handlers
  const formatBold = () => insertText('**', '**');
  const formatItalic = () => insertText('*', '*');
  const formatUnderline = () => insertText('<u>', '</u>');
  const formatCode = () => insertText('`', '`');
  const formatQuote = () => insertText('\n> ', '');
  const formatH1 = () => insertText('\n# ', '');
  const formatH2 = () => insertText('\n## ', '');
  const formatH3 = () => insertText('\n### ', '');
  const formatList = () => insertText('\n- ', '');
  const formatNumberedList = () => insertText('\n1. ', '');
  const formatLink = () => {
    const url = prompt('Enter URL:');
    if (url) {
      insertText(`[${selectedText || 'Link text'}](${url})`);
    }
  };

  const formatImage = () => {
    const url = prompt('Enter image URL:');
    const alt = prompt('Enter alt text (optional):') || 'Image';
    if (url) {
      insertText(`![${alt}](${url})`);
    }
  };

  const formatAlign = (alignment: string) => {
    insertText(`<div style="text-align: ${alignment}">`, '</div>');
  };

  // Undo/Redo functionality (basic implementation)
  const [history, setHistory] = useState<string[]>([content]);
  const [historyIndex, setHistoryIndex] = useState(0);

  const addToHistory = (newContent: string) => {
    if (newContent !== history[historyIndex]) {
      const newHistory = history.slice(0, historyIndex + 1);
      newHistory.push(newContent);
      setHistory(newHistory);
      setHistoryIndex(newHistory.length - 1);
    }
  };

  const undo = () => {
    if (historyIndex > 0) {
      const newIndex = historyIndex - 1;
      setHistoryIndex(newIndex);
      onChange(history[newIndex]);
    }
  };

  const redo = () => {
    if (historyIndex < history.length - 1) {
      const newIndex = historyIndex + 1;
      setHistoryIndex(newIndex);
      onChange(history[newIndex]);
    }
  };

  // Add to history on content change (debounced)
  useEffect(() => {
    const timer = setTimeout(() => {
      addToHistory(content);
    }, 1000);
    return () => clearTimeout(timer);
  }, [content]);

  const toolbarButtons = [
    { icon: Undo, onClick: undo, title: 'Undo', disabled: historyIndex <= 0 },
    { icon: Redo, onClick: redo, title: 'Redo', disabled: historyIndex >= history.length - 1 },
    { divider: true },
    { icon: Bold, onClick: formatBold, title: 'Bold (Ctrl+B)' },
    { icon: Italic, onClick: formatItalic, title: 'Italic (Ctrl+I)' },
    { icon: Underline, onClick: formatUnderline, title: 'Underline (Ctrl+U)' },
    { icon: Code, onClick: formatCode, title: 'Inline Code' },
    { divider: true },
    { 
      text: 'H1', 
      onClick: formatH1, 
      title: 'Heading 1',
      className: 'font-bold text-xs px-2'
    },
    { 
      text: 'H2', 
      onClick: formatH2, 
      title: 'Heading 2',
      className: 'font-semibold text-xs px-2'
    },
    { 
      text: 'H3', 
      onClick: formatH3, 
      title: 'Heading 3',
      className: 'font-medium text-xs px-2'
    },
    { divider: true },
    { icon: List, onClick: formatList, title: 'Bullet List' },
    { icon: ListOrdered, onClick: formatNumberedList, title: 'Numbered List' },
    { icon: Quote, onClick: formatQuote, title: 'Quote' },
    { divider: true },
    { icon: Link, onClick: formatLink, title: 'Insert Link' },
    ...(allowImages ? [{ icon: Image, onClick: formatImage, title: 'Insert Image' }] : []),
    { divider: true },
    { icon: AlignLeft, onClick: () => formatAlign('left'), title: 'Align Left' },
    { icon: AlignCenter, onClick: () => formatAlign('center'), title: 'Align Center' },
    { icon: AlignRight, onClick: () => formatAlign('right'), title: 'Align Right' },
  ];

  const handleKeyDown = (e: React.KeyboardEvent) => {
    // Handle keyboard shortcuts
    if (e.ctrlKey || e.metaKey) {
      switch (e.key) {
        case 'b':
          e.preventDefault();
          formatBold();
          break;
        case 'i':
          e.preventDefault();
          formatItalic();
          break;
        case 'u':
          e.preventDefault();
          formatUnderline();
          break;
        case 'z':
          e.preventDefault();
          if (e.shiftKey) {
            redo();
          } else {
            undo();
          }
          break;
      }
    }
  };

  const renderPreview = () => {
    // Simple markdown-to-HTML conversion for preview
    let html = content
      .replace(/\*\*(.*?)\*\*/g, '<strong>$1</strong>')
      .replace(/\*(.*?)\*/g, '<em>$1</em>')
      .replace(/`(.*?)`/g, '<code class="bg-gray-100 px-1 rounded">$1</code>')
      .replace(/^> (.*$)/gm, '<blockquote class="border-l-4 border-gray-300 pl-4 italic">$1</blockquote>')
      .replace(/^# (.*$)/gm, '<h1 class="text-3xl font-bold mb-4">$1</h1>')
      .replace(/^## (.*$)/gm, '<h2 class="text-2xl font-semibold mb-3">$1</h2>')
      .replace(/^### (.*$)/gm, '<h3 class="text-xl font-medium mb-2">$1</h3>')
      .replace(/^- (.*$)/gm, '<ul class="list-disc ml-6"><li>$1</li></ul>')
      .replace(/^\d+\. (.*$)/gm, '<ol class="list-decimal ml-6"><li>$1</li></ol>')
      .replace(/\[([^\]]+)\]\(([^)]+)\)/g, '<a href="$2" class="text-blue-600 hover:underline" target="_blank">$1</a>')
      .replace(/!\[([^\]]*)\]\(([^)]+)\)/g, '<img src="$2" alt="$1" class="max-w-full h-auto" />')
      .replace(/\n/g, '<br />');

    return { __html: html };
  };

  const containerClass = `
    ${isFullscreen ? 'fixed inset-0 z-50 bg-white' : 'relative'} 
    ${className}
    border border-gray-300 rounded-lg overflow-hidden
  `;

  return (
    <div className={containerClass}>
      {/* Toolbar */}
      <div className="border-b border-gray-200 bg-gray-50 p-2 flex items-center gap-1 flex-wrap">
        {toolbarButtons.map((button, index) => {
          if (button.divider) {
            return <div key={index} className="w-px h-6 bg-gray-300 mx-1" />;
          }

          const buttonClass = `
            flex items-center justify-center w-8 h-8 rounded hover:bg-gray-200 
            transition-colors text-gray-600 hover:text-gray-800
            ${button.disabled ? 'opacity-50 cursor-not-allowed' : 'cursor-pointer'}
            ${button.className || ''}
          `;

          return (
            <button
              key={index}
              className={buttonClass}
              onClick={button.onClick}
              disabled={button.disabled}
              title={button.title}
              type="button"
            >
              {button.icon && <button.icon size={16} />}
              {button.text && button.text}
            </button>
          );
        })}

        <div className="flex-1" />

        {/* Right side controls */}
        {showPreview && (
          <button
            className="flex items-center justify-center w-8 h-8 rounded hover:bg-gray-200 transition-colors"
            onClick={() => setIsPreview(!isPreview)}
            title={isPreview ? 'Edit' : 'Preview'}
          >
            {isPreview ? <EyeOff size={16} /> : <Eye size={16} />}
          </button>
        )}
        
        <button
          className="flex items-center justify-center w-8 h-8 rounded hover:bg-gray-200 transition-colors"
          onClick={() => setIsFullscreen(!isFullscreen)}
          title={isFullscreen ? 'Exit Fullscreen' : 'Fullscreen'}
        >
          {isFullscreen ? <Minimize2 size={16} /> : <Maximize2 size={16} />}
        </button>
      </div>

      {/* Editor/Preview */}
      <div className="relative" style={{ height }}>
        {isPreview ? (
          <div
            ref={previewRef}
            className="p-4 h-full overflow-auto prose prose-sm max-w-none"
            dangerouslySetInnerHTML={renderPreview()}
          />
        ) : (
          <textarea
            ref={editorRef}
            value={content}
            onChange={(e) => onChange(e.target.value)}
            onSelect={handleSelection}
            onKeyDown={handleKeyDown}
            placeholder={placeholder}
            className="w-full h-full p-4 border-none resize-none focus:outline-none font-mono text-sm leading-relaxed"
            maxLength={maxLength}
          />
        )}
      </div>

      {/* Footer */}
      <div className="border-t border-gray-200 bg-gray-50 px-4 py-2 flex items-center justify-between text-xs text-gray-500">
        <div className="flex items-center gap-4">
          <span>{wordCount} words</span>
          <span>{content.length} characters</span>
          {maxLength && (
            <span className={content.length > maxLength * 0.9 ? 'text-orange-500' : ''}>
              {content.length}/{maxLength}
            </span>
          )}
        </div>
        
        <div className="flex items-center gap-2">
          <span>Markdown supported</span>
          <a
            href="https://www.markdownguide.org/basic-syntax/"
            target="_blank"
            rel="noopener noreferrer"
            className="text-blue-600 hover:underline"
          >
            Help
          </a>
        </div>
      </div>
    </div>
  );
}