import { useEffect, useState, type FormEvent } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { Card, CardHeader, CardTitle, CardContent } from '@/components/ui/Card'
import { Button } from '@/components/ui/Button'
import { Input } from '@/components/ui/Input'
import { Badge } from '@/components/ui/Badge'
import { EmptyState } from '@/components/ui/EmptyState'
import { api, ApiError, type Tournament } from '@/lib/api'
import { useAuth } from '@/contexts/AuthContext'
import { formatDate } from '@/lib/utils'

export interface Match {
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
  evidence_screenshot_url?: string | null
  evidence_video_url?: string | null
  evidence_notes?: string | null
  result_submitted_by?: string | null
  result_confirmed_at?: string | null
  dispute_status: string
  dispute_reason?: string | null
  dispute_opened_by?: string | null
  dispute_opened_at?: string | null
  dispute_resolved_at?: string | null
}

interface Registration {
  user_id: string
  username?: string | null
  display_name?: string | null
  avatar_url?: string | null
}

export function MatchDetail() {
  const { id } = useParams<{ id: string }>()
  const { user } = useAuth()
  const navigate = useNavigate()

  const [match, setMatch] = useState<Match | null>(null)
  const [tournament, setTournament] = useState<Tournament | null>(null)
  const [players, setPlayers] = useState<Registration[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [actionError, setActionError] = useState<string | null>(null)
  const [submitting, setSubmitting] = useState(false)

  const [winnerId, setWinnerId] = useState('')
  const [score, setScore] = useState('')
  const [screenshot, setScreenshot] = useState('')
  const [video, setVideo] = useState('')
  const [notes, setNotes] = useState('')
  const [disputeReason, setDisputeReason] = useState('')

  async function load(matchId: string) {
    setLoading(true)
    try {
      const m = await api.get<Match>(`/api/v1/matches/${matchId}`)
      setMatch(m)
      const t = await api.get<Tournament>(`/api/v1/tournaments/${m.tournament_id}`)
      setTournament(t)
      const p = await api.get<Registration[]>(`/api/v1/tournaments/${m.tournament_id}/players`).catch(() => [])
      setPlayers(p || [])
    } catch (err: any) {
      setError(err instanceof ApiError ? err.message : 'Failed to load match')
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => { if (id) load(id) }, [id])

  if (loading) return <div className="p-12 text-center text-text-secondary">Loading match…</div>
  if (error || !match || !tournament) {
    return <EmptyState title="Match not found" description={error || ''} />
  }

  const playerA = players.find(p => p.user_id === match.player_a_id)
  const playerB = players.find(p => p.user_id === match.player_b_id)
  const isParticipant = user && (match.player_a_id === user.id || match.player_b_id === user.id)
  const isOrganizer = user && tournament.organizer_id === user.id
  const submittedByMe = match.result_submitted_by === user?.id
  const canSubmit = isParticipant && (match.status === 'pending' || match.status === 'in_progress')
  const canConfirmOrDispute = isParticipant && match.status === 'awaiting_confirmation' && !submittedByMe
  const canResolve = isOrganizer && match.dispute_status === 'pending'

  function nameFor(id: string | null | undefined) {
    if (!id) return 'TBD'
    const p = players.find(x => x.user_id === id)
    return p?.display_name || p?.username || id.slice(0, 8)
  }

  async function handleSubmit(e: FormEvent) {
    e.preventDefault()
    if (!match) return
    setSubmitting(true)
    setActionError(null)
    try {
      await api.post(`/api/v1/matches/${match.id}/result`, {
        winner_id: winnerId,
        score,
        evidence_screenshot_url: screenshot || null,
        evidence_video_url: video || null,
        evidence_notes: notes || null,
      })
      await load(match.id)
    } catch (err: any) {
      setActionError(err instanceof ApiError ? err.message : 'Failed to submit result')
    } finally {
      setSubmitting(false)
    }
  }

  async function handleConfirm() {
    if (!match) return
    setSubmitting(true)
    setActionError(null)
    try {
      await api.post(`/api/v1/matches/${match.id}/confirm`)
      await load(match.id)
    } catch (err: any) {
      setActionError(err instanceof ApiError ? err.message : 'Failed to confirm')
    } finally {
      setSubmitting(false)
    }
  }

  async function handleDispute() {
    if (!match) return
    if (!disputeReason.trim()) {
      setActionError('Please describe the issue')
      return
    }
    setSubmitting(true)
    setActionError(null)
    try {
      await api.post(`/api/v1/matches/${match.id}/dispute`, { reason: disputeReason })
      await load(match.id)
    } catch (err: any) {
      setActionError(err instanceof ApiError ? err.message : 'Failed to open dispute')
    } finally {
      setSubmitting(false)
    }
  }

  async function handleResolve(winnerId: string, score: string) {
    if (!match) return
    setSubmitting(true)
    setActionError(null)
    try {
      await api.post(`/api/v1/matches/${match.id}/resolve`, { winner_id: winnerId, score })
      await load(match.id)
    } catch (err: any) {
      setActionError(err instanceof ApiError ? err.message : 'Failed to resolve dispute')
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <div className="max-w-3xl mx-auto space-y-6">
      <div>
        <button onClick={() => navigate(`/tournaments/${match.tournament_id}`)} className="text-body-sm text-text-secondary hover:text-on-surface mb-2">
          ← Back to tournament
        </button>
        <div className="flex items-center gap-3 flex-wrap">
          <Badge status={match.status === 'completed' ? 'completed' : match.status === 'disputed' ? 'cancelled' : 'pending'}>
            {match.status.replace(/_/g, ' ')}
          </Badge>
          <span className="text-body-sm text-text-secondary">
            {tournament.title} · Round {match.round} · Match {match.position + 1}
          </span>
        </div>
      </div>

      <Card padding="lg">
        <div className="grid grid-cols-3 items-center gap-4">
          <div className="text-center">
            <p className="text-label-sm text-text-secondary">Player A</p>
            <p className={`text-title-lg ${match.winner_id === match.player_a_id ? 'font-bold text-success' : 'text-on-surface'}`}>
              {nameFor(match.player_a_id)}
            </p>
          </div>
          <div className="text-center">
            <p className="text-headline-sm text-text-secondary">vs</p>
            {match.score && <p className="text-body-md font-medium text-on-surface mt-1">{match.score}</p>}
          </div>
          <div className="text-center">
            <p className="text-label-sm text-text-secondary">Player B</p>
            <p className={`text-title-lg ${match.winner_id === match.player_b_id ? 'font-bold text-success' : 'text-on-surface'}`}>
              {nameFor(match.player_b_id)}
            </p>
          </div>
        </div>
        {match.completed_at && (
          <p className="text-center text-label-md text-text-secondary mt-4">
            Completed {formatDate(match.completed_at)}
          </p>
        )}
      </Card>

      {actionError && <Card padding="md"><p className="text-body-sm text-danger">{actionError}</p></Card>}

      {canSubmit && (
        <Card padding="lg">
          <CardHeader><CardTitle>Submit result</CardTitle></CardHeader>
          <CardContent>
            <form onSubmit={handleSubmit} className="space-y-4">
              <div>
                <label className="block text-label-md text-on-surface mb-2">Winner</label>
                <div className="grid grid-cols-2 gap-3">
                  {[match.player_a_id, match.player_b_id].filter(Boolean).map(pid => (
                    <button
                      type="button"
                      key={pid}
                      onClick={() => setWinnerId(pid as string)}
                      className={`p-3 rounded-lg border text-body-md ${
                        winnerId === pid ? 'border-primary bg-primary-50/50 font-medium' : 'border-border hover:border-primary'
                      }`}
                    >
                      {nameFor(pid)}
                    </button>
                  ))}
                </div>
              </div>
              <Input label="Score (e.g. 2-1)" value={score} onChange={e => setScore(e.target.value)} required />
              <Input label="Screenshot URL (optional)" value={screenshot} onChange={e => setScreenshot(e.target.value)} placeholder="https://…" />
              <Input label="Video URL (optional)" value={video} onChange={e => setVideo(e.target.value)} placeholder="https://…" />
              <div>
                <label className="block text-label-md text-on-surface mb-2">Notes (optional)</label>
                <textarea
                  className="w-full px-3 py-2 border border-border rounded-lg min-h-[80px]"
                  value={notes}
                  onChange={e => setNotes(e.target.value)}
                />
              </div>
              <Button type="submit" loading={submitting} disabled={!winnerId || !score}>
                Submit result
              </Button>
            </form>
          </CardContent>
        </Card>
      )}

      {match.status === 'awaiting_confirmation' && (
        <Card padding="lg">
          <CardHeader><CardTitle>Result submitted — awaiting confirmation</CardTitle></CardHeader>
          <CardContent>
            <p className="text-body-md text-on-surface mb-2">
              Winner: <strong>{nameFor(match.winner_id)}</strong>
              {match.score && <> · Score: {match.score}</>}
            </p>
            {match.evidence_screenshot_url && <p className="text-body-sm"><a href={match.evidence_screenshot_url} target="_blank" rel="noreferrer" className="text-primary hover:underline">View screenshot</a></p>}
            {match.evidence_video_url && <p className="text-body-sm"><a href={match.evidence_video_url} target="_blank" rel="noreferrer" className="text-primary hover:underline">View video</a></p>}
            {match.evidence_notes && <p className="text-body-sm text-text-secondary mt-2 whitespace-pre-wrap">{match.evidence_notes}</p>}

            {canConfirmOrDispute && (
              <div className="space-y-4 mt-4 pt-4 border-t border-border">
                <Button onClick={handleConfirm} loading={submitting} className="w-full">
                  Confirm this result
                </Button>
                <div>
                  <label className="block text-label-md text-on-surface mb-2">Or dispute the result</label>
                  <textarea
                    className="w-full px-3 py-2 border border-border rounded-lg min-h-[80px] mb-2"
                    placeholder="Explain what's wrong"
                    value={disputeReason}
                    onChange={e => setDisputeReason(e.target.value)}
                  />
                  <Button variant="outline" onClick={handleDispute} loading={submitting} disabled={!disputeReason.trim()}>
                    Open dispute
                  </Button>
                </div>
              </div>
            )}
            {submittedByMe && <p className="text-body-sm text-text-secondary mt-3">Waiting for your opponent to confirm or dispute.</p>}
          </CardContent>
        </Card>
      )}

      {match.dispute_status === 'pending' && (
        <Card padding="lg" className="border-danger/40 bg-danger/5">
          <CardHeader><CardTitle>Dispute pending</CardTitle></CardHeader>
          <CardContent>
            <p className="text-body-md text-on-surface mb-2 whitespace-pre-wrap">{match.dispute_reason}</p>
            <p className="text-label-md text-text-secondary">
              Opened by {nameFor(match.dispute_opened_by)} {match.dispute_opened_at && `· ${formatDate(match.dispute_opened_at)}`}
            </p>
            {canResolve && (
              <div className="mt-4 pt-4 border-t border-border space-y-3">
                <p className="text-body-sm text-on-surface">Resolve by declaring the winner:</p>
                <div className="grid grid-cols-2 gap-3">
                  <Button onClick={() => handleResolve(match.player_a_id || '', match.score || '')} loading={submitting}>
                    {nameFor(match.player_a_id)} wins
                  </Button>
                  <Button onClick={() => handleResolve(match.player_b_id || '', match.score || '')} loading={submitting}>
                    {nameFor(match.player_b_id)} wins
                  </Button>
                </div>
              </div>
            )}
          </CardContent>
        </Card>
      )}

      {match.status === 'completed' && match.evidence_notes && (
        <Card padding="lg">
          <CardHeader><CardTitle>Evidence</CardTitle></CardHeader>
          <CardContent>
            {match.evidence_screenshot_url && <p className="text-body-sm"><a href={match.evidence_screenshot_url} target="_blank" rel="noreferrer" className="text-primary hover:underline">Screenshot</a></p>}
            {match.evidence_video_url && <p className="text-body-sm"><a href={match.evidence_video_url} target="_blank" rel="noreferrer" className="text-primary hover:underline">Video</a></p>}
            <p className="text-body-sm text-text-secondary mt-2 whitespace-pre-wrap">{match.evidence_notes}</p>
          </CardContent>
        </Card>
      )}
    </div>
  )
}
