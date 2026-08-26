import { Badge } from '@chakra-ui/react'
import type { Tier } from '../types'

// libellés et couleurs par gamme, alignés sur le CHECK en base
const tiers: Record<Tier, { label: string; palette: string }> = {
  standard: { label: 'Standard', palette: 'gray' },
  premium: { label: 'Premium', palette: 'blue' },
  enterprise: { label: 'Entreprise', palette: 'purple' },
}

interface TierBadgeProps {
  tier: Tier
}

export default function TierBadge({ tier }: TierBadgeProps) {
  const { label, palette } = tiers[tier]
  return (
    <Badge colorPalette={palette} alignSelf="start">
      {label}
    </Badge>
  )
}
