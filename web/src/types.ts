// Miroir du modele Go (internal/service/model.go), a garder synchronise.
export type Tier = 'standard' | 'premium' | 'enterprise'

export interface Service {
  id: number
  slug: string
  title: string
  description: string
  tier: Tier
  duration_h: number
  price_cents: number
  published: boolean
  created_at: string
  updated_at: string
}

export interface ListResponse {
  items: Service[]
  count: number
}
