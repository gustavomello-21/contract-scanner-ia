import { ShieldCheck } from "lucide-react"

export function RiskLensLogo() {
  return (
    <div className="flex items-center justify-center gap-2">
      <ShieldCheck className="size-8 text-primary" />
      <span className="text-2xl font-bold tracking-tight text-foreground">
        Risk<span className="text-primary">Lens</span>
      </span>
    </div>
  )
}
