import { SignUp } from "@clerk/nextjs"

export default function SignupPage() {
  return (
    <main className="flex min-h-svh items-center justify-center bg-background">
      <SignUp fallbackRedirectUrl="/" />
    </main>
  )
}
