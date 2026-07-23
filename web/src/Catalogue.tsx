import { useEffect, useState } from 'react'
import {
  Badge,
  Box,
  Container,
  Heading,
  SimpleGrid,
  Spinner,
  Stack,
  Text,
} from '@chakra-ui/react'
import { listServices } from './api'
import type { Service } from './types'

// l'API stocke le prix en centimes, on l'affiche en euros
function formatPrice(cents: number): string {
  return (cents / 100).toLocaleString('fr-FR', {
    style: 'currency',
    currency: 'EUR',
  })
}

export default function Catalogue() {
  const [services, setServices] = useState<Service[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    listServices()
      .then((res) => setServices(res.items))
      .catch((err) => setError(String(err)))
      .finally(() => setLoading(false))
  }, [])

  return (
    <Container maxW="5xl" py="10">
      <Heading size="xl" mb="2">
        Catalogue de services
      </Heading>
      <Text color="gray.500" mb="8">
        Les prestations proposees par le portail.
      </Text>

      {loading && <Spinner />}
      {error && <Text color="red.500">Chargement impossible : {error}</Text>}
      {!loading && !error && services.length === 0 && (
        <Text color="gray.500">Aucun service pour le moment.</Text>
      )}

      <SimpleGrid columns={{ base: 1, md: 2, lg: 3 }} gap="6">
        {services.map((s) => (
          <Box key={s.id} borderWidth="1px" rounded="lg" p="5">
            <Stack gap="2">
              <Badge alignSelf="start">{s.tier}</Badge>
              <Heading size="md">{s.title}</Heading>
              <Text color="gray.600">{s.description}</Text>
              <Text fontWeight="semibold">{formatPrice(s.price_cents)}</Text>
              <Text fontSize="sm" color="gray.500">
                {s.duration_h} h
              </Text>
            </Stack>
          </Box>
        ))}
      </SimpleGrid>
    </Container>
  )
}
