"use client"

import { useCallback, useRef, useState, type DragEvent } from "react"
import { UploadCloud, FileText, X } from "lucide-react"
import { Alert, AlertDescription } from "@/components/ui/alert"

interface UploadAreaProps {
  file: File | null
  onFileSelect: (file: File | null) => void
  error: string | null
  onError: (error: string | null) => void
}

export function UploadArea({ file, onFileSelect, error, onError }: UploadAreaProps) {
  const [isDragOver, setIsDragOver] = useState(false)
  const inputRef = useRef<HTMLInputElement>(null)

  const validateFile = useCallback(
    (f: File) => {
      const isPdf =
        f.type === "application/pdf" || f.name.toLowerCase().endsWith(".pdf")
      if (!isPdf) {
        onError("Formato inválido. Por favor, envie um arquivo PDF.")
        return false
      }
      onError(null)
      return true
    },
    [onError]
  )

  const handleFile = useCallback(
    (f: File) => {
      if (validateFile(f)) {
        onFileSelect(f)
      }
    },
    [validateFile, onFileSelect]
  )

  const handleDrop = useCallback(
    (e: DragEvent<HTMLDivElement>) => {
      e.preventDefault()
      setIsDragOver(false)
      const droppedFile = e.dataTransfer.files[0]
      if (droppedFile) handleFile(droppedFile)
    },
    [handleFile]
  )

  const handleDragOver = useCallback((e: DragEvent<HTMLDivElement>) => {
    e.preventDefault()
    setIsDragOver(true)
  }, [])

  const handleDragLeave = useCallback((e: DragEvent<HTMLDivElement>) => {
    e.preventDefault()
    setIsDragOver(false)
  }, [])

  const handleInputChange = useCallback(
    (e: React.ChangeEvent<HTMLInputElement>) => {
      const selectedFile = e.target.files?.[0]
      if (selectedFile) handleFile(selectedFile)
      if (inputRef.current) inputRef.current.value = ""
    },
    [handleFile]
  )

  const removeFile = useCallback(() => {
    onFileSelect(null)
    onError(null)
  }, [onFileSelect, onError])

  function formatSize(bytes: number) {
    if (bytes < 1024) return `${bytes} B`
    if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`
    return `${(bytes / (1024 * 1024)).toFixed(1)} MB`
  }

  return (
    <div className="flex flex-col gap-4">
      {error && (
        <Alert variant="destructive">
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      )}

      {!file ? (
        <div
          role="button"
          tabIndex={0}
          aria-label="Arraste ou clique para enviar um arquivo PDF"
          onDrop={handleDrop}
          onDragOver={handleDragOver}
          onDragLeave={handleDragLeave}
          onClick={() => inputRef.current?.click()}
          onKeyDown={(e) => {
            if (e.key === "Enter" || e.key === " ") {
              e.preventDefault()
              inputRef.current?.click()
            }
          }}
          className={`group flex cursor-pointer flex-col items-center justify-center gap-3 rounded-lg border-2 border-dashed px-6 py-12 transition-colors ${
            isDragOver
              ? "border-accent bg-accent/5"
              : "border-border hover:border-muted-foreground/40 hover:bg-secondary/50"
          }`}
        >
          <div className="flex h-12 w-12 items-center justify-center rounded-full bg-secondary">
            <UploadCloud className="h-6 w-6 text-muted-foreground transition-colors group-hover:text-foreground" />
          </div>
          <div className="text-center">
            <p className="text-sm font-medium text-foreground">
              Arraste e solte seu PDF aqui
            </p>
            <p className="mt-1 text-xs text-muted-foreground">
              ou clique para selecionar um arquivo
            </p>
          </div>
          <input
            ref={inputRef}
            type="file"
            accept=".pdf,application/pdf"
            className="sr-only"
            onChange={handleInputChange}
            aria-label="Selecionar arquivo PDF"
          />
        </div>
      ) : (
        <div className="flex items-center gap-3 rounded-lg border border-border bg-secondary/50 px-4 py-3">
          <div className="flex h-10 w-10 shrink-0 items-center justify-center rounded-md bg-primary/10">
            <FileText className="h-5 w-5 text-foreground" />
          </div>
          <div className="min-w-0 flex-1">
            <p className="truncate text-sm font-medium text-foreground">
              {file.name}
            </p>
            <p className="text-xs text-muted-foreground">
              {formatSize(file.size)}
            </p>
          </div>
          <button
            type="button"
            onClick={removeFile}
            className="flex h-8 w-8 shrink-0 items-center justify-center rounded-md text-muted-foreground transition-colors hover:bg-secondary hover:text-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
            aria-label="Remover arquivo"
          >
            <X className="h-4 w-4" />
          </button>
        </div>
      )}
    </div>
  )
}
