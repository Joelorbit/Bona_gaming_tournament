import { useEffect, useMemo, useState } from 'react'
import { Link, useNavigate, useParams } from 'react-router-dom'
import { Button } from '@/components/ui/Button'
import { Card, CardHeader, CardTitle, CardContent } from '@/components/ui/Card'
import { Badge } from '@/components/ui/Badge'
import { Tabs } from '@/components/ui/Tabs'
import { Avatar } from '@/components/ui/Avatar'
import { EmptyState } from '@/components/ui/EmptyState'
import { Calendar, Share2, Check, Trophy, MapPin, Users, Clock, Award } from 'lucide-react'
import { api, ApiError, computeFeeBreakdown, type Tournament } from '@/lib/api'
import { useAuth } from '@/contexts/AuthContext'
import { formatCountdown, formatDate } from '@/lib/utils'

interface Registration {
  id: string
  tournament_id: string
  user_id: string
  payment_status: string
  seed?: number | null
  registered_at: string
  username?: string | null
  display_name?: string | null
  avatar_url?: string | null
}

interface Match {
  id: string
  tournament_id: string
  round: number
  position: number
  player_a_id?: string | null
  player_b_id?: string | null
  winner_id?: string | null
  score?: string | null
  status: string
  scheduled_at?: string | null
  completed_at?: string | null
}

interface PaymentResult {
  payment: { id: string; status: string }
  payment_url?: string
}

interface Profile {
  id: string
  username: string
  display_name?: string | null
  avatar_url?: string | null
}

