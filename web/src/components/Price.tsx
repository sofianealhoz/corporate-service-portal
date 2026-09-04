import { Text } from '@chakra-ui/react'
import { formatPrice } from '../format'
import { surface } from '../theme'

interface PriceProps {
  cents: number
  size?: 'md' | 'lg'
}

// Reçoit un montant en centimes et l'affiche. Ne sait rien d'où il vient.
export default function Price({ cents, size = 'md' }: PriceProps) {
  return (
    <Text
      fontWeight="700"
      color={surface.ink}
      fontSize={size === 'lg' ? '3xl' : 'lg'}
      letterSpacing="-0.02em"
    >
      {formatPrice(cents)}
    </Text>
  )
}
