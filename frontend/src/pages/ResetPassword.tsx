import { useEffect, useState, type FormEvent } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { Button } from '@/components/ui/Button'
import { Input } from '@/components/ui/Input'
import { Card } from '@/components/ui/Card'
import { Trophy, Lock, Eye, EyeOff, AlertTriangle } from 'lucide-react'
import { supabase } from '@/lib/supabase'

function hashHasRecovery(): boolean {
  if (typeof window === 'undefined') return false
  const hash = window.location.hash.startsWith('#')
    ? window.location.hash.slice(1)
    : window.location.hash
  if (!hash) return false
  const params = new URLSearchParams(hash)
  return params.get('type') === 'recovery'
}

export function ResetPassword() {
  const navigate = useNavigate()
  const [showPassword, setShowPassword] = useState(false)
  const [password, setPassword] = useState('')
  const [confirmPassword, setConfirmPassword] = useState('')
  const [error, setError] = useState<string | null>(null)
  const [submitting, setSubmitting] = useState(false)
  const [ready, setReady] = useState<boolean>(hashHasRecovery)
  const [checking, setChecking] = useState(true)

  useEffect(() => {
    let cancelled = false

    const { data: sub } = supabase.auth.onAuthStateChange((event) => {
      if (event === 'PASSWORD_RECOVERY' && !cancelled) {
        setReady(true)
        setChecking(false)
      }
    })

    const timer = window.setTimeout(() => {
      if (cancelled) return
      setChecking(false)
    }, 600)

    return () => {
      cancelled = true
      window.clearTimeout(timer)
      sub.subscription.unsubscribe()
    }
  }, [])

  async function onSubmit(e: FormEvent) {
    e.preventDefault()
    setError(null)

    if (password !== confirmPassword) {
      setError('Passwords do not match')
      return
    }

    if (password.length < 6) {
      setError('Password must be at least 6 characters')
      return
    }

    setSubmitting(true)
    try {
      const { error } = await supabase.auth.updateUser({ password })
      if (error) throw error
      await supabase.auth.signOut()
      navigate('/login', {
        replace: true,
        state: { notice: 'Password updated. Sign in with your new password.' },
      })
    } catch (err: any) {
      setError(err?.message ?? 'Failed to reset password')
    } finally {
      setSubmitting(false)
    }
  }

  if (checking && !ready) {
    return (
      <div className="min-h-[calc(100vh-6rem)] flex items-center justify-center">
        <p className="text-body-md text-text-secondary">Verifying reset link...</p>
      </div>
    )
  }

  if (!ready) {
    return (
      <div className="min-h-[calc(100vh-6rem)] flex items-center justify-center">
        <div className="w-full max-w-md text-center">
          <div className="w-12 h-12 rounded-xl bg-danger flex items-center justify-center mx-auto mb-4">
            <AlertTriangle className="w-6 h-6 text-white" />
          </div>
          <h1 className="text-headline-md text-on-surface mb-2">Reset link invalid</h1>
          <p className="text-body-md text-text-secondary mb-6">
            This password reset link is missing, expired, or has already been used.
          </p>
          <Link to="/forgot-password">
            <Button variant="outline">Request a new link</Button>
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
          <h1 className="text-headline-md text-on-surface mb-2">Set new password</h1>
          <p className="text-body-md text-text-secondary">
            Choose a strong password for your account
          </p>
        </div>

        <Card padding="lg">
          <form onSubmit={onSubmit} className="space-y-5">
            <div className="relative">
              <Input
                label="New password"
                type={showPassword ? 'text' : 'password'}
                required
                minLength={6}
                autoComplete="new-password"
                placeholder="Enter new password"
                icon={<Lock className="w-4 h-4" />}
                value={password}
                onChange={e => setPassword(e.target.value)}
              />
              <button
                type="button"
                onClick={() => setShowPassword(!showPassword)}
                className="absolute right-3 top-[42px] text-text-secondary hover:text-on-surface"
              >
                {showPassword ? <EyeOff className="w-4 h-4" /> : <Eye className="w-4 h-4" />}
              </button>
            </div>

            <Input
              label="Confirm password"
              type={showPassword ? 'text' : 'password'}
              required
              minLength={6}
              autoComplete="new-password"
              placeholder="Confirm new password"
              icon={<Lock className="w-4 h-4" />}
              value={confirmPassword}
              onChange={e => setConfirmPassword(e.target.value)}
            />

            {error && (
              <p className="text-body-sm text-danger" role="alert">{error}</p>
            )}

            <Button type="submit" className="w-full" size="lg" loading={submitting} disabled={submitting}>
              Reset password
            </Button>
          </form>
        </Card>
      </div>
    </div>
  )
}
