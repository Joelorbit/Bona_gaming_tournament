import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { Card, CardHeader, CardTitle, CardContent } from '@/components/ui/Card'
import { Button } from '@/components/ui/Button'
import { Badge } from '@/components/ui/Badge'
import { EmptyState } from '@/components/ui/EmptyState'
import { AlertTriangle, CheckCircle2, DollarSign, Landmark, Plus, Smartphone, Trophy } from 'lucide-react'
import { api, type Tournament } from '@/lib/api'
import { formatDate } from '@/lib/utils'

interface MatchRow {
  id: string
  tournament_id: string
  round: number
  position: number
  status: string
  dispute_status: string
  dispute_reason?: string | null
  dispute_opened_at?: string | null
}

interface PayoutRow {
  id: string
  tournament_id: string
  winner_id: string
  amount: number
  currency: string
  status: string
  payout_method?: 'telebirr' | 'bank' | null
  phone_number?: string | null
  telebirr_number?: string | null
  bank_name?: string | null
  bank_account_name?: string | null
  bank_account_number?: string | null
  payout_details_submitted_at?: string | null
  paid_at?: string | null
  created_at: string
}

interface PaymentRow {
  id: string
  user_id: string
  tournament_id: string
  amount: number
  currency: string
  status: string
  refund_status: string
  refund_reason?: string | null
  refund_requested_at?: string | null
  refunded_at?: string | null
  created_at: string
}

