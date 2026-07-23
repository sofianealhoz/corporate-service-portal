import type { ListResponse, Service } from './types'

// Un seul endroit qui parle a l'API : les ecrans ne connaissent pas les URLs.
async function getJSON<T>(url: string): Promise<T> {
  const res = await fetch(url)
  if (!res.ok) {
    throw new Error(`${res.status} ${res.statusText}`)
  }
  return res.json() as Promise<T>
}

export function listServices(): Promise<ListResponse> {
  return getJSON<ListResponse>('/api/services')
}

export function getService(slug: string): Promise<Service> {
  return getJSON<Service>(`/api/services/${slug}`)
}
