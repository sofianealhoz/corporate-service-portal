import { Badge, Box, Heading, Stack, Text } from '@chakra-ui/react'
import type { ReactNode } from 'react'
import type { Service } from '../types'
import TierBadge from './TierBadge'
import Price from './Price'

interface ServiceCardProps {
  service: Service
  // le composant n'a pas d'avis sur la navigation : l'écran fournit le lien
  action?: ReactNode
}

// Reçoit le service à afficher, ne va jamais le chercher lui-même. C'est ce qui
// permet de le réutiliser dans le catalogue, une recherche ou une future page
// d'administration sans le modifier.
export default function ServiceCard({ service, action }: ServiceCardProps) {
  return (
    <Box borderWidth="1px" rounded="lg" p="5" h="full">
      <Stack gap="2" h="full">
        <Stack direction="row" gap="2">
          <TierBadge tier={service.tier} />
          {!service.published && <Badge colorPalette="orange">Brouillon</Badge>}
        </Stack>

        <Heading size="md">{service.title}</Heading>
        <Text color="gray.600" lineClamp={3}>
          {service.description}
        </Text>

        <Box flex="1" />

        <Price cents={service.price_cents} />
        <Text fontSize="sm" color="gray.500">
          {service.duration_h} h
        </Text>
        {action}
      </Stack>
    </Box>
  )
}