export function OrganizerDashboard() {
  const [tournaments, setTournaments] = useState<Tournament[]>([])
  const [disputes, setDisputes] = useState<MatchRow[]>([])
  const [payouts, setPayouts] = useState<PayoutRow[]>([])
  const [payments, setPayments] = useState<PaymentRow[]>([])
  const [tournamentTitles, setTournamentTitles] = useState<Record<string, string>>({})
  const [loading, setLoading] = useState(true)
  const [marking, setMarking] = useState<string | null>(null)
  const [refunding, setRefunding] = useState<string | null>(null)
  const [error, setError] = useState('')

  async function reload() {
    const [t, d, p, payments] = await Promise.all([
      api.get<Tournament[]>('/api/v1/tournaments/my').catch(() => []),
      api.get<MatchRow[]>('/api/v1/me/disputes').catch(() => []),
      api.get<PayoutRow[]>('/api/v1/me/organizer-payouts').catch(() => []),
      api.get<PaymentRow[]>('/api/v1/me/organizer-payments').catch(() => []),
    ])
    setTournaments(t || [])
    setDisputes(d || [])
    setPayouts(p || [])
    setPayments(payments || [])
    setTournamentTitles(Object.fromEntries((t || []).map(row => [row.id, row.title])))
  }

  useEffect(() => {
    let cancelled = false
    reload().finally(() => { if (!cancelled) setLoading(false) })
    return () => { cancelled = true }
  }, [])

  if (loading) return <div className="p-12 text-center text-text-secondary">Loading…</div>

  const active = tournaments.filter(t => t.status === 'open' || t.status === 'in_progress' || t.status === 'registration_closed')
  const pendingPayouts = payouts.filter(p => p.status === 'pending')
  const pendingRefunds = payments.filter(payment => payment.refund_status === 'pending')
  const totalPrizePool = tournaments.reduce((s, t) => s + (t.prize_pool || 0), 0)

  async function markPaid(id: string) {
    setMarking(id)
    setError('')
    try {
      await api.post(`/api/v1/payouts/${id}/mark-paid`)
      await reload()
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to mark payout paid')
    } finally {
      setMarking(null)
    }
  }

  async function markRefunded(id: string) {
    setRefunding(id)
    setError('')
    try {
      await api.post(`/api/v1/payments/${id}/mark-refunded`)
      await reload()
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to mark refund paid')
    } finally {
      setRefunding(null)
    }
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-headline-md text-on-surface">Organizer dashboard</h1>
          <p className="text-body-md text-text-secondary">Manage tournaments, disputes, and payouts.</p>
        </div>
        <Link to="/create">
          <Button icon={<Plus className="w-4 h-4" />}>Create tournament</Button>
        </Link>
      </div>

      {error && (
        <Card padding="md" className="border-danger/40 bg-danger/5">
          <p className="text-body-sm text-danger">{error}</p>
        </Card>
      )}

      <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
        <StatCard icon={<Trophy />} label="Tournaments" value={tournaments.length} />
        <StatCard icon={<Trophy />} label="Active" value={active.length} />
        <StatCard icon={<DollarSign />} label="Prize pool" value={`${totalPrizePool.toLocaleString()} ETB`} />
        <StatCard icon={<AlertTriangle />} label="Open disputes" value={disputes.length} alert={disputes.length > 0} />
      </div>

      {pendingRefunds.length > 0 && (
        <Card padding="lg" className="border-warning-500/40 bg-warning-50">
          <CardHeader><CardTitle>Refunds required</CardTitle></CardHeader>
          <CardContent>
            <div className="space-y-2">
              {pendingRefunds.map(payment => (
                <div key={payment.id} className="flex flex-col gap-3 rounded-lg border border-warning-200 bg-white p-3 sm:flex-row sm:items-center sm:justify-between">
                  <div className="min-w-0">
                    <p className="text-body-sm font-medium text-on-surface">
                      {tournamentTitles[payment.tournament_id] || 'Cancelled tournament'}
                    </p>
                    <p className="text-label-md text-text-secondary">
                      Player {payment.user_id.slice(0, 8)}... · {payment.amount.toLocaleString()} {payment.currency}
                      {payment.refund_requested_at ? ` · Requested ${formatDate(payment.refund_requested_at)}` : ''}
                    </p>
                    {payment.refund_reason && <p className="text-label-md text-text-secondary">{payment.refund_reason}</p>}
                  </div>
                  <Button
                    size="sm"
                    variant="outline"
                    loading={refunding === payment.id}
                    onClick={() => markRefunded(payment.id)}
                  >
                    Mark refunded
                  </Button>
                </div>
              ))}
            </div>
          </CardContent>
        </Card>
      )}

      {disputes.length > 0 && (
        <Card padding="lg" className="border-danger/40 bg-danger/5">
          <CardHeader><CardTitle>Pending disputes</CardTitle></CardHeader>
          <CardContent>
            <div className="space-y-2">
              {disputes.map(d => (
                <Link key={d.id} to={`/matches/${d.id}`} className="block p-3 border border-border bg-white rounded-lg hover:border-primary/50 transition-all">
                  <div className="flex items-center justify-between gap-3">
                    <div className="flex-1 min-w-0">
                      <p className="text-body-sm font-medium text-on-surface">Round {d.round} · Match {d.position + 1}</p>
                      {d.dispute_reason && <p className="text-label-md text-text-secondary truncate">{d.dispute_reason}</p>}
                    </div>
                    <Badge status="cancelled">disputed</Badge>
                  </div>
                </Link>
              ))}
            </div>
          </CardContent>
        </Card>
      )}

      <Card padding="lg">
        <CardHeader><CardTitle>My tournaments</CardTitle></CardHeader>
        <CardContent>
          {tournaments.length === 0 ? (
            <EmptyState
              title="No tournaments yet"
              description="Create your first tournament to get started."
              action={{ label: 'Create tournament', onClick: () => window.location.href = '/create' }}
            />
          ) : (
            <div className="space-y-2">
              {tournaments.map(t => (
                <Link key={t.id} to={`/tournaments/${t.id}`} className="block p-3 border border-border rounded-lg hover:border-primary/50 hover:bg-surface-container-low transition-all">
                  <div className="flex items-center justify-between gap-3">
                    <div className="flex-1 min-w-0">
                      <p className="text-body-sm font-medium text-on-surface truncate">{t.title}</p>
                      <p className="text-label-md text-text-secondary">
                        {t.game} · {t.format.replace(/_/g, ' ')} · starts {formatDate(t.start_date)}
                      </p>
                    </div>
                    <div className="flex items-center gap-2">
                      <span className="text-label-md text-text-secondary">{t.entry_fee > 0 ? `${t.entry_fee.toLocaleString()} ${t.currency}` : 'Free'}</span>
                      <Badge status={t.status}>{t.status.replace(/_/g, ' ')}</Badge>
                    </div>
                  </div>
                </Link>
              ))}
            </div>
          )}
        </CardContent>
      </Card>

      <Card padding="lg">
        <CardHeader><CardTitle>Payouts owed</CardTitle></CardHeader>
        <CardContent>
          {payouts.length === 0 ? (
            <EmptyState title="No payouts" description="Payouts appear here when your tournaments complete." />
          ) : (
            <div className="space-y-2">
              {payouts.map(p => (
                <div key={p.id} className="rounded-lg border border-border p-3">
                  <div className="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
                    <div className="flex-1 min-w-0">
                      <p className="text-body-sm font-medium text-on-surface truncate">
                        <Link to={`/tournaments/${p.tournament_id}`} className="hover:text-primary">{tournamentTitles[p.tournament_id] || 'Tournament'}</Link>
                      </p>
                      <p className="text-label-md text-text-secondary">Winner: {p.winner_id.slice(0, 8)}... · {formatDate(p.created_at)}</p>
                    </div>
                    <div className="flex items-center justify-between gap-3 sm:block sm:text-right">
                      <p className="text-body-md font-bold text-on-surface">{p.amount.toLocaleString()} {p.currency}</p>
                      {p.status === 'paid' ? (
                        <Badge status="completed">paid {p.paid_at && `· ${formatDate(p.paid_at)}`}</Badge>
                      ) : p.payout_details_submitted_at ? (
                        <Button
                          size="sm"
                          variant="outline"
                          loading={marking === p.id}
                          icon={<CheckCircle2 className="h-4 w-4" />}
                          onClick={() => markPaid(p.id)}
                        >
                          Mark paid
                        </Button>
                      ) : (
                        <Badge status="pending">awaiting details</Badge>
                      )}
                    </div>
                  </div>
                  <PayoutDetails payout={p} />
                </div>
              ))}
            </div>
          )}
          {pendingPayouts.length > 0 && (
            <p className="text-body-sm text-text-secondary mt-3">
              {pendingPayouts.length} payout{pendingPayouts.length === 1 ? '' : 's'} awaiting payment to winners.
            </p>
          )}
        </CardContent>
      </Card>
    </div>
  )
}