export function TournamentDetails() {
  const { id } = useParams<{ id: string }>()
  const { user } = useAuth()
  const navigate = useNavigate()

  const [tournament, setTournament] = useState<Tournament | null>(null)
  const [participants, setParticipants] = useState<Registration[]>([])
  const [matches, setMatches] = useState<Match[]>([])
  const [organizer, setOrganizer] = useState<Profile | null>(null)
  const [activeTab, setActiveTab] = useState('overview')
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [copied, setCopied] = useState(false)
  const [actionLoading, setActionLoading] = useState(false)
  const [actionError, setActionError] = useState<string | null>(null)
  const [now, setNow] = useState(Date.now())

  useEffect(() => {
    const i = window.setInterval(() => setNow(Date.now()), 60000)
    return () => window.clearInterval(i)
  }, [])

  useEffect(() => {
    if (!id) return
    let cancelled = false
    setLoading(true)
    setError(null)
    Promise.all([
      api.get<Tournament>(`/api/v1/tournaments/${id}`),
      api.get<Registration[]>(`/api/v1/tournaments/${id}/players`).catch(() => []),
      api.get<Match[]>(`/api/v1/tournaments/${id}/bracket`).catch(() => []),
    ])
      .then(([t, p, m]) => {
        if (cancelled) return
        setTournament(t)
        setParticipants(p || [])
        setMatches(m || [])
        api.get<Profile>(`/api/v1/users/id/${t.organizer_id}`).then(profile => {
          if (!cancelled) setOrganizer(profile)
        }).catch(() => {
          if (!cancelled) setOrganizer(null)
        })
      })
      .catch(err => {
        if (cancelled) return
        setError(err instanceof ApiError ? err.message : 'Failed to load tournament')
      })
      .finally(() => { if (!cancelled) setLoading(false) })
    return () => { cancelled = true }
  }, [id])

  const myRegistration = useMemo(() => {
    if (!user) return null
    return participants.find(p => p.user_id === user.id) || null
  }, [participants, user])

  const isOrganizer = !!(tournament && user && tournament.organizer_id === user.id)
  const paidCount = participants.filter(p => p.payment_status === 'paid').length
  const fees = tournament ? computeFeeBreakdown(tournament, paidCount) : null

  const tabs = [
    { id: 'overview', label: 'Overview' },
    { id: 'participants', label: 'Participants', count: participants.length },
    { id: 'bracket', label: 'Bracket' },
    { id: 'matches', label: 'Matches', count: matches.length },
    { id: 'rules', label: 'Rules' },
  ]

  function handleShare() {
    navigator.clipboard.writeText(window.location.href)
    setCopied(true)
    setTimeout(() => setCopied(false), 2000)
  }

  async function refreshParticipants() {
    if (!id) return
    const p = await api.get<Registration[]>(`/api/v1/tournaments/${id}/players`).catch(() => [])
    setParticipants(p || [])
  }

  async function handleRegister() {
    if (!tournament) return
    setActionLoading(true)
    setActionError(null)
    try {
      await api.post(`/api/v1/tournaments/${tournament.id}/join`)
      await refreshParticipants()
      if (tournament.entry_fee > 0) {
        const pay = await api.post<PaymentResult>('/api/v1/payments/create', { tournament_id: tournament.id })
        if (pay.payment_url) {
          window.location.href = pay.payment_url
          return
        }
        setActionError('Payment URL not available')
      }
    } catch (err: any) {
      setActionError(err instanceof ApiError ? err.message : 'Failed to register')
    } finally {
      setActionLoading(false)
    }
  }

  async function handlePay() {
    if (!tournament) return
    setActionLoading(true)
    setActionError(null)
    try {
      const pay = await api.post<PaymentResult>('/api/v1/payments/create', { tournament_id: tournament.id })
      if (pay.payment_url) {
        window.location.href = pay.payment_url
        return
      }
      setActionError('Payment URL not available')
    } catch (err: any) {
      setActionError(err instanceof ApiError ? err.message : 'Failed to start payment')
    } finally {
      setActionLoading(false)
    }
  }

  async function handleLeave() {
    if (!tournament) return
    if (!confirm('Leave this tournament?')) return
    setActionLoading(true)
    try {
      await api.delete(`/api/v1/tournaments/${tournament.id}/leave`)
      await refreshParticipants()
    } catch (err: any) {
      setActionError(err instanceof ApiError ? err.message : 'Failed to leave')
    } finally {
      setActionLoading(false)
    }
  }

  async function handleStatus(newStatus: string) {
    if (!tournament) return
    setActionLoading(true)
    setActionError(null)
    try {
      const updated = await api.patch<Tournament>(`/api/v1/tournaments/${tournament.id}/status`, { status: newStatus })
      setTournament(updated)
    } catch (err: any) {
      setActionError(err instanceof ApiError ? err.message : 'Failed to update status')
    } finally {
      setActionLoading(false)
    }
  }

  async function handleGenerateBracket() {
    if (!tournament) return
    setActionLoading(true)
    setActionError(null)
    try {
      await api.post(`/api/v1/tournaments/${tournament.id}/bracket/generate`)
      const m = await api.get<Match[]>(`/api/v1/tournaments/${tournament.id}/bracket`)
      setMatches(m || [])
      const t = await api.get<Tournament>(`/api/v1/tournaments/${tournament.id}`)
      setTournament(t)
    } catch (err: any) {
      setActionError(err instanceof ApiError ? err.message : 'Failed to generate bracket')
    } finally {
      setActionLoading(false)
    }
  }

  if (loading) return <div className="p-12 text-center text-text-secondary">Loading tournament…</div>
  if (error || !tournament) {
    return <EmptyState title="Tournament not found" description={error || ''} />
  }

  const countdown = formatCountdown(tournament.start_date)
  const spotsRemaining = Math.max(0, tournament.max_participants - paidCount)

  return (
    <div className="space-y-6">
      <div className="flex items-start justify-between gap-4 flex-wrap">
        <div className="space-y-3 flex-1 min-w-0">
          <div className="flex items-center gap-3 flex-wrap">
            <Badge status={tournament.status} />
            <span className="text-body-sm text-text-secondary">{tournament.game}</span>
            <span className="text-body-sm text-text-secondary">· Bo{tournament.best_of}</span>
            <span className="text-body-sm text-text-secondary">· {tournament.format.replace(/_/g, ' ')}</span>
          </div>
          <h1 className="text-headline-lg text-on-surface">{tournament.title}</h1>
          {tournament.description && (
            <p className="text-body-md text-text-secondary max-w-2xl">{tournament.description}</p>
          )}
        </div>
        <div className="flex items-center gap-2">
          <Button variant="secondary" size="md" onClick={handleShare} icon={copied ? <Check className="w-4 h-4 text-success" /> : <Share2 className="w-4 h-4" />}>
            {copied ? 'Copied' : 'Share'}
          </Button>
        </div>
      </div>

      {tournament.status === 'open' && (
        <Card padding="md">
          <div className="flex items-center gap-3 text-body-sm text-on-surface">
            <Clock className="w-5 h-5 text-primary" />
            <span>Starts in <strong>{countdown}</strong> · {formatDate(tournament.start_date)}</span>
          </div>
        </Card>
      )}

      <div className="grid lg:grid-cols-3 gap-6">
        <div className="lg:col-span-2 space-y-6">
          <Tabs tabs={tabs} activeTab={activeTab} onTabChange={setActiveTab} />

          {activeTab === 'overview' && (
            <div className="space-y-6 animate-fade-in">
              <Card padding="lg">
                <CardHeader><CardTitle>Tournament Details</CardTitle></CardHeader>
                <CardContent>
                  <div className="grid grid-cols-2 sm:grid-cols-4 gap-6">
                    <DetailItem icon={<Trophy className="w-4 h-4" />} label="Format" value={`${tournament.format.replace(/_/g, ' ')} · Bo${tournament.best_of}`} />
                    <DetailItem icon={<Users className="w-4 h-4" />} label="Players" value={`${paidCount}/${tournament.max_participants}`} />
                    <DetailItem icon={<Calendar className="w-4 h-4" />} label="Starts" value={formatDate(tournament.start_date)} />
                    <DetailItem icon={<MapPin className="w-4 h-4" />} label="Location" value={tournament.location} />
                  </div>
                </CardContent>
              </Card>

              {fees && (
                <Card padding="lg">
                  <CardHeader><CardTitle>Prize pool breakdown</CardTitle></CardHeader>
                  <CardContent>
                    <p className="text-body-sm text-text-secondary mb-3">
                      Based on {paidCount} paid {paidCount === 1 ? 'participant' : 'participants'} so far.
                    </p>
                    <div className="space-y-1 text-body-sm">
                      <FeeRow label="Collected (entry fees)" value={`${fees.collected.toLocaleString()} ${tournament.currency}`} />
                      <FeeRow label={`Platform fee (${tournament.platform_fee_pct}%)`} value={`−${fees.platform_cut.toLocaleString()} ${tournament.currency}`} />
                      <FeeRow label={`Organizer cut (${tournament.organizer_fee_pct}%)`} value={`−${fees.organizer_cut.toLocaleString()} ${tournament.currency}`} />
                      <div className="flex justify-between border-t border-border pt-2 mt-1 font-medium text-on-surface">
                        <span>Winner prize</span>
                        <span>{fees.winner_prize.toLocaleString()} {tournament.currency}</span>
                      </div>
                    </div>
                  </CardContent>
                </Card>
              )}
            </div>
          )}

          {activeTab === 'participants' && (
            <Card padding="lg" className="animate-fade-in">
              {participants.length === 0 ? (
                <EmptyState title="No one has joined yet" description="Be the first to register." />
              ) : (
                <div className="space-y-2">
                  {participants.map((p, i) => (
                    <Link key={p.id} to={p.username ? `/u/${p.username}` : '#'} className="flex items-center gap-3 p-3 rounded-lg hover:bg-surface-container-low transition-colors">
                      <span className="w-6 text-center text-body-sm font-medium text-text-secondary">#{i + 1}</span>
                      <Avatar src={p.avatar_url || undefined} name={p.display_name || p.username || p.user_id} size="sm" />
                      <div className="flex-1 min-w-0">
                        <p className="text-body-sm font-medium text-on-surface truncate">{p.display_name || p.username || 'Player'}</p>
                        {p.username && <p className="text-label-sm text-text-secondary">@{p.username}</p>}
                      </div>
                      <Badge status={p.payment_status === 'paid' ? 'active' : 'pending'}>
                        {p.payment_status}
                      </Badge>
                    </Link>
                  ))}
                </div>
              )}
            </Card>
          )}

          {activeTab === 'bracket' && (
            <Card padding="lg" className="animate-fade-in">
              {matches.length === 0 ? (
                <EmptyState
                  title="Bracket not generated yet"
                  description={isOrganizer && tournament.status === 'registration_closed' ? 'Generate it from the sidebar.' : 'Available once the organizer generates the bracket.'}
                />
              ) : (
                <BracketGrid matches={matches} participants={participants} />
              )}
            </Card>
          )}

          {activeTab === 'matches' && (
            <Card padding="lg" className="animate-fade-in">
              {matches.length === 0 ? (
                <EmptyState title="No matches yet" />
              ) : (
                <div className="space-y-2">
                  {matches.map(m => (
                    <MatchRow key={m.id} match={m} participants={participants} onClick={() => navigate(`/matches/${m.id}`)} />
                  ))}
                </div>
              )}
            </Card>
          )}

          {activeTab === 'rules' && (
            <Card padding="lg" className="animate-fade-in">
              <div className="prose prose-sm max-w-none text-body-md text-on-surface whitespace-pre-wrap">
                {tournament.rules || 'No rules provided.'}
              </div>
            </Card>
          )}
        </div>

        <div className="space-y-4">
          <Card padding="lg">
            <h3 className="text-title-lg text-on-surface mb-4">Registration</h3>
            <div className="space-y-4">
              <div className="flex items-center justify-between">
                <span className="text-body-sm text-text-secondary">Paid spots</span>
                <span className="text-body-sm font-medium text-on-surface">{paidCount}/{tournament.max_participants}</span>
              </div>
              <div className="w-full h-2 rounded-full bg-surface-container overflow-hidden">
                <div className="h-full rounded-full bg-primary transition-all" style={{ width: `${Math.min(100, (paidCount / tournament.max_participants) * 100)}%` }} />
              </div>

              {actionError && <p className="text-body-sm text-danger" role="alert">{actionError}</p>}

              {myRegistration ? (
                <>
                  <div className="bg-primary-50 text-primary p-3 rounded-lg text-body-sm">
                    {myRegistration.payment_status === 'paid'
                      ? "You're registered and paid."
                      : tournament.entry_fee > 0
                        ? "You're registered. Complete payment to secure your spot."
                        : "You're registered."}
                  </div>
                  {myRegistration.payment_status !== 'paid' && tournament.entry_fee > 0 && (
                    <Button className="w-full" size="lg" onClick={handlePay} loading={actionLoading}>
                      Pay {tournament.entry_fee.toLocaleString()} {tournament.currency}
                    </Button>
                  )}
                  {tournament.status === 'open' && (
                    <Button className="w-full" variant="outline" size="md" onClick={handleLeave} disabled={actionLoading}>
                      Leave tournament
                    </Button>
                  )}
                </>
              ) : tournament.status === 'open' ? (
                <Button className="w-full" size="lg" onClick={handleRegister} loading={actionLoading}>
                  {tournament.entry_fee > 0 ? `Register for ${tournament.entry_fee.toLocaleString()} ${tournament.currency}` : 'Register (free)'}
                </Button>
              ) : (
                <p className="text-center text-body-sm text-text-secondary">Registration is closed.</p>
              )}

              <p className="text-center text-body-sm text-text-secondary">
                {spotsRemaining} {spotsRemaining === 1 ? 'spot' : 'spots'} remaining
              </p>
            </div>
          </Card>

          {isOrganizer && (
            <Card padding="lg">
              <h3 className="text-title-lg text-on-surface mb-4">Organizer actions</h3>
              <div className="space-y-2">
                {tournament.status === 'draft' && (
                  <Button className="w-full" onClick={() => handleStatus('open')} loading={actionLoading}>Open registration</Button>
                )}
                {tournament.status === 'open' && (
                  <Button className="w-full" variant="outline" onClick={() => handleStatus('registration_closed')} loading={actionLoading}>Close registration</Button>
                )}
                {tournament.status === 'registration_closed' && (
                  <>
                    <Button className="w-full" onClick={handleGenerateBracket} loading={actionLoading}>Generate bracket</Button>
                    <Button className="w-full" variant="outline" onClick={() => handleStatus('open')} loading={actionLoading}>Reopen registration</Button>
                  </>
                )}
                {tournament.status === 'in_progress' && (
                  <Button className="w-full" onClick={() => handleStatus('completed')} loading={actionLoading}>Mark completed</Button>
                )}
                {(tournament.status === 'open' || tournament.status === 'in_progress') && (
                  <Button className="w-full" variant="danger" onClick={() => handleStatus('cancelled')} loading={actionLoading}>Cancel tournament</Button>
                )}
              </div>
            </Card>
          )}

          <Card padding="lg">
            <h3 className="text-title-lg text-on-surface mb-3">Organizer</h3>
            <Link to={organizer?.username ? `/u/${organizer.username}` : '#'} className="flex items-center gap-3">
              <Avatar src={organizer?.avatar_url || undefined} name={organizer?.display_name || organizer?.username || tournament.organizer_id} size="md" />
              <div>
                <p className="text-body-sm font-medium text-on-surface">{organizer?.display_name || organizer?.username || 'Organizer'}</p>
                <p className="text-label-sm text-text-secondary truncate max-w-[200px]">{organizer?.username ? `@${organizer.username}` : 'Profile unavailable'}</p>
              </div>
            </Link>
          </Card>
        </div>
      </div>
    </div>
  )
}

