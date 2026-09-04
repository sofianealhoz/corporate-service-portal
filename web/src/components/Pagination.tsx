import { Button, Stack, Text } from '@chakra-ui/react'
import { surface } from '../theme'

interface PaginationProps {
  page: number // 0 pour la première
  pageCount: number
  onChange: (page: number) => void
  disabled?: boolean
}

// Ne connaît ni l'API ni les données : on lui dit où on est et combien il y a
// de pages, il annonce la page demandée.
export default function Pagination({ page, pageCount, onChange, disabled }: PaginationProps) {
  if (pageCount <= 1) return null

  return (
    <Stack direction="row" gap="4" align="center" justify="center" pt="10">
      <Button
        size="sm"
        variant="outline"
        borderRadius="full"
        px="5"
        disabled={disabled || page === 0}
        onClick={() => onChange(page - 1)}
      >
        Précédent
      </Button>

      <Text fontSize="sm" color={surface.muted} minW="24" textAlign="center">
        Page {page + 1} sur {pageCount}
      </Text>

      <Button
        size="sm"
        variant="outline"
        borderRadius="full"
        px="5"
        disabled={disabled || page >= pageCount - 1}
        onClick={() => onChange(page + 1)}
      >
        Suivant
      </Button>
    </Stack>
  )
}
