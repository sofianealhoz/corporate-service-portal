// L'API stocke les prix en centimes, jamais en flottant. La conversion vit ici
// et nulle part ailleurs : un seul endroit du front a le droit de diviser par 100.
export function formatPrice(cents: number): string {
  return (cents / 100).toLocaleString('fr-FR', {
    style: 'currency',
    currency: 'EUR',
    // « 12 500 € » plutôt que « 12 500,00 € » : les centimes ne sont affichés
    // que s'il y en a
    minimumFractionDigits: cents % 100 === 0 ? 0 : 2,
  })
    // fr-FR sépare les milliers par une espace fine insécable, que beaucoup de
    // polices ne dessinent pas : « 31000 » au lieu de « 31 000 »
    .replace(/\u202f/g, '\u00a0')
}
