import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { Card, CardHeader, CardTitle, CardContent } from '@/components/ui/Card'
import { Badge } from '@/components/ui/Badge'
import { EmptyState } from '@/components/ui/EmptyState'
import { AlertTriangle, Clock, DollarSign, Landmark, Shield, Smartphone, Trophy, Users } from 'lucide-react'
import { api, ApiError, type Tournament } from '@/lib/api'
import { useAuth } from '@/contexts/AuthContext'
import { formatDate, formatTimeAgo } from '@/lib/utils'

interface PlatformStats {
  total_users: number
  total_tournaments: number
  active_tournaments: number
  total_registrations: number
  paid_registrations: number
  total_payments_paid: number
  pending_disputes: number
  pending_payouts: number
  revenue_etb: number
}

interface AuditEntry {
  id: string
  actor_id?: string | null
  actor_role?: string | null
  action: string
  entity_type: string
  entity_id?: string | null
  details?: Record<string, any> | null
  created_at: string
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
  addispay_ref?: string | null
  payment_url?: string | null
  provider_status?: string | null
  provider_payment_id?: string | null
  verified_at?: string | null
  webhook_received_at?: string | null
  failure_reason?: string | null
  refund_status: string
  refund_reason?: string | null
  refund_requested_at?: string | null
  refunded_at?: string | null
  refunded_by?: string | null
  created_at: string
  updated_at: string
}

