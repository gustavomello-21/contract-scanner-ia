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
  CheckSquare,
  DollarSign,
  Clock,
  AlertCircle,
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

function SectionCard({ icon, title, children }: { icon: React.ReactNode; title: string; children: React.ReactNode }) {
  return (
    <Card className="border-border bg-card">
      <CardHeader className="flex flex-row items-center gap-3 pb-3">
        <div className="flex h-8 w-8 items-center justify-center rounded-md bg-secondary">
          {icon}
        </div>
        <h3 className="text-sm font-medium text-foreground">{title}</h3>
      </CardHeader>
      <CardContent>{children}</CardContent>
    </Card>
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

  const {
    summary,
    parties,
    contract_type,
    key_clauses,
    risks,
    recommendations,
    overall_risk,
    term_and_termination,
    financials,
    obligations_of_signatory,
    ambiguous_terms,
    quick_checks,
    analysis_warnings,
  } = result

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

      {/* Avisos de extração */}
      {analysis_warnings && analysis_warnings.length > 0 && (
        <Card className="border-yellow-500/50 bg-yellow-500/5">
          <CardContent className="flex flex-col gap-1 pt-4">
            {analysis_warnings.map((w, i) => (
              <p key={i} className="flex items-start gap-2 text-xs text-yellow-700 dark:text-yellow-400">
                <AlertCircle className="mt-0.5 h-3.5 w-3.5 shrink-0" />
                {w}
              </p>
            ))}
          </CardContent>
        </Card>
      )}

      {/* Resumo */}
      <SectionCard icon={<FileText className="h-4 w-4 text-muted-foreground" />} title="Resumo">
        <p className="text-sm leading-relaxed text-muted-foreground">{summary}</p>
      </SectionCard>

      {/* Partes + Tipo de Contrato */}
      <div className="grid gap-4 sm:grid-cols-2">
        <SectionCard icon={<Users className="h-4 w-4 text-muted-foreground" />} title="Partes">
          <ul className="flex flex-col gap-2">
            {parties.map((party, i) => (
              <li key={i} className="text-sm">
                <span className="font-medium text-foreground">{party.name}</span>
                <span className="ml-1 text-muted-foreground">— {party.role}</span>
                {party.identifier && (
                  <span className="block text-xs text-muted-foreground">{party.identifier}</span>
                )}
              </li>
            ))}
          </ul>
        </SectionCard>

        <SectionCard icon={<Scale className="h-4 w-4 text-muted-foreground" />} title="Tipo de Contrato">
          <p className="text-sm text-muted-foreground">{contract_type ?? "Não identificado"}</p>
          {term_and_termination.effective_date && (
            <p className="mt-2 text-xs text-muted-foreground">
              <span className="font-medium">Vigência:</span> {term_and_termination.effective_date}
              {term_and_termination.expiry_or_term && ` · ${term_and_termination.expiry_or_term}`}
            </p>
          )}
          {term_and_termination.renewal && (
            <p className="mt-1 text-xs text-muted-foreground">
              <span className="font-medium">Renovação:</span> {term_and_termination.renewal}
            </p>
          )}
        </SectionCard>
      </div>

      {/* Financeiro */}
      {(financials.payment_terms || financials.amounts.length > 0) && (
        <SectionCard icon={<DollarSign className="h-4 w-4 text-muted-foreground" />} title="Financeiro">
          {financials.payment_terms && (
            <p className="text-sm text-muted-foreground">
              <span className="font-medium text-foreground">Pagamento:</span> {financials.payment_terms}
            </p>
          )}
          {financials.amounts.length > 0 && (
            <ul className="mt-2 flex flex-col gap-1">
              {financials.amounts.map((a, i) => (
                <li key={i} className="text-sm text-muted-foreground">· {a}</li>
              ))}
            </ul>
          )}
          {financials.penalties_and_interest && (
            <p className="mt-2 text-xs text-muted-foreground">
              <span className="font-medium">Multas/Juros:</span> {financials.penalties_and_interest}
            </p>
          )}
        </SectionCard>
      )}

      {/* Cláusulas Críticas */}
      <SectionCard icon={<AlertTriangle className="h-4 w-4 text-muted-foreground" />} title="Cláusulas Críticas">
        {key_clauses.length === 0 ? (
          <p className="text-sm text-muted-foreground">Nenhuma cláusula identificada.</p>
        ) : (
          <Accordion type="multiple" className="w-full">
            {key_clauses.map((clause, i) => (
              <AccordionItem key={i} value={`clause-${i}`}>
                <AccordionTrigger>
                  <div className="flex items-center gap-2 text-left">
                    <span>{clause.clause_name}</span>
                    <RiskBadge level={clause.risk_level} />
                  </div>
                </AccordionTrigger>
                <AccordionContent className="flex flex-col gap-2">
                  <p className="text-sm text-muted-foreground">{clause.clause_summary}</p>
                  {clause.clause_text_snippet && (
                    <blockquote className="border-l-2 border-border pl-3 text-xs italic text-muted-foreground">
                      "{clause.clause_text_snippet}"
                    </blockquote>
                  )}
                  {clause.risk_explanation && (
                    <p className="text-xs text-muted-foreground">
                      <span className="font-medium">Risco:</span> {clause.risk_explanation}
                    </p>
                  )}
                  {clause.recommended_fix && (
                    <p className="text-xs text-muted-foreground">
                      <span className="font-medium">Sugestão:</span> {clause.recommended_fix}
                    </p>
                  )}
                  {clause.clause_location && (
                    <p className="text-xs text-muted-foreground opacity-60">{clause.clause_location}</p>
                  )}
                </AccordionContent>
              </AccordionItem>
            ))}
          </Accordion>
        )}
      </SectionCard>

      {/* Riscos + Recomendações */}
      <div className="grid gap-4 sm:grid-cols-2">
        <SectionCard icon={<ShieldAlert className="h-4 w-4 text-muted-foreground" />} title="Riscos Identificados">
          <ul className="flex flex-col gap-3">
            {risks.map((risk, i) => (
              <li key={i} className="flex flex-col gap-1 text-sm">
                <div className="flex items-start gap-2">
                  <AlertTriangle className="mt-0.5 h-3.5 w-3.5 shrink-0 text-destructive" />
                  <span className="text-muted-foreground">{risk.risk}</span>
                  <RiskBadge level={risk.impact} />
                </div>
                {risk.mitigation && (
                  <p className="ml-5 text-xs text-muted-foreground opacity-75">{risk.mitigation}</p>
                )}
              </li>
            ))}
          </ul>
        </SectionCard>

        <SectionCard icon={<Lightbulb className="h-4 w-4 text-muted-foreground" />} title="Recomendações">
          <ul className="flex flex-col gap-3">
            {recommendations.map((rec, i) => (
              <li key={i} className="flex flex-col gap-1 text-sm">
                <div className="flex items-start gap-2">
                  <MessageSquareText className="mt-0.5 h-3.5 w-3.5 shrink-0 text-accent" />
                  <span className="text-muted-foreground">{rec.action}</span>
                  <RiskBadge level={rec.priority} />
                </div>
                {rec.proposed_text && (
                  <p className="ml-5 text-xs italic text-muted-foreground opacity-75">"{rec.proposed_text}"</p>
                )}
              </li>
            ))}
          </ul>
        </SectionCard>
      </div>

      {/* Obrigações */}
      {obligations_of_signatory.length > 0 && (
        <SectionCard icon={<Clock className="h-4 w-4 text-muted-foreground" />} title="Obrigações do Contratante">
          <ul className="flex flex-col gap-2">
            {obligations_of_signatory.map((ob, i) => (
              <li key={i} className="text-sm">
                <span className="text-foreground">{ob.obligation}</span>
                {ob.deadline_or_frequency && (
                  <span className="ml-1 text-xs text-muted-foreground">· {ob.deadline_or_frequency}</span>
                )}
              </li>
            ))}
          </ul>
        </SectionCard>
      )}

      {/* Termos ambíguos */}
      {ambiguous_terms.length > 0 && (
        <SectionCard icon={<AlertCircle className="h-4 w-4 text-muted-foreground" />} title="Termos Ambíguos">
          <ul className="flex flex-col gap-3">
            {ambiguous_terms.map((t, i) => (
              <li key={i} className="text-sm">
                <span className="font-medium text-foreground">"{t.term}"</span>
                <p className="mt-0.5 text-xs text-muted-foreground">{t.why_problematic}</p>
                {t.suggested_clarification && (
                  <p className="mt-0.5 text-xs text-muted-foreground opacity-75">
                    <span className="font-medium">Sugestão:</span> {t.suggested_clarification}
                  </p>
                )}
              </li>
            ))}
          </ul>
        </SectionCard>
      )}

      {/* Quick Checks */}
      {quick_checks.length > 0 && (
        <SectionCard icon={<CheckSquare className="h-4 w-4 text-muted-foreground" />} title="Checklist antes de assinar">
          <ul className="flex flex-col gap-2">
            {quick_checks.map((check, i) => (
              <li key={i} className="flex items-start gap-2 text-sm text-muted-foreground">
                <CheckSquare className="mt-0.5 h-3.5 w-3.5 shrink-0 text-accent" />
                {check}
              </li>
            ))}
          </ul>
        </SectionCard>
      )}
    </div>
  )
}
