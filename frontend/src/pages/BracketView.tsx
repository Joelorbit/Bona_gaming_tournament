import { useEffect, useMemo, useState } from 'react'
import { Link, useNavigate, useParams } from 'react-router-dom'
import { Card } from '@/components/ui/Card'
import { Badge } from '@/components/ui/Badge'
import { EmptyState } from '@/components/ui/EmptyState'
import { Trophy } from 'lucide-react'
import { api, ApiError, type Tournament } from '@/lib/api'

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
}

export function BracketView() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const [tournament, setTournament] = useState<Tournament | null>(null)
  const [participants, setParticipants] = useState<Registration[]>([])
  const [matches, setMatches] = useState<Match[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

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
      })
      .catch(err => {
        if (cancelled) return
        setError(err instanceof ApiError ? err.message : 'Could not load bracket')
      })
      .finally(() => { if (!cancelled) setLoading(false) })
    return () => { cancelled = true }
  }, [id])

  const rounds = useMemo(() => {
    const grouped = new Map<number, Match[]>()
    matches.forEach(match => {
      const existing = grouped.get(match.round) || []
      existing.push(match)
      grouped.set(match.round, existing)
    })
    return Array.from(grouped.entries())
      .sort((a, b) => a[0] - b[0])
      .map(([round, roundMatches]) => [round, roundMatches.sort((a, b) => a.position - b.position)] as const)
  }, [matches])

  const champion = useMemo(() => {
    const finalMatch = [...matches].sort((a, b) => b.round - a.round || b.position - a.position)[0]
    if (!finalMatch?.winner_id) return null
    return participants.find(p => p.user_id === finalMatch.winner_id) || null
  }, [matches, participants])

  if (loading) return <div className="p-12 text-center text-text-secondary">Loading bracket...</div>
  if (error) return <EmptyState title="Could not load bracket" description={error} />

  return (
    <div className="space-y-6">
      <div className="flex items-start justify-between gap-4 flex-wrap">
        <div>
          <h1 className="text-headline-md text-on-surface">Bracket</h1>
          <p className="text-body-md text-text-secondary">{tournament?.title || 'Tournament'}</p>
        </div>
        {id && (
          <Link to={`/tournaments/${id}`} className="text-body-sm text-primary hover:text-primary-600">
            Back to tournament
          </Link>
        )}
      </div>

      {matches.length === 0 ? (
        <EmptyState title="No bracket yet" description="The organizer has not generated a bracket from paid participants." />
      ) : (
        <div className="overflow-x-auto pb-4">
          <div className="flex gap-6 min-w-max">
            {rounds.map(([round, roundMatches]) => (
              <div key={round} className="flex flex-col gap-4 min-w-[230px]">
                <div className="text-center">
                  <h2 className="text-title-lg text-on-surface mb-1">{roundLabel(round, rounds.length)}</h2>
                  <p className="text-label-sm text-text-secondary">{roundMatches.length} match{roundMatches.length === 1 ? '' : 'es'}</p>
                </div>
                <div className="flex flex-col justify-around flex-1 gap-3">
                  {roundMatches.map(match => (
                    <MatchCard key={match.id} match={match} participants={participants} onClick={() => navigate(`/matches/${match.id}`)} />
                  ))}
                </div>
              </div>
            ))}

            {champion && (
              <div className="flex flex-col justify-center min-w-[230px]">
                <div className="text-center mb-4">
                  <h2 className="text-title-lg text-primary">Champion</h2>
                </div>
                <Card padding="md" className="border-warning-200 bg-warning-50/50 text-center">
                  <Trophy className="w-8 h-8 text-warning-500 mx-auto mb-2" />
                  <p className="text-title-lg font-bold text-on-surface">{playerName(champion)}</p>
                  {champion.username && <p className="text-label-sm text-text-secondary">@{champion.username}</p>}
                  <Badge status="winner" className="mt-2">winner</Badge>
                </Card>
              </div>
            )}
          </div>
        </div>
      )}
    </div>
  )
}

function MatchCard({ match, participants, onClick }: { match: Match; participants: Registration[]; onClick: () => void }) {
  const playerA = participants.find(p => p.user_id === match.player_a_id)
  const playerB = participants.find(p => p.user_id === match.player_b_id)
  const winnerA = match.winner_id && match.winner_id === match.player_a_id
  const winnerB = match.winner_id && match.winner_id === match.player_b_id

  return (
    <button onClick={onClick} className="text-left">
      <Card padding="sm" className={match.status === 'completed' ? 'border-success-200' : ''}>
        <div className="space-y-2">
          <PlayerSlot player={playerA} isWinner={!!winnerA} score={winnerA ? match.score : undefined} fallback={match.player_a_id ? 'Player pending profile' : 'TBD'} />
          <div className="border-t border-border/50" />
          <PlayerSlot player={playerB} isWinner={!!winnerB} score={winnerB ? match.score : undefined} fallback={match.player_b_id ? 'Player pending profile' : 'TBD'} />
        </div>
        <div className="mt-2 flex justify-center">
          <Badge status={match.status === 'completed' ? 'completed' : 'pending'}>{match.status.replace(/_/g, ' ')}</Badge>
        </div>
      </Card>
    </button>
  )
}

function PlayerSlot({ player, isWinner, score, fallback }: { player?: Registration; isWinner: boolean; score?: string | null; fallback: string }) {
  return (
    <div className={`flex items-center justify-between gap-2 p-2 rounded-lg text-body-sm ${isWinner ? 'bg-success-50 text-success-700 font-medium' : 'text-on-surface'}`}>
      <div className="min-w-0">
        <p className="truncate">{player ? playerName(player) : fallback}</p>
        {player?.username && <p className="text-label-sm text-text-secondary truncate">@{player.username}</p>}
      </div>
      {score && <span className="text-label-sm font-mono shrink-0">{score}</span>}
    </div>
  )
}

function playerName(player: Registration) {
  return player.display_name || player.username || 'Player'
}

function roundLabel(round: number, totalRounds: number) {
  if (round === totalRounds) return 'Final'
  if (round === totalRounds - 1) return 'Semifinals'
  if (round === totalRounds - 2) return 'Quarterfinals'
  return `Round ${round}`
}
