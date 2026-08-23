// FRONTEND_AVIATION_BASEMAP_V1
// Zero-budget production map contract for the GFA frontend.
// OpenFreeMap public styles are used without API keys or billing.

export const aviationBasemap = {
  provider: 'OpenFreeMap',
  styleURL: 'https://tiles.openfreemap.org/styles/liberty',
  attribution: 'OpenFreeMap © OpenMapTiles · Data from OpenStreetMap',
  requiresAPIKey: false,
  requiresBillingAccount: false,
} as const

export const aviationMapView = {
  worldCenter: [0, 20] as [number, number],
  worldZoom: 0.8,
  maxRegionalZoom: 7,
  maxEvidenceZoom: 9,
} as const
