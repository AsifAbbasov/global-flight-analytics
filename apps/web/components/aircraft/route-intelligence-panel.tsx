'use client'

import { APIRequestError, getRequestErrorMessage } from '@/lib/api/client'
import type {
  RouteIntelligenceConfidenceLevel,
  RouteIntelligenceEndpoint,
  RouteIntelligenceRecord,
  RouteIntelligenceStatus,
} from '@/types/route-intelligence'

interface Props {
  selectedICAO24: string | null
  trajectoryID: string | null
  record: RouteIntelligenceRecord | undefined
  isPending: boolean
  isFetching: boolean
  error: Error | null
  onRetry: () => void
}

export function RouteIntelligencePanel({ selectedICAO24, trajectoryID, record, isPending, isFetching, error, onRetry }: Props) {
  if (selectedICAO24 === null) return null
  const notFound = error instanceof APIRequestError && error.status === 404
  return (
    <aside className='rounded-xl border border-slate-700 bg-slate-950/95 p-5' aria-labelledby='route-intelligence-title'>
      <div className='flex items-start justify-between gap-4'>
        <div><p className='text-xs font-semibold uppercase tracking-[0.18em] text-emerald-300'>Production analysis</p><h3 id='route-intelligence-title' className='mt-2 text-lg font-semibold text-white'>Route Intelligence</h3><p className='mt-1 text-xs leading-5 text-slate-400'>Versioned, validated and persisted route inference.</p></div>
        {isFetching ? <span className='text-xs text-sky-300'>Updating…</span> : null}
      </div>
      {record ? <Content record={record} /> : null}
      {trajectoryID === null && !error ? <p className='mt-4 rounded-lg border border-slate-800 bg-slate-900/60 p-3 text-sm leading-6 text-slate-400'>Waiting for a persisted trajectory identifier before running production Route Intelligence.</p> : null}
      {trajectoryID !== null && isPending && !error ? <p className='mt-4 text-sm leading-6 text-slate-400'>Resolving airport candidates, validating endpoint evidence and storing the current Route Intelligence result…</p> : null}
      {notFound ? <p className='mt-4 rounded-lg border border-slate-700 bg-slate-900/70 p-3 text-sm leading-6 text-slate-300'>Production Route Intelligence is unavailable because the persisted trajectory could not be found.</p> : null}
      {error && !notFound ? <div className='mt-4 rounded-lg border border-amber-400/30 bg-amber-400/10 p-3'><p className='text-sm leading-6 text-amber-100'>{getRequestErrorMessage(error)}</p><button type='button' onClick={onRetry} disabled={isFetching} className='mt-3 rounded-md border border-amber-300/40 px-3 py-1.5 text-sm font-medium text-amber-100 disabled:opacity-60'>Retry Route Intelligence</button></div> : null}
    </aside>
  )
}

