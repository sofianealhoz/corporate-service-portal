// Un seul endroit pour les couleurs de gamme, partagé par le badge, la carte et
// la fiche : trois composants qui doivent rester d'accord.
import type { Tier } from './types'

export interface TierStyle {
  label: string
  accent: string
  bg: string
  fg: string
  border: string
}

export const tierStyles: Record<Tier, TierStyle> = {
  standard: {
    label: 'Standard',
    accent: '#64748b',
    bg: '#f1f5f9',
    fg: '#475569',
    border: '#e2e8f0',
  },
  premium: {
    label: 'Premium',
    accent: '#6366f1',
    bg: '#eef2ff',
    fg: '#4338ca',
    border: '#e0e7ff',
  },
  enterprise: {
    label: 'Entreprise',
    accent: '#d946ef',
    bg: '#fdf4ff',
    fg: '#a21caf',
    border: '#f5d0fe',
  },
}

export const surface = {
  card: '#ffffff',
  border: '#e6eaf3',
  muted: '#64748b',
  ink: '#0f172a',
  shadow: '0 1px 2px rgba(15, 23, 42, 0.04)',
  shadowHover: '0 14px 32px rgba(15, 23, 42, 0.12)',
  hero: 'linear-gradient(135deg, #0f172a 0%, #1e293b 45%, #312e81 100%)',
}