function PayoutDetails({ payout }: { payout: PayoutRow }) {
  if (!payout.payout_details_submitted_at) {
    return (
      <div className="mt-3 rounded-lg bg-surface-container-low px-3 py-2 text-body-sm text-text-secondary">
        Winner has not submitted payout details yet.
      </div>
    )
  }

  return (
    <div className="mt-3 rounded-lg bg-surface-container-low p-3">
      <div className="mb-2 flex items-center gap-2 text-body-sm font-medium text-on-surface">
        {payout.payout_method === 'bank' ? <Landmark className="h-4 w-4" /> : <Smartphone className="h-4 w-4" />}
        {payout.payout_method === 'bank' ? 'Bank payout' : 'Telebirr payout'}
      </div>
      <div className="grid gap-2 text-body-sm sm:grid-cols-2">
        {payout.phone_number && <Detail label="Phone" value={payout.phone_number} />}
        {payout.payout_method === 'telebirr' && payout.telebirr_number && <Detail label="Telebirr number" value={payout.telebirr_number} />}
        {payout.payout_method === 'bank' && payout.bank_name && <Detail label="Bank" value={payout.bank_name} />}
        {payout.payout_method === 'bank' && payout.bank_account_name && <Detail label="Account name" value={payout.bank_account_name} />}
        {payout.payout_method === 'bank' && payout.bank_account_number && <Detail label="Account number" value={payout.bank_account_number} />}
      </div>
      <p className="mt-2 text-label-md text-text-secondary">Submitted {formatDate(payout.payout_details_submitted_at)}</p>
    </div>
  )
}

function Detail({ label, value }: { label: string; value: string }) {
  return (
    <div>
      <p className="text-label-md text-text-secondary">{label}</p>
      <p className="break-words text-body-sm font-medium text-on-surface">{value}</p>
    </div>
  )
}

function StatCard({ icon, label, value, alert }: { icon: React.ReactNode; label: string; value: number | string; alert?: boolean }) {
  return (
    <Card padding="md">
      <div className="flex items-center gap-3">
        <div className={`w-10 h-10 rounded-lg flex items-center justify-center ${alert ? 'bg-danger/10 text-danger' : 'bg-primary/10 text-primary'}`}>
          {icon}
        </div>
        <div>
          <p className="text-label-md text-text-secondary">{label}</p>
          <p className="text-title-lg font-bold text-on-surface">{value}</p>
        </div>
      </div>
    </Card>
  )
}
