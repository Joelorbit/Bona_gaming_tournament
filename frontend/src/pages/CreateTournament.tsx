import { useState, type FormEvent } from 'react'
import { useNavigate } from 'react-router-dom'
import { Button } from '@/components/ui/Button'
import { Card, CardHeader, CardTitle, CardContent } from '@/components/ui/Card'
import { Input } from '@/components/ui/Input'
import { ArrowLeft, ArrowRight, Check, Trophy, Gamepad2, Users, DollarSign, FileText } from 'lucide-react'
import { api, ApiError, computeFeeBreakdown, type Tournament } from '@/lib/api'

const steps = [
  { id: 'basic', label: 'Basic Info', icon: Trophy },
  { id: 'details', label: 'Tournament Details', icon: Gamepad2 },
  { id: 'participants', label: 'Participants', icon: Users },
  { id: 'prize', label: 'Prize & Fees', icon: DollarSign },
  { id: 'rules', label: 'Rules & Publishing', icon: FileText },
]

const FORMATS = [
  { value: 'single_elimination', label: 'Single Elimination', desc: "Lose once, you're out" },
  { value: 'double_elimination', label: 'Double Elimination', desc: 'Second chance bracket' },
  { value: 'round_robin', label: 'Round Robin', desc: 'Everyone plays everyone' },
  { value: 'swiss', label: 'Swiss', desc: 'Matchmaking by record' },
] as const

const TEAM_SIZES = [
  { value: 1, label: '1v1' },
  { value: 2, label: '2v2 (Duo)' },
  { value: 3, label: '3v3 (Trio)' },
  { value: 5, label: '5v5 (Team)' },
] as const

const BEST_OF = [1, 3, 5, 7] as const

const LOCATIONS = ['Online', 'In-Person', 'Hybrid'] as const

interface Form {
  title: string
  game: string
  description: string
  start_date: string
  end_date: string
  registration_close_at: string
  format: string
  best_of: number
  location: string
  max_participants: number
  min_participants: number
  team_size: number
  prize_pool: number
  entry_fee: number
  organizer_fee_pct: number
  currency: string
  rules: string
}

const initialForm: Form = {
  title: '',
  game: '',
  description: '',
  start_date: '',
  end_date: '',
  registration_close_at: '',
  format: 'single_elimination',
  best_of: 1,
  location: 'Online',
  max_participants: 64,
  min_participants: 2,
  team_size: 1,
  prize_pool: 0,
  entry_fee: 0,
  organizer_fee_pct: 0,
  currency: 'ETB',
  rules: '',
}

