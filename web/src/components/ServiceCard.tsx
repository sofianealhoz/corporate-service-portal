import { Box, Heading, Stack, Text } from '@chakra-ui/react'
import type { ReactNode } from 'react'
import type { Service } from '../types'
import { surface, tierStyles } from '../theme'
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
    <Box
      display="flex"
      flexDirection="column"
      h="full"
      bg={surface.card}
      borderWidth="1px"
      borderColor={surface.border}
      borderRadius="16px"
      overflow="hidden"
      boxShadow={surface.shadow}
      transition="transform .18s ease, box-shadow .18s ease"
      _hover={{ transform: 'translateY(-4px)', boxShadow: surface.shadowHover }}
    >
      {/* filet de couleur : la gamme se lit avant même de lire le texte */}
      <Box h="4px" bg={tierStyles[service.tier].accent} />

      <Stack gap="3" p="6" flex="1">
        <Stack direction="row" gap="2" align="center">
          <TierBadge tier={service.tier} />
          {!service.published && (
            <Box
              as="span"
              px="2.5"
              py="1"
              borderRadius="full"
              bg="#fff7ed"
              color="#c2410c"
              borderWidth="1px"
              borderColor="#fed7aa"
              fontSize="xs"
              fontWeight="600"
            >
              Brouillon
            </Box>
          )}
        </Stack>

        <Heading size="md" color={surface.ink} lineHeight="1.3">
          {service.title}
        </Heading>

        <Text color={surface.muted} fontSize="sm" lineHeight="1.6" lineClamp={3}>
          {service.description}
        </Text>

        <Box flex="1" />

        <Stack
          direction="row"
          justify="space-between"
          align="baseline"
          pt="3"
          borderTopWidth="1px"
          borderColor={surface.border}
        >
          <Price cents={service.price_cents} />
          <Text fontSize="sm" color={surface.muted}>
            {service.duration_h} h
          </Text>
        </Stack>

        {action}
      </Stack>
    </Box>
  )
}
