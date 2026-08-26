// L'API stocke les prix en centimes, jamais en flottant. La conversion vit ici
// et nulle part ailleurs : un seul endroit du front a le droit de diviser par 100.
export function formatPrice(cents: number): string {
  return (cents / 100).toLocaleString('fr-FR', {
    style: 'currency',
    currency: 'EUR',
  })
}
