import { useCallback, useEffect, useState } from 'react'
import { Button, Container, Heading, Separator, Stack, Text } from '@chakra-ui/react'
import { Link, useParams } from 'react-router-dom'
import { ApiError, getService } from './api'
import type { Service } from './types'
import TierBadge from './components/TierBadge'
import Price from './components/Price'
import { EmptyState, ErrorState, LoadingState } from './components/States'

export default function ServiceDetail() {
  const { slug = '' } = useParams()
  const [item, setItem] = useState<Service | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<Error | null>(null)

  const load = useCallback(() => {
    setLoading(true)
    setError(null)

    getService(slug)
      .then(setItem)
      .catch((err: Error) => setError(err))
      .finally(() => setLoading(false))
  }, [slug])

  useEffect(load, [load])

  // l'API répond 404 aussi bien pour un slug inconnu que pour un brouillon :
  // de l'extérieur les deux cas sont indistinguables, et c'est voulu
  const notFound = error instanceof ApiError && error.status === 404

  return (
    <Container maxW="3xl" py="10">
      <Button asChild size="sm" variant="ghost" mb="6">
        <Link to="/">Retour au catalogue</Link>
      </Button>

      {loading && <LoadingState label="Chargement de la fiche..." />}

      {!loading && notFound && (
        <EmptyState
          title="Service introuvable"
          hint={`Aucun service publié ne correspond à « ${slug} ».`}
        />
      )}

      {!loading && error && !notFound && <ErrorState message={error.message} onRetry={load} />}

      {!loading && !error && item && (
        <Stack gap="4">
          <TierBadge tier={item.tier} />
          <Heading size="2xl">{item.title}</Heading>
          <Text color="gray.600">{item.description}</Text>

          <Separator my="2" />

          <Stack direction="row" gap="10">
            <Stack gap="0">
              <Text fontSize="sm" color="gray.500">
                Tarif
              </Text>
              <Price cents={item.price_cents} size="lg" />
            </Stack>
            <Stack gap="0">
              <Text fontSize="sm" color="gray.500">
                Durée
              </Text>
              <Text fontSize="2xl" fontWeight="semibold">
                {item.duration_h} h
              </Text>
            </Stack>
          </Stack>
        </Stack>
      )}
    </Container>
  )
}
