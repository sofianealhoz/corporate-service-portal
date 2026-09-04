import { useCallback, useEffect, useState } from 'react'
import { Box, Button, Container, Heading, SimpleGrid, Stack, Text } from '@chakra-ui/react'
import { Link } from 'react-router-dom'
import { listServices } from './api'
import type { Service } from './types'
import { surface } from './theme'
import ServiceCard from './components/ServiceCard'
import Pagination from './components/Pagination'
import { EmptyState, ErrorState, LoadingState } from './components/States'

const PAGE_SIZE = 6

// borne haute de l'API : suffit pour compter un catalogue de cette taille
const COUNT_LIMIT = 100

export default function Catalogue() {
  const [services, setServices] = useState<Service[]>([])
  const [total, setTotal] = useState(0)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [page, setPage] = useState(0)
  const [showDrafts, setShowDrafts] = useState(false)

  const load = useCallback(() => {
    setLoading(true)
    setError(null)

    // published=false demande au serveur de ne pas filtrer, donc d'inclure les
    // brouillons ; published=true est le comportement public par défaut
    const published = !showDrafts

    Promise.all([
      listServices({ published, limit: PAGE_SIZE, offset: page * PAGE_SIZE }),
      // le listing rend le nombre d'éléments de la page, pas un total : on le
      // demande à part pour afficher le nombre de pages
      listServices({ published, limit: COUNT_LIMIT }),
    ])
      .then(([current, all]) => {
        setServices(current.items)
        setTotal(all.count)
      })
      .catch((err: Error) => setError(err.message))
      .finally(() => setLoading(false))
  }, [page, showDrafts])

  useEffect(load, [load])

  const pageCount = Math.max(1, Math.ceil(total / PAGE_SIZE))

  return (
    <Box minH="100vh">
      <Box bg={surface.hero} color="white">
        <Container maxW="6xl" py={{ base: '14', md: '20' }}>
          <Stack gap="5" maxW="2xl">
            <Text
              fontSize="xs"
              fontWeight="700"
              letterSpacing="0.18em"
              textTransform="uppercase"
              color="#a5b4fc"
            >
              Catalogue de services
            </Text>

            <Heading
              fontSize={{ base: '3xl', md: '5xl' }}
              lineHeight="1.1"
              letterSpacing="-0.03em"
              fontWeight="800"
            >
              Des prestations claires, des tarifs affichés.
            </Heading>

            <Text fontSize={{ base: 'md', md: 'lg' }} color="#cbd5e1" lineHeight="1.7">
              Chaque prestation est décrite, chiffrée et estimée en durée. Pas de devis
              opaque, pas de fourchette : le prix est celui affiché.
            </Text>

            <Stack direction="row" gap="10" pt="4">
              <Stat value={String(total)} label={total > 1 ? 'prestations' : 'prestation'} />
              <Stat value="3" label="gammes" />
              <Stat value="48 h" label="délai de réponse" />
            </Stack>
          </Stack>
        </Container>
      </Box>

      <Container maxW="6xl" py={{ base: '8', md: '12' }}>
        <Stack
          direction={{ base: 'column', sm: 'row' }}
          justify="space-between"
          align={{ base: 'stretch', sm: 'center' }}
          gap="3"
          mb="8"
        >
          <Text color={surface.muted} fontSize="sm">
            {services.length > 0
              ? `Affichés ${page * PAGE_SIZE + 1} à ${page * PAGE_SIZE + services.length} sur ${total}`
              : `${total} service${total > 1 ? 's' : ''}`}
          </Text>

          <Button
            size="sm"
            borderRadius="full"
            px="5"
            variant={showDrafts ? 'solid' : 'outline'}
            onClick={() => {
              setShowDrafts((v) => !v)
              setPage(0) // le filtre change, les numéros de page ne valent plus rien
            }}
          >
            {showDrafts ? 'Masquer les brouillons' : 'Afficher les brouillons'}
          </Button>
        </Stack>

        {loading && <LoadingState count={PAGE_SIZE} />}
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
                  // l'API refuse le détail d'un brouillon : ne pas proposer un
                  // lien qui mènerait à un 404
                  action={
                    s.published ? (
                      <Button
                        asChild
                        size="sm"
                        variant="outline"
                        borderRadius="full"
                        mt="1"
                        w="full"
                      >
                        <Link to={`/services/${s.slug}`}>Voir la fiche</Link>
                      </Button>
                    ) : (
                      <Text fontSize="xs" color={surface.muted} mt="1" textAlign="center">
                        Fiche publiée une fois le service publié
                      </Text>
                    )
                  }
                />
              ))}
            </SimpleGrid>

            <Pagination
              page={page}
              pageCount={pageCount}
              onChange={setPage}
              disabled={loading}
            />
          </>
        )}
      </Container>
    </Box>
  )
}

function Stat({ value, label }: { value: string; label: string }) {
  return (
    <Stack gap="0">
      <Text fontSize="3xl" fontWeight="800" letterSpacing="-0.02em" lineHeight="1.1">
        {value}
      </Text>
      <Text fontSize="sm" color="#94a3b8">
        {label}
      </Text>
    </Stack>
  )
}