function DetailItem({ icon, label, value }: { icon: React.ReactNode; label: string; value: string }) {
  return (
    <div>
      <p className="text-label-sm text-text-secondary mb-1 flex items-center gap-1.5">
        <span className="text-text-secondary">{icon}</span>
        {label}
      </p>
      <p className="text-body-sm font-medium text-on-surface">{value}</p>
    </div>
  )
}

function FeeRow({ label, value }: { label: string; value: string }) {
  return (
    <div className="flex justify-between">
      <span className="text-text-secondary">{label}</span>
      <span>{value}</span>
    </div>
  )
}

function MatchRow({ match, participants, onClick }: { match: Match; participants: Registration[]; onClick: () => void }) {
  const a = participants.find(p => p.user_id === match.player_a_id)
  const b = participants.find(p => p.user_id === match.player_b_id)
  const winnerIsA = match.winner_id && match.winner_id === match.player_a_id
  const winnerIsB = match.winner_id && match.winner_id === match.player_b_id
  return (
    <button onClick={onClick} className="w-full text-left p-3 rounded-lg border border-border hover:border-primary/50 hover:bg-surface-container-low transition-all">
      <div className="flex items-center justify-between mb-2">
        <span className="text-label-md text-text-secondary">Round {match.round} · Match {match.position + 1}</span>
        <Badge status={match.status === 'completed' ? 'completed' : match.status === 'walkover' ? 'completed' : 'pending'}>
          {match.status}
        </Badge>
      </div>
      <div className="flex items-center justify-between text-body-sm">
        <div className={`flex items-center gap-2 ${winnerIsA ? 'font-semibold text-on-surface' : 'text-text-secondary'}`}>
          {winnerIsA && <Award className="w-4 h-4 text-warning-500" />}
          {a?.display_name || a?.username || (match.player_a_id ? 'Profile pending' : 'TBD')}
        </div>
        <span className="text-text-secondary">vs</span>
        <div className={`flex items-center gap-2 ${winnerIsB ? 'font-semibold text-on-surface' : 'text-text-secondary'}`}>
          {b?.display_name || b?.username || (match.player_b_id ? 'Profile pending' : 'TBD')}
          {winnerIsB && <Award className="w-4 h-4 text-warning-500" />}
        </div>
      </div>
      {match.score && <p className="text-label-sm text-text-secondary mt-1 text-center">{match.score}</p>}
    </button>
  )
}

