import { Box, Button, Heading, Spinner, Stack, Text } from '@chakra-ui/react'

// Les trois états qu'un écran branché sur une API doit savoir montrer.
// Aucun d'eux ne connaît l'API : ils reçoivent un texte et, au besoin, une
// action à déclencher.

interface LoadingStateProps {
  label?: string
}

export function LoadingState({ label = 'Chargement...' }: LoadingStateProps) {
  return (
    <Stack direction="row" gap="3" align="center" py="10">
      <Spinner />
      <Text color="gray.500">{label}</Text>
    </Stack>
  )
}

interface EmptyStateProps {
  title: string
  hint?: string
}

export function EmptyState({ title, hint }: EmptyStateProps) {
  return (
    <Box borderWidth="1px" borderStyle="dashed" rounded="lg" p="10" textAlign="center">
      <Heading size="sm" mb="1">
        {title}
      </Heading>
      {hint && (
        <Text color="gray.500" fontSize="sm">
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
    <Box borderWidth="1px" borderColor="red.200" rounded="lg" p="6">
      <Heading size="sm" color="red.600" mb="1">
        {title}
      </Heading>
      <Text color="gray.600" fontSize="sm" mb={onRetry ? '4' : '0'}>
        {message}
      </Text>
      {onRetry && (
        <Button size="sm" variant="outline" onClick={onRetry}>
          Réessayer
        </Button>
      )}
    </Box>
  )
}
