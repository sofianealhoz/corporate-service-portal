import type { ListResponse, Service } from './types'

// Porte le code HTTP jusqu'aux écrans, qui distinguent ainsi un 404 (la
// ressource n'existe pas, ou n'est pas publiée) d'une panne.
export class ApiError extends Error {
  readonly status: number

  constructor(status: number, message: string) {
    super(message)
    this.name = 'ApiError'
    this.status = status
  }
}

// Un seul endroit qui parle a l'API : les ecrans ne connaissent pas les URLs.
async function getJSON<T>(url: string): Promise<T> {
  const res = await fetch(url)

  if (!res.ok) {
    // l'API renvoie {"error": "..."} ; on retombe sur le statut si le corps
    // n'est pas exploitable
    const message = await res
      .json()
      .then((body) => (body as { error?: string }).error)
      .catch(() => undefined)

    throw new ApiError(res.status, message ?? `${res.status} ${res.statusText}`)
  }

  return res.json() as Promise<T>
}

export interface ListParams {
  published?: boolean
  limit?: number
  offset?: number
}

export function listServices(params: ListParams = {}): Promise<ListResponse> {
  const query = new URLSearchParams()
  if (params.published !== undefined) query.set('published', String(params.published))
  if (params.limit !== undefined) query.set('limit', String(params.limit))
  if (params.offset !== undefined) query.set('offset', String(params.offset))

  const qs = query.toString()
  return getJSON<ListResponse>(`/api/services${qs ? `?${qs}` : ''}`)
}

export function getService(slug: string): Promise<Service> {
  return getJSON<Service>(`/api/services/${encodeURIComponent(slug)}`)
}