export function Admin() {
  const { user } = useAuth()
  const [profileRole, setProfileRole] = useState<string | null>(null)
  const [stats, setStats] = useState<PlatformStats | null>(null)
  const [tournaments, setTournaments] = useState<Tournament[]>([])
  const [payments, setPayments] = useState<PaymentRow[]>([])
  const [payouts, setPayouts] = useState<PayoutRow[]>([])
  const [audit, setAudit] = useState<AuditEntry[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    let cancelled = false
    async function load() {
      try {
        const profile = await api.get<{ role: string }>('/api/v1/users/me')
        if (cancelled) return
        setProfileRole(profile.role)
        if (profile.role !== 'admin') {
          setLoading(false)
          return
        }
        const [s, t, pay, p, a] = await Promise.all([
          api.get<PlatformStats>('/api/v1/admin/stats'),
          api.get<Tournament[]>('/api/v1/admin/tournaments').catch(() => []),
          api.get<PaymentRow[]>('/api/v1/admin/payments').catch(() => []),
          api.get<PayoutRow[]>('/api/v1/admin/payouts').catch(() => []),
          api.get<AuditEntry[]>('/api/v1/admin/audit').catch(() => []),
        ])
        if (cancelled) return
        setStats(s)
        setTournaments(t || [])
        setPayments(pay || [])
        setPayouts(p || [])
        setAudit(a || [])
      } catch (err: any) {
        if (cancelled) return
        setError(err instanceof ApiError ? err.message : 'Failed to load admin data')
      } finally {
        if (!cancelled) setLoading(false)
      }
    }
    load()
    return () => { cancelled = true }
  }, [user])

  if (loading) return <div className="p-12 text-center text-text-secondary">Loading…</div>

  if (profileRole !== 'admin') {
    return (
      <EmptyState
        icon={<Shield className="w-6 h-6" />}
        title="Admin access required"
        description="Your account does not have admin permissions."
      />
    )
  }

  if (error) {
    return <EmptyState title="Couldn't load admin data" description={error} />
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center gap-3">
        <Shield className="w-6 h-6 text-primary" />
        <h1 className="text-headline-md text-on-surface">Admin</h1>
      </div>

      {stats && (
        <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
          <StatCard icon={<Users />} label="Users" value={stats.total_users} />
          <StatCard icon={<Trophy />} label="Tournaments" value={stats.total_tournaments} />
          <StatCard icon={<Trophy />} label="Active" value={stats.active_tournaments} />
          <StatCard icon={<DollarSign />} label="Platform revenue" value={`${stats.revenue_etb.toLocaleString()} ETB`} />
          <StatCard icon={<Users />} label="Registrations" value={stats.total_registrations} />
          <StatCard icon={<Users />} label="Paid registrations" value={stats.paid_registrations} />
          <StatCard icon={<AlertTriangle />} label="Open disputes" value={stats.pending_disputes} alert={stats.pending_disputes > 0} />
          <StatCard icon={<DollarSign />} label="Pending payouts" value={stats.pending_payouts} />
        </div>
      )}

      <Card padding="lg">
        <CardHeader><CardTitle>Tournaments ({tournaments.length})</CardTitle></CardHeader>
        <CardContent>
          {tournaments.length === 0 ? (
            <EmptyState title="No tournaments" />
          ) : (
            <div className="space-y-2">
              {tournaments.slice(0, 20).map(t => (
                <Link key={t.id} to={`/tournaments/${t.id}`} className="block p-3 border border-border rounded-lg hover:border-primary/50 transition-all">
                  <div className="flex items-center justify-between gap-3">
                    <div className="flex-1 min-w-0">
                      <p className="text-body-sm font-medium text-on-surface truncate">{t.title}</p>
                      <p className="text-label-md text-text-secondary">{t.game} · Organizer {t.organizer_id.slice(0, 8)}…</p>
                    </div>
                    <Badge status={t.status}>{t.status.replace(/_/g, ' ')}</Badge>
                  </div>
                </Link>
              ))}
            </div>
          )}
        </CardContent>
      </Card>

      <Card padding="lg">
        <CardHeader><CardTitle>Payments ({payments.length})</CardTitle></CardHeader>
        <CardContent>
          {payments.length === 0 ? (
            <EmptyState title="No payments" />
          ) : (
            <div className="space-y-2">
              {payments.slice(0, 50).map(payment => (
                <div key={payment.id} className="rounded-lg border border-border p-3">
                  <div className="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
                    <div className="min-w-0">
                      <p className="text-body-sm font-medium text-on-surface">
                        <Link to={`/tournaments/${payment.tournament_id}`} className="hover:text-primary">
                          Payment {payment.id.slice(0, 8)}...
                        </Link>
                      </p>
                      <p className="text-label-md text-text-secondary">
                        Player {payment.user_id.slice(0, 8)}... · Created {formatDate(payment.created_at)}
                      </p>
                    </div>
                    <div className="flex items-center justify-between gap-3 sm:block sm:text-right">
                      <p className="text-body-md font-bold text-on-surface">{payment.amount.toLocaleString()} {payment.currency}</p>
                      <Badge status={payment.status === 'paid' ? 'completed' : payment.status === 'failed' ? 'cancelled' : 'pending'}>
                        {payment.status}
                      </Badge>
                    </div>
                  </div>
                  <AdminPaymentDetails payment={payment} />
                </div>
              ))}
            </div>
          )}
        </CardContent>
      </Card>

      <Card padding="lg">
        <CardHeader><CardTitle>Payouts ({payouts.length})</CardTitle></CardHeader>
        <CardContent>
          {payouts.length === 0 ? (
            <EmptyState title="No payouts" />
          ) : (
            <div className="space-y-2">
              {payouts.slice(0, 50).map(p => (
                <div key={p.id} className="rounded-lg border border-border p-3">
                  <div className="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
                    <div className="min-w-0">
                      <p className="text-body-sm font-medium text-on-surface">
                        <Link to={`/tournaments/${p.tournament_id}`} className="hover:text-primary">
                          Tournament {p.tournament_id.slice(0, 8)}...
                        </Link>
                      </p>
                      <p className="text-label-md text-text-secondary">
                        Winner {p.winner_id.slice(0, 8)}... · Created {formatDate(p.created_at)}
                      </p>
                    </div>
                    <div className="flex items-center justify-between gap-3 sm:block sm:text-right">
                      <p className="text-body-md font-bold text-on-surface">{p.amount.toLocaleString()} {p.currency}</p>
                      <Badge status={p.status === 'paid' ? 'completed' : 'pending'}>
                        {p.status === 'paid' && p.paid_at ? `paid · ${formatDate(p.paid_at)}` : p.status}
                      </Badge>
                    </div>
                  </div>
                  <AdminPayoutDetails payout={p} />
                </div>
              ))}
            </div>
          )}
        </CardContent>
      </Card>

      <Card padding="lg">
        <CardHeader><CardTitle>Audit log</CardTitle></CardHeader>
        <CardContent>
          {audit.length === 0 ? (
            <EmptyState title="No audit entries yet" />
          ) : (
            <div className="space-y-1">
              {audit.map(e => (
                <div key={e.id} className="flex items-center gap-3 p-2 text-body-sm border-b border-border last:border-0">
                  <span className="text-label-md text-text-secondary w-24 shrink-0">{formatTimeAgo(e.created_at)}</span>
                  <span className="font-mono text-label-md text-text-secondary">{e.actor_role || '—'}</span>
                  <span className="font-medium text-on-surface">{e.action}</span>
                  <span className="text-text-secondary truncate">{e.entity_type}/{e.entity_id?.slice(0, 8)}</span>
                  {e.details && (
                    <span className="text-label-md text-text-secondary truncate ml-auto">
                      {JSON.stringify(e.details)}
                    </span>
                  )}
                </div>
              ))}
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  )
}

function AdminPaymentDetails({ payment }: { payment: PaymentRow }) {
  const hasWebhook = !!payment.webhook_received_at
  const isStalePending = payment.status === 'pending' && Date.now() - new Date(payment.created_at).getTime() > 30 * 60 * 1000

  return (
    <div className="mt-3 rounded-lg bg-surface-container-low p-3">
      <div className="mb-2 flex flex-wrap items-center gap-2">
        {!hasWebhook && <Badge status="pending">no webhook yet</Badge>}
        {payment.verified_at && <Badge status="completed">verified</Badge>}
        {payment.refund_status === 'pending' && <Badge status="pending">refund pending</Badge>}
        {payment.refund_status === 'refunded' && <Badge status="completed">refunded</Badge>}
        {isStalePending && (
          <span className="inline-flex items-center gap-1 text-label-md font-medium text-warning-700">
            <Clock className="h-4 w-4" />
            pending over 30m
          </span>
        )}
      </div>
      <div className="grid gap-2 text-body-sm sm:grid-cols-2 lg:grid-cols-3">
        {payment.addispay_ref && <Detail label="AddisPay ref" value={payment.addispay_ref} />}
        {payment.provider_status && <Detail label="Provider status" value={payment.provider_status} />}
        {payment.provider_payment_id && <Detail label="Provider payment ID" value={payment.provider_payment_id} />}
        {payment.webhook_received_at && <Detail label="Webhook received" value={formatDate(payment.webhook_received_at)} />}
        {payment.verified_at && <Detail label="Verified" value={formatDate(payment.verified_at)} />}
        {payment.failure_reason && <Detail label="Failure reason" value={payment.failure_reason} />}
        {payment.refund_reason && <Detail label="Refund reason" value={payment.refund_reason} />}
        {payment.refund_requested_at && <Detail label="Refund requested" value={formatDate(payment.refund_requested_at)} />}
        {payment.refunded_at && <Detail label="Refunded" value={formatDate(payment.refunded_at)} />}
      </div>
    </div>
  )
}

function AdminPayoutDetails({ payout }: { payout: PayoutRow }) {
  if (!payout.payout_details_submitted_at) {
    return (
      <div className="mt-3 rounded-lg bg-surface-container-low px-3 py-2 text-body-sm text-text-secondary">
        Payout details not submitted yet.
      </div>
    )
  }

  return (
    <div className="mt-3 rounded-lg bg-surface-container-low p-3">
      <div className="mb-2 flex items-center gap-2 text-body-sm font-medium text-on-surface">
        {payout.payout_method === 'bank' ? <Landmark className="h-4 w-4" /> : <Smartphone className="h-4 w-4" />}
        {payout.payout_method === 'bank' ? 'Bank payout' : 'Telebirr payout'}
      </div>
      <p className="text-body-sm text-text-secondary">Sensitive payout details are visible only to the hosting organizer.</p>
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
