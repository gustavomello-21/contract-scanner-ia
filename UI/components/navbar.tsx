"use client"

import Link from "next/link"
import { useUser, useClerk } from "@clerk/nextjs"
import { Button } from "@/components/ui/button"
import { ShieldCheck, LogOut } from "lucide-react"

export function Navbar() {
  const { isSignedIn, user } = useUser()
  const { signOut } = useClerk()

  return (
    <nav className="flex items-center justify-between py-5">
      <div className="flex items-center gap-2">
        <ShieldCheck className="h-6 w-6 text-accent" />
        <span className="text-lg font-semibold tracking-tight text-foreground">
          RiskLens
        </span>
      </div>

      {isSignedIn ? (
        <div className="flex items-center gap-3">
          <span className="text-sm text-muted-foreground">
            {user?.firstName || user?.emailAddresses[0]?.emailAddress}
          </span>
          <Button
            variant="outline"
            size="sm"
            className="text-sm"
            onClick={() => signOut({ redirectUrl: "/login" })}
          >
            <LogOut className="mr-1 h-4 w-4" />
            Sair
          </Button>
        </div>
      ) : (
        <Link href="/login">
          <Button variant="outline" size="sm" className="text-sm">
            Entrar
          </Button>
        </Link>
      )}
    </nav>
  )
}
