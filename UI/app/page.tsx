"use client"

import { useState, useCallback } from "react"
import { toast } from "sonner"
import { Navbar } from "@/components/navbar"
import { UploadArea } from "@/components/upload-area"
import { ResultPreview } from "@/components/result-preview"
import { Card, CardContent, CardHeader } from "@/components/ui/card"
import { Button } from "@/components/ui/button"
import { Separator } from "@/components/ui/separator"
import { Loader2, ArrowRight, FileSearch } from "lucide-react"
import { analyzeContract, type AnalysisResult, type AnalysisStep } from "@/lib/api"

const stepLabels: Record<AnalysisStep, string> = {
  presigning: "Preparando upload...",
  uploading: "Enviando arquivo...",
  processing: "Analisando contrato com IA...",
}

export default function HomePage() {
  const [file, setFile] = useState<File | null>(null)
  const [error, setError] = useState<string | null>(null)
  const [isAnalyzing, setIsAnalyzing] = useState(false)
  const [showResult, setShowResult] = useState(false)
  const [result, setResult] = useState<AnalysisResult | null>(null)
  const [step, setStep] = useState<AnalysisStep | null>(null)

  const handleAnalyze = useCallback(async () => {
    if (!file) return
    setIsAnalyzing(true)
    setShowResult(true)
    setResult(null)

    try {
      const data = await analyzeContract(file, setStep)
      setResult(data)
      toast.success("Contrato analisado com sucesso!")
    } catch (err) {
      const message = err instanceof Error ? err.message : "Erro desconhecido"
      toast.error(message)
      setShowResult(false)
    } finally {
      setIsAnalyzing(false)
      setStep(null)
    }
  }, [file])

  const handleFileSelect = useCallback((f: File | null) => {
    setFile(f)
    setShowResult(false)
    setIsAnalyzing(false)
    setResult(null)
    setStep(null)
  }, [])

  return (
    <div className="min-h-screen bg-background">
      <div className="mx-auto max-w-3xl px-4">
        <Navbar />

        <main className="pb-16 pt-8 md:pt-16">
          <Card className="border-border bg-card shadow-sm">
            <CardHeader className="gap-2 pb-2 text-center">
              <div className="mx-auto mb-2 flex h-12 w-12 items-center justify-center rounded-full bg-secondary">
                <FileSearch className="h-6 w-6 text-foreground" />
              </div>
              <h1 className="text-balance text-2xl font-bold tracking-tight text-foreground md:text-3xl">
                Analise seu contrato em minutos
              </h1>
              <p className="mx-auto max-w-lg text-pretty text-sm leading-relaxed text-muted-foreground">
                Envie um PDF e receba um resumo claro dos riscos, cláusulas
                críticas e pontos para negociar.
              </p>
            </CardHeader>

            <CardContent className="flex flex-col gap-6 pt-4">
              <UploadArea
                file={file}
                onFileSelect={handleFileSelect}
                error={error}
                onError={setError}
              />

              <div className="flex flex-col items-center gap-3 sm:flex-row sm:justify-between">
                <Button
                  size="lg"
                  disabled={!file || isAnalyzing}
                  onClick={handleAnalyze}
                  className="w-full sm:w-auto"
                >
                  {isAnalyzing ? (
                    <>
                      <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                      {step ? stepLabels[step] : "Analisando..."}
                    </>
                  ) : (
                    <>
                      Analisar contrato
                      <ArrowRight className="ml-2 h-4 w-4" />
                    </>
                  )}
                </Button>
                <button
                  type="button"
                  className="text-sm text-muted-foreground underline underline-offset-4 transition-colors hover:text-foreground"
                >
                  Ver exemplo de relatório
                </button>
              </div>
            </CardContent>
          </Card>

          {showResult && (
            <div className="mt-8">
              <Separator className="mb-8" />
              <ResultPreview isLoading={isAnalyzing} result={result} />
            </div>
          )}
        </main>
      </div>
    </div>
  )
}
