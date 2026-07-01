import { supabase } from './supabase'

const BASE = import.meta.env.VITE_API_BASE_URL || ''

export class ApiError extends Error {
  status: number
  constructor(status: number, message: string) {
    super(message)
    this.status = status
  }
}

async function request<T>(path: string, init: RequestInit = {}): Promise<T> {
  const { data } = await supabase.auth.getSession()
  const token = data.session?.access_token

  const headers = new Headers(init.headers)
  if (token) headers.set('Authorization', `Bearer ${token}`)
  if (init.body && !headers.has('Content-Type')) {
    headers.set('Content-Type', 'application/json')
  }

  const res = await fetch(`${BASE}${path}`, { ...init, headers })
  const text = await res.text()
  let payload: any = null
  if (text) {
    try { payload = JSON.parse(text) } catch { payload = text }
  }

  if (!res.ok) {
    const msg = payload?.message || payload?.error || res.statusText
    throw new ApiError(res.status, msg)
  }

  if (payload && typeof payload === 'object' && 'data' in payload) {
    return payload.data as T
  }
  return payload as T
}

export const api = {
  get:    <T>(path: string)            => request<T>(path),
  post:   <T>(path: string, body?: any) => request<T>(path, { method: 'POST', body: body ? JSON.stringify(body) : undefined }),
  patch:  <T>(path: string, body?: any) => request<T>(path, { method: 'PATCH', body: body ? JSON.stringify(body) : undefined }),
  delete: <T>(path: string)            => request<T>(path, { method: 'DELETE' }),
}

export interface Tournament {
  id: string
  title: string
  game: string
  description?: string | null
  rules?: string | null
  format: string
  status: string
  max_participants: number
  min_participants: number
  team_size: number
  entry_fee: number
  prize_pool: number
  currency: string
  location: string
  best_of: number
  platform_fee_pct: number
  organizer_fee_pct: number
  start_date: string
  end_date?: string | null
  registration_deadline?: string | null
  registration_close_at?: string | null
  organizer_id: string
  banner_url?: string | null
  created_at: string
  updated_at: string
}

export interface FeeBreakdown {
  paid_participants: number
  collected: number
  platform_cut: number
  organizer_cut: number
  winner_prize: number
}

export function computeFeeBreakdown(t: { entry_fee: number; platform_fee_pct: number; organizer_fee_pct: number }, paidCount: number): FeeBreakdown {
  const collected = (t.entry_fee || 0) * paidCount
  const platform_cut = Math.floor(collected * (t.platform_fee_pct || 5) / 100)
  const organizer_cut = Math.floor(collected * (t.organizer_fee_pct || 0) / 100)
  const winner_prize = Math.max(0, collected - platform_cut - organizer_cut)
  return { paid_participants: paidCount, collected, platform_cut, organizer_cut, winner_prize }
}

export interface Profile {
  id: string
  username: string
  display_name?: string | null
  email?: string | null
  avatar_url?: string | null
  role: string
  bio?: string | null
  country?: string | null
  country_code?: string | null
  created_at: string
  updated_at: string
}

export interface ProfileStats {
  tournaments_played: number
  tournaments_hosted: number
  wins: number
}

export interface ProfileWithStats extends Profile {
  stats: ProfileStats
}
