export type RiskLevel = "low" | "medium" | "high"

export interface Party {
  name: string
  role: string
  identifier: string | null
  address: string | null
}

export interface KeyClause {
  clause_name: string
  clause_summary: string
  clause_text_snippet: string
  clause_location: string
  risk_level: RiskLevel
  risk_explanation: string
  recommended_fix: string | null
}

export interface Risk {
  risk: string
  impact: RiskLevel
  clause_ref: string
  mitigation: string
}

export interface Recommendation {
  priority: RiskLevel
  action: string
  proposed_text: string | null
}

export interface AmbiguousTerm {
  term: string
  why_problematic: string
  suggested_clarification: string
}

export interface Obligation {
  obligation: string
  deadline_or_frequency: string
  clause_ref: string
}

export interface AnalysisResult {
  metadata: {
    language: string
    filename: string | null
    analysis_date: string
    confidence: number
  }
  summary: string
  parties: Party[]
  contract_type: string | null
  term_and_termination: {
    effective_date: string | null
    expiry_or_term: string | null
    renewal: string | null
    termination_rights: Array<{
      party: string
      notice_period: string
      cause: string
      clause_ref: string
    }>
  }
  financials: {
    payment_terms: string | null
    amounts: string[]
    penalties_and_interest: string | null
    security_or_guarantee: string | null
  }
  key_clauses: KeyClause[]
  obligations_of_signatory: Obligation[]
  ambiguous_terms: AmbiguousTerm[]
  indemnities_and_liabilities: {
    indemnity_summary: string | null
    liability_limit: string | null
    exclusions: string | null
  }
  confidentiality_and_ip: {
    confidentiality_summary: string | null
    ip_assignment_or_license: string | null
    risks: string[]
  }
  dispute_resolution: {
    governing_law: string | null
    forum_or_arbitration: string | null
    costs_allocation: string | null
  }
  risks: Risk[]
  recommendations: Recommendation[]
  quick_checks: string[]
  overall_risk: RiskLevel
  confidence?: number
  analysis_warnings?: string[]
}

interface PresignResponse {
  analysis_id: string
  upload_url: string
  s3_key: string
}

interface ProcessResponse {
  analysis_id: string
  status: string
  result: AnalysisResult
}

async function presignUpload(file: File): Promise<PresignResponse> {
  const res = await fetch("/api/uploads/presign", {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify({
      filename: file.name,
      content_type: file.type || "application/pdf",
      size_bytes: file.size,
    }),
  })

  if (!res.ok) {
    const body = await res.json().catch(() => ({}))
    throw new Error(body.error || "Falha ao preparar upload")
  }

  return res.json()
}

async function uploadToS3(uploadUrl: string, file: File): Promise<void> {
  const res = await fetch(uploadUrl, {
    method: "PUT",
    headers: { "Content-Type": file.type || "application/pdf" },
    body: file,
  })

  if (!res.ok) {
    throw new Error("Falha ao enviar arquivo para o S3")
  }
}

async function processAnalysis(analysisId: string): Promise<ProcessResponse> {
  const res = await fetch(`/api/analyses/${analysisId}/process`, {
    method: "POST",
  })

  if (!res.ok) {
    const body = await res.json().catch(() => ({}))
    throw new Error(body.error || "Falha ao processar contrato")
  }

  return res.json()
}

export type AnalysisStep = "presigning" | "uploading" | "processing"

export async function analyzeContract(
  file: File,
  onStepChange?: (step: AnalysisStep) => void
): Promise<AnalysisResult> {
  onStepChange?.("presigning")
  const { analysis_id, upload_url } = await presignUpload(file)

  onStepChange?.("uploading")
  await uploadToS3(upload_url, file)

  onStepChange?.("processing")
  const { result } = await processAnalysis(analysis_id)

  return result
}
