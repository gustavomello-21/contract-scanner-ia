export type RiskLevel = "low" | "medium" | "high"

export interface KeyClause {
  clause: string
  description: string
  risk_level: RiskLevel
}

export interface AnalysisResult {
  summary: string
  parties: string[]
  contract_type: string
  key_clauses: KeyClause[]
  risks: string[]
  recommendations: string[]
  overall_risk: RiskLevel
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

async function presignUpload(file: File, token: string): Promise<PresignResponse> {
  const res = await fetch("/api/uploads/presign", {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      "Authorization": `Bearer ${token}`,
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

async function processAnalysis(analysisId: string, token: string): Promise<ProcessResponse> {
  const res = await fetch(`/api/analyses/${analysisId}/process`, {
    method: "POST",
    headers: {
      "Authorization": `Bearer ${token}`,
    },
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
  token: string,
  onStepChange?: (step: AnalysisStep) => void
): Promise<AnalysisResult> {
  onStepChange?.("presigning")
  const { analysis_id, upload_url } = await presignUpload(file, token)

  onStepChange?.("uploading")
  await uploadToS3(upload_url, file)

  onStepChange?.("processing")
  const { result } = await processAnalysis(analysis_id, token)

  return result
}
