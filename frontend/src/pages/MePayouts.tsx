import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { ArrowLeft, Landmark, MoreHorizontal, Receipt, Save, Smartphone, Trophy, Wallet } from 'lucide-react'
import { Badge } from '@/components/ui/Badge'
import { Button } from '@/components/ui/Button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/Card'
import { EmptyState } from '@/components/ui/EmptyState'
import { Input } from '@/components/ui/Input'
import { api, type Tournament } from '@/lib/api'
import { formatDate } from '@/lib/utils'

type PayoutMethod = 'telebirr' | 'bank' | 'other'

interface PayoutRow {
  id: string
  tournament_id: string
  winner_id: string
  amount: number
  currency: string
  status: string
  payout_method?: PayoutMethod | null
  phone_number?: string | null
  telebirr_number?: string | null
  bank_name?: string | null
  bank_account_name?: string | null
  bank_account_number?: string | null
  payout_instructions?: string | null
  payout_details_submitted_at?: string | null
  note?: string | null
  paid_at?: string | null
  created_at: string
}

interface PayoutDetailsForm {
  payout_method: PayoutMethod
  phone_number: string
  telebirr_number: string
  bank_name: string
  bank_account_name: string
  bank_account_number: string
  payout_instructions: string
}

export function MePayouts() {
  const [payouts, setPayouts] = useState<PayoutRow[]>([])
  const [tournamentTitles, setTournamentTitles] = useState<Record<string, string>>({})
  const [forms, setForms] = useState<Record<string, PayoutDetailsForm>>({})
  const [saving, setSaving] = useState<string | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  useEffect(() => {
    let cancelled = false

    async function load() {
      setError('')
      try {
        const rows = await api.get<PayoutRow[]>('/api/v1/me/payouts')
        if (cancelled) return

        const nextPayouts = rows || []
        setPayouts(nextPayouts)
        setForms(Object.fromEntries(nextPayouts.map(payout => [payout.id, formFromPayout(payout)])))

        const tournamentIDs = Array.from(new Set(nextPayouts.map(row => row.tournament_id)))
        const titles = await Promise.all(
          tournamentIDs.map(id =>
            api.get<Tournament>(`/api/v1/tournaments/${id}`)
              .then(tournament => [id, tournament.title] as const)
              .catch(() => [id, 'Tournament'] as const)
          )
        )

        if (!cancelled) setTournamentTitles(Object.fromEntries(titles))
      } catch (err) {
        if (!cancelled) setError(err instanceof Error ? err.message : 'Failed to load payouts')
      } finally {
        if (!cancelled) setLoading(false)
      }
    }

    load()
    return () => { cancelled = true }
  }, [])

  if (loading) return <div className="p-12 text-center text-text-secondary">Loading...</div>

  const paidPayouts = payouts.filter(payout => payout.status === 'paid')
  const pendingPayouts = payouts.filter(payout => payout.status === 'pending')
  const totalPaid = paidPayouts.reduce((sum, payout) => sum + payout.amount, 0)
  const totalPending = pendingPayouts.reduce((sum, payout) => sum + payout.amount, 0)
  const currency = payouts[0]?.currency || 'ETB'

  function setFormValue(id: string, key: keyof PayoutDetailsForm, value: string) {
    setForms(current => ({
      ...current,
      [id]: {
        ...(current[id] || emptyForm()),
        [key]: value,
      },
    }))
  }

  async function submitDetails(payout: PayoutRow) {
    const form = forms[payout.id] || emptyForm()
    setSaving(payout.id)
    setError('')
    try {
      const payload = {
        payout_method: form.payout_method,
        phone_number: form.phone_number || undefined,
        telebirr_number: form.payout_method === 'telebirr' ? form.telebirr_number : undefined,
        bank_name: form.payout_method === 'bank' ? form.bank_name : undefined,
        bank_account_name: form.payout_method === 'bank' ? form.bank_account_name : undefined,
        bank_account_number: form.payout_method === 'bank' ? form.bank_account_number : undefined,
        payout_instructions: form.payout_method === 'other' ? form.payout_instructions : undefined,
      }
      const updated = await api.post<PayoutRow>(`/api/v1/payouts/${payout.id}/details`, payload)
      setPayouts(current => current.map(row => row.id === updated.id ? updated : row))
      setForms(current => ({ ...current, [updated.id]: formFromPayout(updated) }))
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to save payout details')
    } finally {
      setSaving(null)
    }
  }

  return (
    <div className="space-y-6">
      <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <Link to="/me/dashboard" className="inline-flex items-center gap-2 text-body-sm text-text-secondary hover:text-primary">
            <ArrowLeft className="h-4 w-4" />
            Player dashboard
          </Link>
          <h1 className="mt-2 text-headline-md text-on-surface">My payouts</h1>
          <p className="text-body-md text-text-secondary">When you win, add where the organizer should send your prize money.</p>
        </div>
        <Link to="/tournaments">
          <Button variant="secondary" icon={<Trophy className="h-4 w-4" />}>Browse tournaments</Button>
        </Link>
      </div>

      {error && (
        <Card padding="md" className="border-danger/40 bg-danger/5">
          <p className="text-body-sm text-danger">{error}</p>
        </Card>
      )}

      <div className="grid gap-4 sm:grid-cols-3">
        <StatCard icon={<Receipt />} label="Prize records" value={payouts.length} />
        <StatCard icon={<Wallet />} label="Pending" value={`${totalPending.toLocaleString()} ${currency}`} />
        <StatCard icon={<Trophy />} label="Paid" value={`${totalPaid.toLocaleString()} ${currency}`} />
      </div>

      <Card padding="lg">
        <CardHeader>
          <CardTitle>Prize history</CardTitle>
        </CardHeader>
        <CardContent>
          {payouts.length === 0 ? (
            <EmptyState title="No payouts yet" description="Tournament prizes appear here after you win." />
          ) : (
            <div className="space-y-2">
              {payouts.map(payout => (
                <div key={payout.id} className="rounded-lg border border-border p-3">
                  <div className="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
                    <div className="min-w-0">
                      <p className="text-body-sm font-medium text-on-surface">
                        <Link to={`/tournaments/${payout.tournament_id}`} className="hover:text-primary">
                          {tournamentTitles[payout.tournament_id] || 'Tournament'}
                        </Link>
                      </p>
                      <p className="text-label-md text-text-secondary">
                        Created {formatDate(payout.created_at)}
                        {payout.paid_at ? ` · Paid ${formatDate(payout.paid_at)}` : ''}
                        {payout.payout_details_submitted_at ? ` · Details sent ${formatDate(payout.payout_details_submitted_at)}` : ''}
                      </p>
                      {payout.note && <p className="mt-1 text-label-md text-text-secondary">{payout.note}</p>}
                    </div>
                    <div className="flex items-center justify-between gap-3 sm:block sm:text-right">
                      <p className="text-body-md font-bold text-on-surface">{payout.amount.toLocaleString()} {payout.currency}</p>
                      <Badge status={payout.status === 'paid' ? 'completed' : 'pending'}>{payout.status}</Badge>
                    </div>
                  </div>

                  {payout.status === 'pending' && (
                    <PayoutDetailsEditor
                      payout={payout}
                      form={forms[payout.id] || emptyForm()}
                      saving={saving === payout.id}
                      onChange={(key, value) => setFormValue(payout.id, key, value)}
                      onSubmit={() => submitDetails(payout)}
                    />
                  )}
                </div>
              ))}
            </div>
          )}

          {pendingPayouts.length > 0 && (
            <p className="mt-3 text-body-sm text-text-secondary">
              {pendingPayouts.length} payout{pendingPayouts.length === 1 ? '' : 's'} still pending from organizer.
            </p>
          )}
        </CardContent>
      </Card>
    </div>
  )
}

