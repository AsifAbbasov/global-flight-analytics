export type GlobeLayerType =
  | 'geography'
  | 'countries'
  | 'capitals'
  | 'airports'
  | 'flights'
  | 'air-corridors'

export interface GlobeLayerConfig {
  type: GlobeLayerType
  label: string
  description: string
  enabledByDefault: boolean
}

export type GlobeLayerVisibility = Record<GlobeLayerType, boolean>

export const GLOBE_LAYERS: GlobeLayerConfig[] = [
  {
    type: 'geography',
    label: 'Geography',
    description: 'Continents, oceans, mountains and physical geography.',
    enabledByDefault: true,
  },
  {
    type: 'countries',
    label: 'Countries',
    description: 'Country borders and political regions.',
    enabledByDefault: false,
  },
  {
    type: 'capitals',
    label: 'Capitals',
    description: 'Capital cities and major urban centers.',
    enabledByDefault: false,
  },
  {
    type: 'airports',
    label: 'Airports',
    description: 'Major airports and aviation infrastructure.',
    enabledByDefault: true,
  },
  {
    type: 'flights',
    label: 'Flights',
    description: 'Current aircraft positions and live traffic.',
    enabledByDefault: true,
  },
  {
    type: 'air-corridors',
    label: 'Air Corridors',
    description: 'Major aviation routes and traffic flows.',
    enabledByDefault: false,
  },
]
