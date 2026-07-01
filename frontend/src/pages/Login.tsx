import { useState, type FormEvent } from 'react'
import { useNavigate, useLocation, Link } from 'react-router-dom'
import { Button } from '@/components/ui/Button'
import { Input } from '@/components/ui/Input'
import { Card } from '@/components/ui/Card'
import { Trophy, Mail, Lock, Eye, EyeOff } from 'lucide-react'
import { useAuth } from '@/contexts/AuthContext'

type LocationState = { from?: { pathname?: string }; notice?: string } | null

export function Login() {
  const { signInWithPassword, signUp } = useAuth()
  const navigate = useNavigate()
  const location = useLocation()
  const initialNotice = (location.state as LocationState)?.notice ?? null
  const [showPassword, setShowPassword] = useState(false)
  const [isRegister, setIsRegister] = useState(false)
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [error, setError] = useState<string | null>(null)
  const [notice, setNotice] = useState<string | null>(initialNotice)
  const [submitting, setSubmitting] = useState(false)

  async function onSubmit(e: FormEvent) {
    e.preventDefault()
    setError(null)
    setNotice(null)
    setSubmitting(true)
    try {
      if (isRegister) {
        await signUp(email, password)
        setNotice('Check your inbox to confirm your email, then sign in.')
        setIsRegister(false)
      } else {
        await signInWithPassword(email, password)
        const from = (location.state as LocationState)?.from?.pathname
        navigate(from && from !== '/login' ? from : '/tournaments', { replace: true })
      }
    } catch (err: any) {
      setError(err?.message ?? 'Authentication failed')
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <div className="min-h-[calc(100vh-6rem)] flex items-center justify-center">
      <div className="w-full max-w-md">
        <div className="text-center mb-8">
          <div className="w-12 h-12 rounded-xl bg-primary flex items-center justify-center mx-auto mb-4">
            <Trophy className="w-6 h-6 text-white" />
          </div>
          <h1 className="text-headline-md text-on-surface mb-2">
            {isRegister ? 'Create your account' : 'Welcome back'}
          </h1>
          <p className="text-body-md text-text-secondary">
            {isRegister
              ? 'Start organizing or competing in tournaments'
              : 'Sign in to manage your tournaments'}
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

            <div className="relative">
              <Input
                label="Password"
                type={showPassword ? 'text' : 'password'}
                required
                minLength={6}
                autoComplete={isRegister ? 'new-password' : 'current-password'}
                placeholder="Enter your password"
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

            {!isRegister && (
              <div className="flex justify-end">
                <Link to="/forgot-password" className="text-body-sm text-primary hover:text-primary-600">
                  Forgot password?
                </Link>
              </div>
            )}

            {notice && (
              <p className="text-body-sm text-success" role="status">{notice}</p>
            )}

            {error && (
              <p className="text-body-sm text-danger" role="alert">{error}</p>
            )}

            <Button type="submit" className="w-full" size="lg" loading={submitting} disabled={submitting}>
              {isRegister ? 'Create account' : 'Sign in'}
            </Button>
          </form>
        </Card>

        <p className="text-center mt-6 text-body-sm text-text-secondary">
          {isRegister ? 'Already have an account?' : "Don't have an account?"}{' '}
          <button
            onClick={() => { setIsRegister(!isRegister); setError(null); setNotice(null) }}
            className="text-primary font-medium hover:text-primary-600"
          >
            {isRegister ? 'Sign in' : 'Create one'}
          </button>
        </p>
      </div>
    </div>
  )
}
