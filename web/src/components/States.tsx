import { Box, Button, Heading, SimpleGrid, Stack, Text } from '@chakra-ui/react'
import { surface } from '../theme'

// Les trois états qu'un écran branché sur une API doit savoir montrer.
// Aucun d'eux ne connaît l'API : ils reçoivent un texte et, au besoin, une
// action à déclencher.

interface LoadingStateProps {
  count?: number
}

// Des silhouettes aux dimensions des vraies cartes plutôt qu'un compteur qui
// tourne : la page ne saute pas quand les données arrivent.
export function LoadingState({ count = 6 }: LoadingStateProps) {
  return (
    <SimpleGrid columns={{ base: 1, md: 2, lg: 3 }} gap="6">
      {Array.from({ length: count }, (_, i) => (
        <Box
          key={i}
          h="248px"
          bg={surface.card}
          borderWidth="1px"
          borderColor={surface.border}
          borderRadius="16px"
          opacity="0.7"
        />
      ))}
    </SimpleGrid>
  )
}

interface EmptyStateProps {
  title: string
  hint?: string
}

export function EmptyState({ title, hint }: EmptyStateProps) {
  return (
    <Box
      borderWidth="1px"
      borderStyle="dashed"
      borderColor={surface.border}
      borderRadius="16px"
      bg={surface.card}
      p="12"
      textAlign="center"
    >
      <Heading size="sm" mb="2" color={surface.ink}>
        {title}
      </Heading>
      {hint && (
        <Text color={surface.muted} fontSize="sm">
          {hint}
        </Text>
      )}
    </Box>
  )
}

interface ErrorStateProps {
  title?: string
  message: string
  onRetry?: () => void
}

export function ErrorState({ title = 'Chargement impossible', message, onRetry }: ErrorStateProps) {
  return (
    <Box borderWidth="1px" borderColor="#fecaca" bg="#fef2f2" borderRadius="16px" p="6">
      <Stack gap="3" align="start">
        <Heading size="sm" color="#b91c1c">
          {title}
        </Heading>
        <Text color="#7f1d1d" fontSize="sm">
          {message}
        </Text>
        {onRetry && (
          <Button size="sm" variant="outline" borderRadius="full" onClick={onRetry}>
            Réessayer
          </Button>
        )}
      </Stack>
    </Box>
  )
}
