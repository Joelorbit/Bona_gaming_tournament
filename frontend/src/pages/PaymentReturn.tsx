import { useEffect, useMemo, useState } from 'react'
import { Link, useSearchParams } from 'react-router-dom'
import { CheckCircle2, CircleAlert, Loader2 } from 'lucide-react'
import { Button } from '@/components/ui/Button'
import { Card } from '@/components/ui/Card'
import { api, ApiError } from '@/lib/api'

interface PaymentReturnResult {
  payment: {
    id: string
    status: string
  }
  tournament_id: string
  payment_status: string
}

export function PaymentReturn() {
  const [params] = useSearchParams()
  const [loading, setLoading] = useState(true)
  const [result, setResult] = useState<PaymentReturnResult | null>(null)
  const [error, setError] = useState<string | null>(null)

  const paymentID = params.get('payment_id') || ''
  const tournamentID = params.get('tournament_id') || ''
  const status = params.get('status') || ''
  const token = params.get('token') || ''

  const returnState = useMemo(() => {
    const normalized = status.toLowerCase()
    if (normalized === 'success' || normalized === 'paid' || normalized === 'completed') return 'success'
    if (normalized === 'cancelled' || normalized === 'canceled') return 'cancelled'
    return 'failed'
  }, [status])

  useEffect(() => {
    let cancelled = false
    async function confirmReturn() {
      if (!paymentID || !token) {
        setError('Payment return is missing required details.')
        setLoading(false)
        return
      }
      try {
        const data = await api.post<PaymentReturnResult>('/api/v1/payments/return', {
          payment_id: paymentID,
          status,
          token,
        })
        if (!cancelled) setResult(data)
      } catch (err) {
        if (!cancelled) setError(err instanceof ApiError ? err.message : 'Could not confirm payment.')
      } finally {
        if (!cancelled) setLoading(false)
      }
    }
    confirmReturn()
    return () => { cancelled = true }
  }, [paymentID, status, token])

  const finalTournamentID = result?.tournament_id || tournamentID
  const isPaid = result?.payment_status === 'paid'
  const isPending = result?.payment_status === 'pending'
  const isSuccessReturn = returnState === 'success'

  return (
    <div className="max-w-xl mx-auto py-10">
      <Card padding="lg">
        <div className="flex flex-col items-center text-center gap-4">
          <div className={`w-14 h-14 rounded-full flex items-center justify-center ${loading ? 'bg-primary/10 text-primary' : isPaid ? 'bg-success-200 text-success-800' : 'bg-danger/10 text-danger'}`}>
            {loading ? <Loader2 className="w-7 h-7 animate-spin" /> : isPaid ? <CheckCircle2 className="w-7 h-7" /> : <CircleAlert className="w-7 h-7" />}
          </div>

          <div>
            <h1 className="text-headline-sm text-on-surface">
              {loading ? 'Checking payment' : isPaid ? 'Payment confirmed' : isPending && isSuccessReturn ? 'Payment processing' : isSuccessReturn ? 'Payment needs review' : 'Payment not completed'}
            </h1>
            <p className="text-body-md text-text-secondary mt-2">
              {loading
                ? 'Checking your AddisPay return and current payment status.'
                : isPaid
                  ? 'Your tournament spot is secured. Continue to the tournament.'
                  : isPending && isSuccessReturn
                    ? 'We received your return from AddisPay. Your spot will be confirmed after the verified payment webhook arrives.'
                    : error || 'The payment was not confirmed. Return to the tournament and try again.'}
            </p>
          </div>

          <div className="flex flex-col sm:flex-row gap-2 w-full sm:w-auto">
            {finalTournamentID && (
              <Link to={`/tournaments/${finalTournamentID}`} className="w-full sm:w-auto">
                <Button className="w-full">
                  {isPaid ? 'Continue to tournament' : 'Back to tournament'}
                </Button>
              </Link>
            )}
            <Link to="/me/dashboard" className="w-full sm:w-auto">
              <Button variant="secondary" className="w-full">My dashboard</Button>
            </Link>
          </div>
        </div>
      </Card>
    </div>
  )
}
