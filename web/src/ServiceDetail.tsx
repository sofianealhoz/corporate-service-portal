import { useCallback, useEffect, useState } from 'react'
import { Box, Button, Container, Heading, Stack, Text } from '@chakra-ui/react'
import { Link, useParams } from 'react-router-dom'
import { ApiError, getService } from './api'
import type { Service } from './types'
import { surface, tierStyles } from './theme'
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

  if (loading) {
    return (
      <Container maxW="4xl" py="16">
        <LoadingState count={1} />
      </Container>
    )
  }

  if (notFound) {
    return (
      <Container maxW="4xl" py="16">
        <BackLink />
        <EmptyState
          title="Service introuvable"
          hint={`Aucun service publié ne correspond à « ${slug} ».`}
        />
      </Container>
    )
  }

  if (error) {
    return (
      <Container maxW="4xl" py="16">
        <BackLink />
        <ErrorState message={error.message} onRetry={load} />
      </Container>
    )
  }

  if (!item) return null

  return (
    <Box minH="100vh">
      <Box bg={surface.hero} color="white">
        <Container maxW="4xl" py={{ base: '10', md: '16' }}>
          <Button
            asChild
            size="sm"
            variant="ghost"
            color="#cbd5e1"
            mb="6"
            px="0"
            _hover={{ color: 'white', bg: 'transparent' }}
          >
            <Link to="/">← Retour au catalogue</Link>
          </Button>

          <Stack gap="4">
            <TierBadge tier={item.tier} />
            <Heading
              fontSize={{ base: '3xl', md: '5xl' }}
              lineHeight="1.1"
              letterSpacing="-0.03em"
              fontWeight="800"
            >
              {item.title}
            </Heading>
          </Stack>
        </Container>
      </Box>

      <Container maxW="5xl" py={{ base: '8', md: '14' }}>
        <Stack direction={{ base: 'column', md: 'row' }} gap="10" align="start">
          <Stack gap="4" flex="1">
            <Heading size="lg" color={surface.ink} letterSpacing="-0.02em">
              La prestation
            </Heading>
            <Text color={surface.muted} fontSize="lg" lineHeight="1.9">
              {item.description}
            </Text>

            <Stack gap="3" pt="4">
              <Detail label="Gamme" value={tierStyles[item.tier].label} />
              <Detail label="Durée estimée" value={`${item.duration_h} heures`} />
              <Detail label="Référence" value={item.slug} />
            </Stack>
          </Stack>

          <Box
            w={{ base: 'full', md: '320px' }}
            flexShrink="0"
            bg={surface.card}
            borderWidth="1px"
            borderColor={surface.border}
            borderRadius="16px"
            boxShadow={surface.shadow}
            overflow="hidden"
          >
            <Box h="4px" bg={tierStyles[item.tier].accent} />
            <Stack gap="1" p="7">
              <Text fontSize="sm" color={surface.muted}>
                Tarif de la prestation
              </Text>
              <Price cents={item.price_cents} size="lg" />
              <Text fontSize="sm" color={surface.muted} pt="3" lineHeight="1.7">
                Prix ferme, {item.duration_h} heures d'intervention comprises.
              </Text>
            </Stack>
          </Box>
        </Stack>
      </Container>
    </Box>
  )
}

function Detail({ label, value }: { label: string; value: string }) {
  return (
    <Stack direction="row" justify="space-between" borderBottomWidth="1px" borderColor={surface.border} pb="2">
      <Text fontSize="sm" color={surface.muted}>
        {label}
      </Text>
      <Text fontSize="sm" fontWeight="600" color={surface.ink}>
        {value}
      </Text>
    </Stack>
  )
}

function BackLink() {
  return (
    <Button asChild size="sm" variant="ghost" mb="6" px="0">
      <Link to="/">← Retour au catalogue</Link>
    </Button>
  )
}
