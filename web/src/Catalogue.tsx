import { useCallback, useEffect, useState } from 'react'
import { Button, Container, Heading, SimpleGrid, Stack, Text } from '@chakra-ui/react'
import { Link } from 'react-router-dom'
import { listServices } from './api'
import type { Service } from './types'
import ServiceCard from './components/ServiceCard'
import Pagination from './components/Pagination'
import { EmptyState, ErrorState, LoadingState } from './components/States'

const PAGE_SIZE = 6

export default function Catalogue() {
  const [services, setServices] = useState<Service[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [page, setPage] = useState(0)
  const [showDrafts, setShowDrafts] = useState(false)

  const load = useCallback(() => {
    setLoading(true)
    setError(null)

    // published=false demande au serveur de ne pas filtrer, donc d'inclure les
    // brouillons ; published=true est le comportement public par défaut
    listServices({ published: !showDrafts, limit: PAGE_SIZE, offset: page * PAGE_SIZE })
      .then((res) => setServices(res.items))
      .catch((err: Error) => setError(err.message))
      .finally(() => setLoading(false))
  }, [page, showDrafts])

  useEffect(load, [load])

  // le listing renvoie le nombre d'éléments de la page, pas un total : une page
  // pleine est le seul indice qu'il reste peut-être une suite
  const hasNext = services.length === PAGE_SIZE

  return (
    <Container maxW="5xl" py="10">
      <Heading size="xl" mb="2">
        Catalogue de services
      </Heading>
      <Text color="gray.500" mb="6">
        Les prestations proposees par le portail.
      </Text>

      <Stack direction="row" justify="space-between" align="center" mb="6">
        <Button
          size="sm"
          variant={showDrafts ? 'solid' : 'outline'}
          onClick={() => {
            setShowDrafts((v) => !v)
            setPage(0) // le filtre change, les numéros de page ne valent plus rien
          }}
        >
          {showDrafts ? 'Masquer les brouillons' : 'Afficher les brouillons'}
        </Button>
        <Text fontSize="sm" color="gray.500">
          {services.length} service{services.length > 1 ? 's' : ''} sur cette page
        </Text>
      </Stack>

      {loading && <LoadingState />}
      {!loading && error && <ErrorState message={error} onRetry={load} />}

      {!loading && !error && services.length === 0 && (
        <EmptyState
          title="Aucun service sur cette page"
          hint={page > 0 ? 'Revenez à la page précédente.' : 'Le catalogue est vide.'}
        />
      )}

      {!loading && !error && services.length > 0 && (
        <>
          <SimpleGrid columns={{ base: 1, md: 2, lg: 3 }} gap="6">
            {services.map((s) => (
              <ServiceCard
                key={s.id}
                service={s}
                action={
                  <Button asChild size="sm" variant="outline" mt="2">
                    <Link to={`/services/${s.slug}`}>Voir la fiche</Link>
                  </Button>
                }
              />
            ))}
          </SimpleGrid>

          <Pagination page={page} hasNext={hasNext} onChange={setPage} disabled={loading} />
        </>
      )}
    </Container>
  )
}
