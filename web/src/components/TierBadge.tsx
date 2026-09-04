import { Box } from '@chakra-ui/react'
import type { Tier } from '../types'
import { tierStyles } from '../theme'

interface TierBadgeProps {
  tier: Tier
}

export default function TierBadge({ tier }: TierBadgeProps) {
  const s = tierStyles[tier]
  return (
    <Box
      as="span"
      display="inline-block"
      alignSelf="start"
      px="2.5"
      py="1"
      borderRadius="full"
      borderWidth="1px"
      borderColor={s.border}
      bg={s.bg}
      color={s.fg}
      fontSize="xs"
      fontWeight="600"
      letterSpacing="0.02em"
    >
      {s.label}
    </Box>
  )
}
