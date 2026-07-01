import { useEffect, useState, type FormEvent } from 'react'
import { useNavigate } from 'react-router-dom'
import { Card } from '@/components/ui/Card'
import { Input } from '@/components/ui/Input'
import { Button } from '@/components/ui/Button'
import { api, type Profile } from '@/lib/api'

const COUNTRIES = [
  { code: 'ET', name: 'Ethiopia' },
  { code: 'US', name: 'United States' },
  { code: 'GB', name: 'United Kingdom' },
  { code: 'KE', name: 'Kenya' },
  { code: 'NG', name: 'Nigeria' },
  { code: 'EG', name: 'Egypt' },
  { code: 'ZA', name: 'South Africa' },
  { code: 'DE', name: 'Germany' },
  { code: 'FR', name: 'France' },
  { code: 'CA', name: 'Canada' },
  { code: 'IN', name: 'India' },
  { code: 'CN', name: 'China' },
  { code: 'JP', name: 'Japan' },
  { code: 'BR', name: 'Brazil' },
  { code: 'AU', name: 'Australia' },
]

export function EditProfile() {
  const navigate = useNavigate()
  const [loading, setLoading] = useState(true)
  const [submitting, setSubmitting] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [form, setForm] = useState({
    username: '',
    display_name: '',
    avatar_url: '',
    bio: '',
    country: '',
    country_code: '',
  })

  useEffect(() => {
    let cancelled = false
    api.get<Profile>('/api/v1/users/me')
      .then(p => {
        if (cancelled) return
        setForm({
          username: p.username || '',
          display_name: p.display_name || '',
          avatar_url: p.avatar_url || '',
          bio: p.bio || '',
          country: p.country || '',
          country_code: p.country_code || '',
        })
      })
      .catch(err => { if (!cancelled) setError(err?.message || 'Failed to load profile') })
      .finally(() => { if (!cancelled) setLoading(false) })
    return () => { cancelled = true }
  }, [])

  async function onSubmit(e: FormEvent) {
    e.preventDefault()
    setError(null)
    setSubmitting(true)
    try {
      const body: Record<string, string | null> = {}
      if (form.username) body.username = form.username
      body.display_name = form.display_name || null
      body.avatar_url = form.avatar_url || null
      body.bio = form.bio || null
      body.country = form.country || null
      body.country_code = form.country_code || null
      await api.patch('/api/v1/users/me', body)
      navigate('/profile')
    } catch (err: any) {
      setError(err?.message || 'Failed to save profile')
    } finally {
      setSubmitting(false)
    }
  }

  function onCountryChange(code: string) {
    const c = COUNTRIES.find(c => c.code === code)
    setForm(f => ({ ...f, country_code: code, country: c?.name || '' }))
  }

  if (loading) return <div className="p-12 text-center text-text-secondary">Loading…</div>

  return (
    <div className="max-w-2xl mx-auto">
      <h1 className="text-headline-md text-on-surface mb-6">Edit profile</h1>
      <Card padding="lg">
        <form onSubmit={onSubmit} className="space-y-5">
          <Input
            label="Username"
            required
            value={form.username}
            onChange={e => setForm(f => ({ ...f, username: e.target.value }))}
            placeholder="your_username"
          />
          <Input
            label="Display name"
            value={form.display_name}
            onChange={e => setForm(f => ({ ...f, display_name: e.target.value }))}
            placeholder="Shown to other players"
          />
          <Input
            label="Avatar URL"
            type="url"
            value={form.avatar_url}
            onChange={e => setForm(f => ({ ...f, avatar_url: e.target.value }))}
            placeholder="https://…"
          />
          <div>
            <label className="block text-label-md text-on-surface mb-2">Bio</label>
            <textarea
              className="w-full px-3 py-2 border border-border rounded-lg focus:ring-2 focus:ring-primary focus:border-transparent outline-none min-h-[100px]"
              value={form.bio}
              onChange={e => setForm(f => ({ ...f, bio: e.target.value }))}
              placeholder="Tell other players about yourself"
            />
          </div>
          <div>
            <label className="block text-label-md text-on-surface mb-2">Country</label>
            <select
              className="w-full px-3 py-2 border border-border rounded-lg focus:ring-2 focus:ring-primary focus:border-transparent outline-none h-10"
              value={form.country_code}
              onChange={e => onCountryChange(e.target.value)}
            >
              <option value="">Select country</option>
              {COUNTRIES.map(c => (
                <option key={c.code} value={c.code}>{c.name}</option>
              ))}
            </select>
          </div>

          {error && (
            <p className="text-body-sm text-danger" role="alert">{error}</p>
          )}

          <div className="flex gap-3">
            <Button type="submit" loading={submitting} disabled={submitting}>Save changes</Button>
            <Button type="button" variant="outline" onClick={() => navigate('/profile')}>Cancel</Button>
          </div>
        </form>
      </Card>
    </div>
  )
}
