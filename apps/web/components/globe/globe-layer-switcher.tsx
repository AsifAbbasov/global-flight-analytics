import { GLOBE_LAYERS, type GlobeLayerVisibility } from './globe-types'

interface GlobeLayerSwitcherProps {
  visibility: GlobeLayerVisibility
  onToggle: (layer: keyof GlobeLayerVisibility) => void
}

export function GlobeLayerSwitcher({
  visibility,
  onToggle,
}: GlobeLayerSwitcherProps) {
  return (
    <div className='rounded-xl border border-slate-800 bg-slate-950/90 p-4 text-white shadow-xl'>
      <h3 className='text-sm font-semibold uppercase tracking-wide text-sky-300'>
        Globe Layers
      </h3>

      <div className='mt-4 space-y-3'>
        {GLOBE_LAYERS.map(layer => (
          <label
            key={layer.type}
            className='flex cursor-pointer items-start gap-3 rounded-lg border border-slate-800 bg-slate-900/70 p-3 hover:border-sky-500/50'
          >
            <input
              type='checkbox'
              checked={visibility[layer.type]}
              onChange={() => onToggle(layer.type)}
              className='mt-1'
            />

            <span>
              <span className='block text-sm font-medium'>{layer.label}</span>
              <span className='mt-1 block text-xs text-slate-400'>
                {layer.description}
              </span>
            </span>
          </label>
        ))}
      </div>
    </div>
  )
}