function PayoutDetailsEditor({
  payout,
  form,
  saving,
  onChange,
  onSubmit,
}: {
  payout: PayoutRow
  form: PayoutDetailsForm
  saving: boolean
  onChange: (key: keyof PayoutDetailsForm, value: string) => void
  onSubmit: () => void
}) {
  const isTelebirr = form.payout_method === 'telebirr'
  const isBank = form.payout_method === 'bank'
  const isOther = form.payout_method === 'other'
  const complete = isTelebirr
    ? form.telebirr_number.trim().length > 0
    : isBank
      ? form.bank_name.trim().length > 0 && form.bank_account_name.trim().length > 0 && form.bank_account_number.trim().length > 0
      : form.payout_instructions.trim().length > 0

  return (
    <div className="mt-4 border-t border-border pt-4">
      <div className="mb-3 flex flex-col gap-2 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <p className="text-body-sm font-medium text-on-surface">Where should the organizer send your prize?</p>
          <p className="text-label-md text-text-secondary">
            {payout.payout_details_submitted_at ? 'The organizer can see these details and pay you.' : 'Choose one payout method and fill in the required fields.'}
          </p>
        </div>
        {payout.payout_details_submitted_at && <Badge status="completed">submitted</Badge>}
      </div>

      <div className="mb-4 grid grid-cols-3 rounded-lg border border-border bg-surface-container-low p-1">
        <button
          type="button"
          onClick={() => onChange('payout_method', 'telebirr')}
          className={`inline-flex h-10 items-center justify-center gap-2 rounded-md px-2 text-body-sm font-medium transition-colors ${isTelebirr ? 'bg-white text-primary shadow-sm' : 'text-text-secondary hover:text-on-surface'}`}
        >
          <Smartphone className="h-4 w-4" />
          Telebirr
        </button>
        <button
          type="button"
          onClick={() => onChange('payout_method', 'bank')}
          className={`inline-flex h-10 items-center justify-center gap-2 rounded-md px-2 text-body-sm font-medium transition-colors ${isBank ? 'bg-white text-primary shadow-sm' : 'text-text-secondary hover:text-on-surface'}`}
        >
          <Landmark className="h-4 w-4" />
          Bank
        </button>
        <button
          type="button"
          onClick={() => onChange('payout_method', 'other')}
          className={`inline-flex h-10 items-center justify-center gap-2 rounded-md px-2 text-body-sm font-medium transition-colors ${isOther ? 'bg-white text-primary shadow-sm' : 'text-text-secondary hover:text-on-surface'}`}
        >
          <MoreHorizontal className="h-4 w-4" />
          Other
        </button>
      </div>

      <div className="grid gap-3 md:grid-cols-2">
        <Input
          label="Phone number"
          value={form.phone_number}
          onChange={event => onChange('phone_number', event.target.value)}
          placeholder="Optional contact phone"
        />

        {isTelebirr && (
          <Input
            label="Telebirr number *"
            value={form.telebirr_number}
            onChange={event => onChange('telebirr_number', event.target.value)}
            placeholder="Telebirr wallet number"
          />
        )}

        {isBank && (
          <>
            <Input
              label="Bank name *"
              value={form.bank_name}
              onChange={event => onChange('bank_name', event.target.value)}
              placeholder="Commercial Bank of Ethiopia"
            />
            <Input
              label="Account name *"
              value={form.bank_account_name}
              onChange={event => onChange('bank_account_name', event.target.value)}
              placeholder="Name on account"
            />
            <Input
              label="Account number *"
              value={form.bank_account_number}
              onChange={event => onChange('bank_account_number', event.target.value)}
              placeholder="Bank account number"
            />
          </>
        )}
      </div>

      {isOther && (
        <div className="mt-3 space-y-1.5">
          <label className="block text-body-sm font-medium text-on-surface">
            Payment instructions *
          </label>
          <textarea
            className="min-h-24 w-full rounded-lg border border-border bg-white px-3 py-2 text-body-md text-on-surface placeholder:text-text-tertiary transition-all duration-150 focus:border-transparent focus:outline-none focus:ring-2 focus:ring-primary"
            value={form.payout_instructions}
            onChange={event => onChange('payout_instructions', event.target.value)}
            placeholder="Example: send by CBE Birr to 09..., or use this wallet/account..."
          />
        </div>
      )}

      <div className="mt-4 flex justify-end">
        <Button
          size="sm"
          loading={saving}
          disabled={!complete}
          icon={<Save className="h-4 w-4" />}
          onClick={onSubmit}
        >
          Send payout details to organizer
        </Button>
      </div>
    </div>
  )
}

function StatCard({ icon, label, value }: { icon: React.ReactNode; label: string; value: number | string }) {
  return (
    <Card padding="md">
      <div className="flex items-center gap-3">
        <div className="flex h-10 w-10 items-center justify-center rounded-lg bg-primary/10 text-primary">{icon}</div>
        <div className="min-w-0">
          <p className="text-label-md text-text-secondary">{label}</p>
          <p className="truncate text-title-lg font-bold text-on-surface">{value}</p>
        </div>
      </div>
    </Card>
  )
}

function emptyForm(): PayoutDetailsForm {
  return {
    payout_method: 'telebirr',
    phone_number: '',
    telebirr_number: '',
    bank_name: '',
    bank_account_name: '',
    bank_account_number: '',
    payout_instructions: '',
  }
}

function formFromPayout(payout: PayoutRow): PayoutDetailsForm {
  return {
    payout_method: payout.payout_method || 'telebirr',
    phone_number: payout.phone_number || '',
    telebirr_number: payout.telebirr_number || '',
    bank_name: payout.bank_name || '',
    bank_account_name: payout.bank_account_name || '',
    bank_account_number: payout.bank_account_number || '',
    payout_instructions: payout.payout_instructions || '',
  }
}
