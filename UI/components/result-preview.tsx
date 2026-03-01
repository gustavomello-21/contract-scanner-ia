import { Skeleton } from "@/components/ui/skeleton"
import { Card, CardContent, CardHeader } from "@/components/ui/card"
import { Badge } from "@/components/ui/badge"
import {
  Accordion,
  AccordionContent,
  AccordionItem,
  AccordionTrigger,
} from "@/components/ui/accordion"
import {
  AlertTriangle,
  Scale,
  MessageSquareText,
  FileText,
  Users,
  ShieldAlert,
  Lightbulb,
} from "lucide-react"
import type { AnalysisResult, RiskLevel } from "@/lib/api"

interface ResultPreviewProps {
  isLoading: boolean
  result: AnalysisResult | null
}

const riskConfig: Record<RiskLevel, { label: string; variant: "default" | "secondary" | "destructive"; className?: string }> = {
  low: { label: "Baixo", variant: "secondary" },
  medium: { label: "Moderado", variant: "default", className: "bg-yellow-500/90 text-white hover:bg-yellow-500" },
  high: { label: "Alto", variant: "destructive" },
}

function RiskBadge({ level }: { level: RiskLevel }) {
  const config = riskConfig[level]
  return (
    <Badge variant={config.variant} className={config.className}>
      {config.label}
    </Badge>
  )
}

function LoadingSkeleton() {
  return (
    <div className="flex flex-col gap-6">
      <div className="flex items-center gap-2">
        <div className="h-1 w-6 rounded-full bg-accent" />
        <Skeleton className="h-5 w-40" />
      </div>
      <div className="grid gap-4 sm:grid-cols-3">
        {[3, 4, 3].map((lines, i) => (
          <Card key={i} className="border-border bg-card">
            <CardHeader className="flex flex-row items-center gap-3 pb-3">
              <Skeleton className="h-8 w-8 rounded-md" />
              <Skeleton className="h-4 w-28" />
            </CardHeader>
            <CardContent className="flex flex-col gap-2">
              {Array.from({ length: lines }).map((_, j) => (
                <Skeleton key={j} className="h-3 rounded" style={{ width: `${70 + Math.random() * 30}%` }} />
              ))}
            </CardContent>
          </Card>
        ))}
      </div>
    </div>
  )
}

export function ResultPreview({ isLoading, result }: ResultPreviewProps) {
  if (isLoading || !result) {
    return <LoadingSkeleton />
  }

  const { summary, parties, contract_type, key_clauses, risks, recommendations, overall_risk } = result

  return (
    <div className="flex flex-col gap-6">
      {/* Header com risco geral */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-2">
          <div className="h-1 w-6 rounded-full bg-accent" />
          <h2 className="text-lg font-semibold text-foreground">Resultado da Análise</h2>
        </div>
        <div className="flex items-center gap-2">
          <span className="text-sm text-muted-foreground">Risco geral:</span>
          <RiskBadge level={overall_risk} />
        </div>
      </div>

      {/* Resumo */}
      <Card className="border-border bg-card">
        <CardHeader className="flex flex-row items-center gap-3 pb-3">
          <div className="flex h-8 w-8 items-center justify-center rounded-md bg-secondary">
            <FileText className="h-4 w-4 text-muted-foreground" />
          </div>
          <h3 className="text-sm font-medium text-foreground">Resumo</h3>
        </CardHeader>
        <CardContent>
          <p className="text-sm leading-relaxed text-muted-foreground">{summary}</p>
        </CardContent>
      </Card>

      {/* Partes + Tipo de Contrato */}
      <div className="grid gap-4 sm:grid-cols-2">
        <Card className="border-border bg-card">
          <CardHeader className="flex flex-row items-center gap-3 pb-3">
            <div className="flex h-8 w-8 items-center justify-center rounded-md bg-secondary">
              <Users className="h-4 w-4 text-muted-foreground" />
            </div>
            <h3 className="text-sm font-medium text-foreground">Partes</h3>
          </CardHeader>
          <CardContent>
            <ul className="flex flex-col gap-1">
              {parties.map((party, i) => (
                <li key={i} className="text-sm text-muted-foreground">{party}</li>
              ))}
            </ul>
          </CardContent>
        </Card>

        <Card className="border-border bg-card">
          <CardHeader className="flex flex-row items-center gap-3 pb-3">
            <div className="flex h-8 w-8 items-center justify-center rounded-md bg-secondary">
              <Scale className="h-4 w-4 text-muted-foreground" />
            </div>
            <h3 className="text-sm font-medium text-foreground">Tipo de Contrato</h3>
          </CardHeader>
          <CardContent>
            <p className="text-sm text-muted-foreground">{contract_type}</p>
          </CardContent>
        </Card>
      </div>

      {/* Cláusulas Críticas */}
      <Card className="border-border bg-card">
        <CardHeader className="flex flex-row items-center gap-3 pb-3">
          <div className="flex h-8 w-8 items-center justify-center rounded-md bg-secondary">
            <Scale className="h-4 w-4 text-muted-foreground" />
          </div>
          <h3 className="text-sm font-medium text-foreground">Cláusulas Críticas</h3>
        </CardHeader>
        <CardContent>
          <Accordion type="multiple" className="w-full">
            {key_clauses.map((clause, i) => (
              <AccordionItem key={i} value={`clause-${i}`}>
                <AccordionTrigger>
                  <div className="flex items-center gap-2">
                    <span>{clause.clause}</span>
                    <RiskBadge level={clause.risk_level} />
                  </div>
                </AccordionTrigger>
                <AccordionContent>
                  <p className="text-muted-foreground">{clause.description}</p>
                </AccordionContent>
              </AccordionItem>
            ))}
          </Accordion>
        </CardContent>
      </Card>

      {/* Riscos + Recomendações */}
      <div className="grid gap-4 sm:grid-cols-2">
        <Card className="border-border bg-card">
          <CardHeader className="flex flex-row items-center gap-3 pb-3">
            <div className="flex h-8 w-8 items-center justify-center rounded-md bg-secondary">
              <ShieldAlert className="h-4 w-4 text-muted-foreground" />
            </div>
            <h3 className="text-sm font-medium text-foreground">Riscos Identificados</h3>
          </CardHeader>
          <CardContent>
            <ul className="flex flex-col gap-2">
              {risks.map((risk, i) => (
                <li key={i} className="flex items-start gap-2 text-sm text-muted-foreground">
                  <AlertTriangle className="mt-0.5 h-3.5 w-3.5 shrink-0 text-destructive" />
                  {risk}
                </li>
              ))}
            </ul>
          </CardContent>
        </Card>

        <Card className="border-border bg-card">
          <CardHeader className="flex flex-row items-center gap-3 pb-3">
            <div className="flex h-8 w-8 items-center justify-center rounded-md bg-secondary">
              <Lightbulb className="h-4 w-4 text-muted-foreground" />
            </div>
            <h3 className="text-sm font-medium text-foreground">Recomendações</h3>
          </CardHeader>
          <CardContent>
            <ul className="flex flex-col gap-2">
              {recommendations.map((rec, i) => (
                <li key={i} className="flex items-start gap-2 text-sm text-muted-foreground">
                  <MessageSquareText className="mt-0.5 h-3.5 w-3.5 shrink-0 text-accent" />
                  {rec}
                </li>
              ))}
            </ul>
          </CardContent>
        </Card>
      </div>
    </div>
  )
}