function Content({ record }: { record: RouteIntelligenceRecord }) {
  const r=record.result
  return <>
    <div className='mt-4 rounded-lg border border-slate-800 bg-slate-900/70 p-3'>
      <div className='flex flex-wrap items-center justify-between gap-3'><Status status={r.status}/><Confidence level={r.confidence.level} score={r.confidence.score}/></div>
      <div className='mt-3 h-2 overflow-hidden rounded-full bg-slate-800' role='progressbar' aria-label='Route Intelligence confidence score' aria-valuemin={0} aria-valuemax={100} aria-valuenow={Math.round(r.confidence.score*100)}><div className='h-full rounded-full bg-emerald-400' style={{width:`${r.confidence.score*100}%`}}/></div>
      <dl className='mt-3 grid grid-cols-2 gap-x-4 gap-y-3 text-sm'><Detail label='Route distance' value={r.status==='complete'?distance(r.summary.great_circle_distance_km):'Unavailable'}/><Detail label='Evidence records' value={String(r.confidence.evidence_count)}/><Detail label='Analytical as of' value={date(r.window.as_of_time)}/><Detail label='Stored' value={date(record.stored_at)}/></dl>
      <p className='mt-3 break-all font-mono text-[11px] leading-5 text-slate-500'>{r.schema_version} · {record.id}</p>
    </div>
    <div className='mt-3 grid gap-3'><Endpoint label='Resolved origin' endpoint={r.origin}/><Endpoint label='Resolved destination' endpoint={r.destination}/></div>
    <div className='mt-4 rounded-lg border border-slate-800 bg-slate-900/60 p-3'><h4 className='text-xs font-semibold uppercase tracking-wide text-slate-300'>Confidence evidence</h4>{r.confidence.reasons.length?<ul className='mt-2 space-y-2 text-sm leading-5 text-slate-400'>{r.confidence.reasons.map(x=><li key={x.code}>{x.message} <span className='text-slate-500'>({signed(x.contribution)})</span></li>)}</ul>:<p className='mt-2 text-sm text-slate-400'>No route-level confidence reasons were reported.</p>}</div>
    <div className='mt-4 rounded-lg border border-amber-400/25 bg-amber-400/5 p-3'><h4 className='text-xs font-semibold uppercase tracking-wide text-amber-200'>Production limitations</h4>{r.limitations.length?<ul className='mt-2 space-y-2 text-sm leading-5 text-amber-100'>{r.limitations.slice(0,6).map(x=><li key={`${x.scope}:${x.code}`}>{x.message}</li>)}</ul>:<p className='mt-2 text-sm text-slate-300'>No production Route Intelligence limitations were reported.</p>}</div>
    <div className='mt-4 text-xs leading-5 text-slate-500'><p>Resolver: {r.provenance.resolver_version}. Sources: {r.provenance.source_names.join(', ')||'Unknown'}.</p><p className='mt-1 break-all font-mono'>{r.provenance.input_fingerprint}</p></div>
  </>
}
function Endpoint({label,endpoint}:{label:string;endpoint:RouteIntelligenceEndpoint|undefined}){if(!endpoint)return <div className='rounded-lg border border-dashed border-slate-700 bg-slate-900/40 p-3'><p className='text-xs uppercase tracking-wide text-slate-500'>{label}</p><p className='mt-2 text-sm leading-5 text-slate-400'>No endpoint passed the production evidence threshold.</p></div>;const code=[endpoint.airport.icao_code,endpoint.airport.iata_code].filter(Boolean).join(' / ');return <div className='rounded-lg border border-slate-800 bg-slate-900/60 p-3'><div className='flex items-start justify-between gap-3'><div><p className='text-xs uppercase tracking-wide text-slate-500'>{label}</p><p className='mt-1 text-sm font-semibold text-white'>{endpoint.airport.name}</p><p className='mt-1 font-mono text-xs text-emerald-300'>{code}</p></div><Confidence level={endpoint.confidence.level} score={endpoint.confidence.score}/></div><dl className='mt-3 grid grid-cols-2 gap-x-4 gap-y-3 text-sm'><Detail label='Location' value={[endpoint.airport.city,endpoint.airport.country].filter(Boolean).join(', ')}/><Detail label='Endpoint distance' value={distance(endpoint.distance_km)}/><Detail label='Evidence' value={String(endpoint.evidence.length)}/><Detail label='Timezone' value={endpoint.airport.timezone}/></dl></div>}
function Status({status}:{status:RouteIntelligenceStatus}){const c=status==='complete'?'border-emerald-400/40 bg-emerald-400/10 text-emerald-200':status==='partial'?'border-amber-400/40 bg-amber-400/10 text-amber-200':'border-slate-600 bg-slate-800 text-slate-300';return <span className={`rounded-full border px-2.5 py-1 text-xs font-semibold uppercase tracking-wide ${c}`}>{status}</span>}
function Confidence({level,score}:{level:RouteIntelligenceConfidenceLevel;score:number}){const c=level==='high'?'border-emerald-400/40 bg-emerald-400/10 text-emerald-200':level==='medium'?'border-sky-400/40 bg-sky-400/10 text-sky-200':level==='low'?'border-amber-400/40 bg-amber-400/10 text-amber-200':'border-slate-600 bg-slate-800 text-slate-300';return <span className={`rounded-full border px-2.5 py-1 text-xs font-semibold uppercase tracking-wide ${c}`}>{level} · {Math.round(score*100)}%</span>}
function Detail({label,value}:{label:string;value:string}){return <div><dt className='text-xs uppercase tracking-wide text-slate-500'>{label}</dt><dd className='mt-1 break-words text-slate-200'>{value.trim()||'Unknown'}</dd></div>}
function distance(v:number){return Number.isFinite(v)?`${new Intl.NumberFormat(undefined,{maximumFractionDigits:1}).format(v)} km`:'Unknown'}
function date(v:string){const d=new Date(v);return Number.isNaN(d.getTime())?'Unknown':d.toLocaleString()}
function signed(v:number){const n=Math.round(v*100);return `${n>0?'+':''}${n}%`}