export function CreateTournament() {
  const [currentStep, setCurrentStep] = useState(0)
  const [form, setForm] = useState<Form>(initialForm)
  const [submitting, setSubmitting] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const navigate = useNavigate()

  const set = <K extends keyof Form>(k: K, v: Form[K]) => setForm(f => ({ ...f, [k]: v }))

  const canAdvance = () => {
    if (currentStep === 0) return form.title.trim() && form.game.trim() && form.start_date
    if (currentStep === 2) return form.max_participants >= form.min_participants && form.min_participants >= 2
    if (currentStep === 3) return form.organizer_fee_pct >= 0 && form.organizer_fee_pct <= 20
    return true
  }

  async function onSubmit(e: FormEvent) {
    e.preventDefault()
    if (!form.title.trim() || !form.game.trim() || !form.start_date) {
      setError('Title, game, and start date are required')
      setCurrentStep(0)
      return
    }
    setSubmitting(true)
    setError(null)
    try {
      const created = await api.post<Tournament>('/api/v1/tournaments', {
        title: form.title,
        game: form.game,
        description: form.description || null,
        rules: form.rules || null,
        format: form.format,
        best_of: form.best_of,
        max_participants: form.max_participants,
        min_participants: form.min_participants,
        team_size: form.team_size,
        entry_fee: form.entry_fee,
        organizer_fee_pct: form.organizer_fee_pct,
        prize_pool: form.prize_pool,
        currency: form.currency,
        location: form.location,
        start_date: new Date(form.start_date).toISOString(),
        end_date: form.end_date ? new Date(form.end_date).toISOString() : null,
        registration_close_at: form.registration_close_at ? new Date(form.registration_close_at).toISOString() : null,
      })
      navigate(`/tournaments/${created.id}`)
    } catch (err: any) {
      const msg = err instanceof ApiError ? err.message : 'Failed to create tournament'
      setError(msg)
    } finally {
      setSubmitting(false)
    }
  }

  const previewParticipants = Math.min(form.max_participants, Math.max(form.min_participants, 8))
  const fees = computeFeeBreakdown({
    entry_fee: form.entry_fee,
    platform_fee_pct: 5,
    organizer_fee_pct: form.organizer_fee_pct,
  }, previewParticipants)

  return (
    <form onSubmit={onSubmit} className="max-w-3xl mx-auto space-y-6">
      <div>
        <button type="button" onClick={() => navigate(-1)} className="flex items-center gap-2 text-body-sm text-text-secondary hover:text-on-surface mb-4">
          <ArrowLeft className="w-4 h-4" /> Back
        </button>
        <h1 className="text-headline-md text-on-surface">Create Tournament</h1>
        <p className="text-body-md text-text-secondary">Set up your tournament in minutes</p>
      </div>

      <div className="flex gap-2">
        {steps.map((step, i) => (
          <div key={step.id} className="flex-1">
            <div className={`h-1.5 rounded-full transition-colors ${i <= currentStep ? 'bg-primary' : 'bg-surface-container-high'}`} />
            <div className="flex items-center gap-2 mt-2">
              <div className={`w-6 h-6 rounded-full flex items-center justify-center ${
                i < currentStep ? 'bg-primary text-white' :
                i === currentStep ? 'bg-primary-50 text-primary border-2 border-primary' :
                'bg-surface-container text-text-secondary'
              }`}>
                {i < currentStep ? <Check className="w-3 h-3" /> : <step.icon className="w-3 h-3" />}
              </div>
              <span className={`text-label-sm hidden sm:block ${i <= currentStep ? 'text-on-surface font-medium' : 'text-text-secondary'}`}>
                {step.label}
              </span>
            </div>
          </div>
        ))}
      </div>

      <Card padding="lg">
        <CardHeader>
          <CardTitle>{steps[currentStep].label}</CardTitle>
        </CardHeader>
        <CardContent>
          {currentStep === 0 && (
            <div className="space-y-4 animate-fade-in">
              <Input label="Tournament Title" required placeholder="e.g. Bona Championship Series"
                value={form.title} onChange={e => set('title', e.target.value)} />
              <Input label="Game" required placeholder="e.g. Valorant, FIFA, Tekken"
                value={form.game} onChange={e => set('game', e.target.value)} />
              <div className="grid grid-cols-2 gap-4">
                <Input label="Start Date" type="datetime-local" required
                  value={form.start_date} onChange={e => set('start_date', e.target.value)} />
                <Input label="End Date" type="datetime-local"
                  value={form.end_date} onChange={e => set('end_date', e.target.value)} />
              </div>
              <Input label="Registration closes" type="datetime-local"
                value={form.registration_close_at} onChange={e => set('registration_close_at', e.target.value)} />
              <div>
                <label className="block text-body-sm font-medium text-on-surface mb-1.5">Description</label>
                <textarea
                  className="w-full h-32 rounded-lg border border-border bg-white px-3 py-2 text-body-md text-on-surface placeholder:text-text-tertiary focus:outline-none focus:ring-2 focus:ring-primary focus:border-transparent resize-none"
                  placeholder="Describe your tournament..."
                  value={form.description}
                  onChange={e => set('description', e.target.value)}
                />
              </div>
            </div>
          )}

          {currentStep === 1 && (
            <div className="space-y-4 animate-fade-in">
              <div>
                <label className="block text-body-sm font-medium text-on-surface mb-1.5">Tournament Format</label>
                <div className="grid grid-cols-2 gap-3">
                  {FORMATS.map(f => (
                    <button
                      key={f.value}
                      type="button"
                      onClick={() => set('format', f.value)}
                      className={`text-left p-4 rounded-xl border transition-all ${
                        form.format === f.value
                          ? 'border-primary bg-primary-50/50'
                          : 'border-border hover:border-primary hover:bg-primary-50/50'
                      }`}
                    >
                      <p className="text-body-sm font-medium text-on-surface">{f.label}</p>
                      <p className="text-label-sm text-text-secondary mt-1">{f.desc}</p>
                    </button>
                  ))}
                </div>
              </div>
              <div>
                <label className="block text-body-sm font-medium text-on-surface mb-1.5">Best of</label>
                <div className="flex gap-3">
                  {BEST_OF.map(n => (
                    <button
                      key={n}
                      type="button"
                      onClick={() => set('best_of', n)}
                      className={`px-4 py-2 rounded-lg border text-body-sm text-on-surface transition-all ${
                        form.best_of === n
                          ? 'border-primary bg-primary-50/50'
                          : 'border-border hover:border-primary hover:bg-primary-50/50'
                      }`}
                    >
                      Bo{n}
                    </button>
                  ))}
                </div>
              </div>
              <div>
                <label className="block text-body-sm font-medium text-on-surface mb-1.5">Location</label>
                <div className="flex gap-3">
                  {LOCATIONS.map(loc => (
                    <button
                      key={loc}
                      type="button"
                      onClick={() => set('location', loc)}
                      className={`px-4 py-2 rounded-lg border text-body-sm text-on-surface transition-all ${
                        form.location === loc
                          ? 'border-primary bg-primary-50/50'
                          : 'border-border hover:border-primary hover:bg-primary-50/50'
                      }`}
                    >
                      {loc}
                    </button>
                  ))}
                </div>
              </div>
            </div>
          )}

          {currentStep === 2 && (
            <div className="space-y-4 animate-fade-in">
              <div className="grid grid-cols-2 gap-4">
                <Input label="Max Participants" type="number" min={2}
                  value={form.max_participants}
                  onChange={e => set('max_participants', Math.max(2, parseInt(e.target.value || '0', 10)))} />
                <Input label="Min Participants" type="number" min={2}
                  value={form.min_participants}
                  onChange={e => set('min_participants', Math.max(2, parseInt(e.target.value || '0', 10)))} />
              </div>
              <div>
                <label className="block text-body-sm font-medium text-on-surface mb-1.5">Team Size</label>
                <div className="flex gap-3">
                  {TEAM_SIZES.map(t => (
                    <button
                      key={t.value}
                      type="button"
                      onClick={() => set('team_size', t.value)}
                      className={`px-4 py-2 rounded-lg border text-body-sm text-on-surface transition-all ${
                        form.team_size === t.value
                          ? 'border-primary bg-primary-50/50'
                          : 'border-border hover:border-primary hover:bg-primary-50/50'
                      }`}
                    >
                      {t.label}
                    </button>
                  ))}
                </div>
              </div>
            </div>
          )}

          {currentStep === 3 && (
            <div className="space-y-4 animate-fade-in">
              <Input label="Prize Pool (auto-funded by entry fees if 0)" type="number" min={0}
                value={form.prize_pool}
                onChange={e => set('prize_pool', Math.max(0, parseInt(e.target.value || '0', 10)))} />
              <Input label="Entry Fee (0 = free)" type="number" min={0}
                value={form.entry_fee}
                onChange={e => set('entry_fee', Math.max(0, parseInt(e.target.value || '0', 10)))} />
              <Input label="Currency" placeholder="ETB"
                value={form.currency}
                onChange={e => set('currency', e.target.value.toUpperCase().slice(0, 3))} />
              <div>
                <label className="block text-body-sm font-medium text-on-surface mb-1.5">
                  Organizer fee ({form.organizer_fee_pct}%) — max 20%
                </label>
                <input
                  type="range"
                  min={0}
                  max={20}
                  step={1}
                  value={form.organizer_fee_pct}
                  onChange={e => set('organizer_fee_pct', parseInt(e.target.value, 10))}
                  className="w-full"
                />
              </div>

              <div className="bg-surface-container-low rounded-xl p-4">
                <p className="text-body-sm font-medium text-on-surface mb-2">
                  Fee breakdown preview ({previewParticipants} paid participants)
                </p>
                <div className="space-y-1 text-body-sm">
                  <div className="flex justify-between"><span className="text-text-secondary">Collected</span><span>{fees.collected.toLocaleString()} {form.currency}</span></div>
                  <div className="flex justify-between"><span className="text-text-secondary">Platform fee (5%)</span><span>−{fees.platform_cut.toLocaleString()} {form.currency}</span></div>
                  <div className="flex justify-between"><span className="text-text-secondary">Organizer cut ({form.organizer_fee_pct}%)</span><span>−{fees.organizer_cut.toLocaleString()} {form.currency}</span></div>
                  <div className="flex justify-between border-t border-border pt-2 mt-1 font-medium text-on-surface"><span>Winner prize</span><span>{fees.winner_prize.toLocaleString()} {form.currency}</span></div>
                </div>
              </div>
            </div>
          )}

          {currentStep === 4 && (
            <div className="space-y-4 animate-fade-in">
              <div>
                <label className="block text-body-sm font-medium text-on-surface mb-1.5">Tournament Rules</label>
                <textarea
                  className="w-full h-40 rounded-lg border border-border bg-white px-3 py-2 text-body-md text-on-surface placeholder:text-text-tertiary focus:outline-none focus:ring-2 focus:ring-primary focus:border-transparent resize-none"
                  placeholder="Enter the rules and regulations for your tournament..."
                  value={form.rules}
                  onChange={e => set('rules', e.target.value)}
                />
              </div>
              <div className="bg-surface-container-low rounded-xl p-4">
                <h4 className="text-body-sm font-medium text-on-surface mb-2">Summary</h4>
                <div className="space-y-1 text-body-sm text-text-secondary">
                  <p>{form.title || '(no title)'} — {form.game || '(no game)'}</p>
                  <p>Format: {form.format.replace(/_/g, ' ')} · Bo{form.best_of}</p>
                  <p>Participants: {form.min_participants}–{form.max_participants}, team size {form.team_size}</p>
                  <p>Prize Pool: {form.prize_pool.toLocaleString()} {form.currency}</p>
                  <p>Entry Fee: {form.entry_fee.toLocaleString()} {form.currency}</p>
                  <p>Fees: 5% platform + {form.organizer_fee_pct}% organizer</p>
                  <p>Starts: {form.start_date || '(no date)'}</p>
                  {form.registration_close_at && <p>Registration closes: {form.registration_close_at}</p>}
                </div>
              </div>
              {error && <p className="text-body-sm text-danger" role="alert">{error}</p>}
            </div>
          )}
        </CardContent>
      </Card>

      <div className="flex items-center justify-between">
        <Button
          type="button"
          variant="ghost"
          onClick={() => setCurrentStep(Math.max(0, currentStep - 1))}
          disabled={currentStep === 0 || submitting}
        >
          Previous
        </Button>
        <div className="flex items-center gap-2">
          <span className="text-body-sm text-text-secondary">
            Step {currentStep + 1} of {steps.length}
          </span>
          {currentStep < steps.length - 1 ? (
            <Button
              type="button"
              onClick={() => canAdvance() ? setCurrentStep(currentStep + 1) : setError('Please complete the required fields')}
              icon={<ArrowRight className="w-4 h-4" />}
            >
              Next Step
            </Button>
          ) : (
            <Button type="submit" loading={submitting} disabled={submitting} icon={<Check className="w-4 h-4" />}>
              Publish Tournament
            </Button>
          )}
        </div>
      </div>
    </form>
  )
}
