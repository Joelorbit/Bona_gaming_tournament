import { useState, type FormEvent } from 'react'
import { Link } from 'react-router-dom'
import { Button } from '@/components/ui/Button'
import { Input } from '@/components/ui/Input'
import { Card } from '@/components/ui/Card'
import { Trophy, Mail, ArrowLeft } from 'lucide-react'
import { supabase } from '@/lib/supabase'

export function ForgotPassword() {
  const [email, setEmail] = useState('')
  const [error, setError] = useState<string | null>(null)
  const [success, setSuccess] = useState(false)
  const [submitting, setSubmitting] = useState(false)

  async function onSubmit(e: FormEvent) {
    e.preventDefault()
    setError(null)
    setSubmitting(true)
    try {
      const { error } = await supabase.auth.resetPasswordForEmail(email, {
        redirectTo: `${window.location.origin}/reset-password`,
      })
      if (error) throw error
      setSuccess(true)
    } catch (err: any) {
      setError(err?.message ?? 'Failed to send reset email')
    } finally {
      setSubmitting(false)
    }
  }

  if (success) {
    return (
      <div className="min-h-[calc(100vh-6rem)] flex items-center justify-center">
        <div className="w-full max-w-md text-center">
          <div className="w-12 h-12 rounded-xl bg-success flex items-center justify-center mx-auto mb-4">
            <Mail className="w-6 h-6 text-white" />
          </div>
          <h1 className="text-headline-md text-on-surface mb-2">Check your inbox</h1>
          <p className="text-body-md text-text-secondary mb-6">
            We sent a password reset link to <strong>{email}</strong>
          </p>
          <Link to="/login">
            <Button variant="outline">Back to sign in</Button>
          </Link>
        </div>
      </div>
    )
  }

  return (
    <div className="min-h-[calc(100vh-6rem)] flex items-center justify-center">
      <div className="w-full max-w-md">
        <div className="text-center mb-8">
          <div className="w-12 h-12 rounded-xl bg-primary flex items-center justify-center mx-auto mb-4">
            <Trophy className="w-6 h-6 text-white" />
          </div>
          <h1 className="text-headline-md text-on-surface mb-2">Reset your password</h1>
          <p className="text-body-md text-text-secondary">
            Enter your email and we'll send you a reset link
          </p>
        </div>

        <Card padding="lg">
          <form onSubmit={onSubmit} className="space-y-5">
            <Input
              label="Email"
              type="email"
              required
              autoComplete="email"
              placeholder="you@example.com"
              icon={<Mail className="w-4 h-4" />}
              value={email}
              onChange={e => setEmail(e.target.value)}
            />

            {error && (
              <p className="text-body-sm text-danger" role="alert">{error}</p>
            )}

            <Button type="submit" className="w-full" size="lg" loading={submitting} disabled={submitting}>
              Send reset link
            </Button>
          </form>
        </Card>

        <Link to="/login" className="flex items-center justify-center gap-2 mt-6 text-body-sm text-text-secondary hover:text-on-surface">
          <ArrowLeft className="w-4 h-4" />
          Back to sign in
        </Link>
      </div>
    </div>
  )
}
