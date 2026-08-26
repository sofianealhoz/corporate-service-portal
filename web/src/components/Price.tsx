import { Text } from '@chakra-ui/react'
import { formatPrice } from '../format'

interface PriceProps {
  cents: number
  size?: 'sm' | 'md' | 'lg'
}

// Reçoit un montant en centimes et l'affiche. Ne sait rien d'où il vient.
export default function Price({ cents, size = 'md' }: PriceProps) {
  return (
    <Text fontWeight="semibold" fontSize={size === 'lg' ? '2xl' : size}>
      {formatPrice(cents)}
    </Text>
  )
}
