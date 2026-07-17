'use client'

import { useState, type ChangeEvent } from 'react'
import { AircraftDetailPanel } from '@/components/aircraft/aircraft-detail-panel'
import { ProjectionIntelligencePanel } from '@/components/aircraft/projection-intelligence-panel'
import { RouteIntelligencePanel } from '@/components/aircraft/route-intelligence-panel'
import { TrafficGlobe } from '@/components/globe/traffic-globe'
import { TrafficMap } from '@/components/map/traffic-map'
import { getRequestErrorMessage } from '@/lib/api/client'
import { useAircraftRouteContext } from '@/lib/queries/route-context'
import { useProcessedRouteIntelligence } from '@/lib/queries/route-intelligence'
import { useProjectionIntelligence } from '@/lib/queries/projection-intelligence'
import { useCurrentTraffic } from '@/lib/queries/traffic'
import { useLatestAircraftTrajectory } from '@/lib/queries/trajectory'
import type { Region } from '@/types/region'
import type { TrafficAircraft } from '@/types/traffic'

interface TrafficDashboardProps { regions:Region[]; selectedRegion:Region; onSelectedRegionCodeChange:(regionCode:string)=>void; initialTraffic:TrafficAircraft[]; initialError:string|null; regionsWarning:string|null }

export function TrafficDashboard({regions,selectedRegion,onSelectedRegionCodeChange,initialTraffic,initialError,regionsWarning}:TrafficDashboardProps){
  const [selectedAircraftICAO24,setSelectedAircraftICAO24]=useState<string|null>(null)
  const initialData=selectedRegion.code==='world'&&initialError===null?initialTraffic:undefined
  const trafficQuery=useCurrentTraffic(selectedRegion.code,{initialData})
  const routeContextQuery=useAircraftRouteContext(selectedAircraftICAO24)
  const trajectoryQuery=useLatestAircraftTrajectory(selectedAircraftICAO24)
  const routeIntelligenceTrajectoryID=routeContextQuery.data?.trajectory_id??null
  const routeIntelligenceQuery=useProcessedRouteIntelligence(routeIntelligenceTrajectoryID)
  const projectionTrajectoryID=trajectoryQuery.data?.id??routeIntelligenceTrajectoryID
  const projectionAsOfTime=trajectoryQuery.data?.end_time??null
  const projectionQuery=useProjectionIntelligence(projectionTrajectoryID,projectionAsOfTime)
  const traffic=trafficQuery.data??[]
  const selectedAircraft=selectedAircraftICAO24===null?undefined:traffic.find((item: TrafficAircraft)=>normalizeICAO24(item.icao24)===selectedAircraftICAO24)
  const isInitialLoading=trafficQuery.isPending
  const isRefreshing=trafficQuery.isFetching&&!trafficQuery.isPending
  const errorMessage=trafficQuery.error?getRequestErrorMessage(trafficQuery.error):trafficQuery.isPending?initialError:null
  return <>
    <section className='mt-6 rounded-xl border border-slate-800 bg-slate-900 p-4'><div className='flex flex-wrap items-end justify-between gap-4'><div className='min-w-64 flex-1'><label className='block text-sm font-medium text-slate-300' htmlFor='traffic-region'>Region</label><select id='traffic-region' value={selectedRegion.code} onChange={(event: ChangeEvent<HTMLSelectElement>)=>{setSelectedAircraftICAO24(null);onSelectedRegionCodeChange(event.target.value)}} className='mt-2 w-full rounded-lg border border-slate-700 bg-slate-950 px-3 py-2 text-white'>{regions.map(region=><option key={region.code} value={region.code}>{region.name}</option>)}</select></div><button type='button' onClick={()=>{void trafficQuery.refetch()}} disabled={trafficQuery.isFetching} className='rounded-lg border border-slate-700 px-4 py-2 text-sm font-medium text-slate-200 disabled:opacity-60'>{trafficQuery.isFetching?'Refreshing traffic…':'Refresh traffic'}</button></div><div aria-live='polite' className='mt-3 flex flex-wrap items-center gap-3 text-sm'><span className='text-slate-300'>Aircraft: {traffic.length}</span><span className='text-slate-500'>View: {selectedRegion.name}</span>{selectedAircraftICAO24?<span className='text-sky-300'>Selected: {selectedAircraftICAO24.toUpperCase()}</span>:null}{trafficQuery.dataUpdatedAt>0?<span className='text-slate-500'>Updated {formatTimestamp(trafficQuery.dataUpdatedAt)}</span>:null}{regionsWarning?<span className='text-amber-300'>{regionsWarning}</span>:null}{isInitialLoading?<span className='text-sky-300'>Loading current traffic…</span>:null}{isRefreshing?<span className='text-sky-300'>Updating current traffic…</span>:null}{errorMessage?<><span className='text-amber-300'>{errorMessage}</span><button type='button' onClick={()=>{void trafficQuery.refetch()}} disabled={trafficQuery.isFetching} className='rounded-md border border-amber-400/50 px-3 py-1 font-medium text-amber-200 disabled:opacity-60'>Retry</button></>:null}</div></section>
    <div className='mt-4' aria-busy={trafficQuery.isFetching}><TrafficGlobe aircraft={traffic} region={selectedRegion}/></div>
    <section className='mt-8 rounded-xl border border-slate-800 bg-slate-900 p-4 sm:p-6'><h2 className='text-xl font-semibold'>Current Traffic — {selectedRegion.name}</h2><p className='mt-2 text-sm text-slate-400'>Select an aircraft marker to inspect its live state, preliminary route context, validated production Route Intelligence, research Projection Intelligence, registered profile, persisted trajectory and quality limitations.</p><div className='mt-4 grid gap-4 xl:grid-cols-[minmax(0,1fr)_480px]'><div aria-busy={trafficQuery.isFetching}><TrafficMap aircraft={traffic} region={selectedRegion} selectedAircraftICAO24={selectedAircraftICAO24} trajectory={trajectoryQuery.data} onSelectAircraft={(icao24: string)=>setSelectedAircraftICAO24(normalizeICAO24(icao24))}/></div><div className='space-y-4'><AircraftDetailPanel selectedICAO24={selectedAircraftICAO24} aircraft={selectedAircraft} routeContext={routeContextQuery.data} routeContextIsPending={routeContextQuery.isPending} routeContextIsFetching={routeContextQuery.isFetching} routeContextError={routeContextQuery.error} onRetryRouteContext={()=>{void routeContextQuery.refetch()}} trajectory={trajectoryQuery.data} trajectoryIsPending={trajectoryQuery.isPending} trajectoryIsFetching={trajectoryQuery.isFetching} trajectoryError={trajectoryQuery.error} onRetryTrajectory={()=>{void trajectoryQuery.refetch()}} onClose={()=>setSelectedAircraftICAO24(null)}/><RouteIntelligencePanel selectedICAO24={selectedAircraftICAO24} trajectoryID={routeIntelligenceTrajectoryID} record={routeIntelligenceQuery.data} isPending={routeIntelligenceQuery.isPending} isFetching={routeIntelligenceQuery.isFetching} error={routeIntelligenceQuery.error} onRetry={()=>{void routeIntelligenceQuery.refetch()}}/><ProjectionIntelligencePanel selectedICAO24={selectedAircraftICAO24} trajectoryID={projectionTrajectoryID} result={projectionQuery.data} isPending={projectionQuery.isPending} isFetching={projectionQuery.isFetching} error={projectionQuery.error} onRetry={()=>{void projectionQuery.refetch()}}/></div></div>{!trafficQuery.isFetching&&!errorMessage&&traffic.length===0?<p className='mt-4 text-sm text-slate-400'>No aircraft were returned for the selected region.</p>:null}</section>
  </>
}
function normalizeICAO24(value:string){return value.trim().toLowerCase()}
function formatTimestamp(value:number){return new Intl.DateTimeFormat(undefined,{hour:'2-digit',minute:'2-digit',second:'2-digit'}).format(new Date(value))}