function BracketGrid({ matches, participants }: { matches: Match[]; participants: Registration[] }) {
  const rounds = useMemo(() => {
    const m = new Map<number, Match[]>()
    matches.forEach(match => {
      const arr = m.get(match.round) || []
      arr.push(match)
      m.set(match.round, arr)
    })
    return Array.from(m.entries()).sort((a, b) => a[0] - b[0])
  }, [matches])

  return (
    <div className="flex gap-4 overflow-x-auto pb-2">
      {rounds.map(([round, rm]) => (
        <div key={round} className="min-w-[200px] flex-1">
          <p className="text-label-md text-text-secondary mb-2">Round {round}</p>
          <div className="space-y-3">
            {rm.sort((a, b) => a.position - b.position).map(match => {
              const a = participants.find(p => p.user_id === match.player_a_id)
              const b = participants.find(p => p.user_id === match.player_b_id)
              const winA = match.winner_id === match.player_a_id
              const winB = match.winner_id === match.player_b_id
              return (
                <div key={match.id} className="border border-border rounded-lg overflow-hidden text-body-sm">
                  <div className={`px-3 py-2 ${winA ? 'bg-success/10 font-semibold' : ''}`}>
                    {a?.display_name || a?.username || (match.player_a_id ? '—' : 'TBD')}
                  </div>
                  <div className={`px-3 py-2 border-t border-border ${winB ? 'bg-success/10 font-semibold' : ''}`}>
                    {b?.display_name || b?.username || (match.player_b_id ? '—' : 'TBD')}
                  </div>
                  {match.score && <div className="px-3 py-1 border-t border-border bg-surface-container-low text-label-sm text-text-secondary text-center">{match.score}</div>}
                </div>
              )
            })}
          </div>
        </div>
      ))}
    </div>
  )
}
