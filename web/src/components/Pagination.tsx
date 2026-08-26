import { Button, Stack, Text } from '@chakra-ui/react'

interface PaginationProps {
  page: number // 0 pour la première
  hasNext: boolean
  onChange: (page: number) => void
  disabled?: boolean
}

// Ne connaît ni l'API ni les données : on lui dit où on est et s'il existe une
// suite, il annonce la page demandée.
export default function Pagination({ page, hasNext, onChange, disabled }: PaginationProps) {
  return (
    <Stack direction="row" gap="3" align="center" justify="center" pt="8">
      <Button
        size="sm"
        variant="outline"
        disabled={disabled || page === 0}
        onClick={() => onChange(page - 1)}
      >
        Précédent
      </Button>

      <Text fontSize="sm" color="gray.500">
        Page {page + 1}
      </Text>

      <Button size="sm" variant="outline" disabled={disabled || !hasNext} onClick={() => onChange(page + 1)}>
        Suivant
      </Button>
    </Stack>
  )
}
