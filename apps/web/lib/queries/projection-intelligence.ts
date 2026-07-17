'use client'

import { useQuery, type UseQueryResult } from '@tanstack/react-query'
import { APIRequestError } from '@/lib/api/client'
import { getProjectionIntelligence } from '@/lib/api/projection-intelligence'
import type { ProjectionIntelligenceResponse } from '@/types/projection-intelligence'

export const defaultProjectionDurationSeconds = 300
const keys = { all: ['projection-intelligence'] as const, byRequest: (id:string|null,asOf:string|null,duration:number)=>[...keys.all,id,asOf,duration] as const }

export function useProjectionIntelligence(trajectoryID:string|null,asOfTime:string|null,durationSeconds=defaultProjectionDurationSeconds):UseQueryResult<ProjectionIntelligenceResponse,Error>{
  const id=normalize(trajectoryID)
  const asOf=normalize(asOfTime)
  return useQuery({
    queryKey:keys.byRequest(id,asOf,durationSeconds),
    queryFn:({signal})=>{if(id===null||asOf===null)throw new APIRequestError('Projection Intelligence requires a trajectory and an analytical timestamp.');return getProjectionIntelligence({trajectoryID:id,asOfTime:asOf,durationSeconds,signal})},
    enabled:id!==null&&asOf!==null,
    staleTime:30_000,
    refetchInterval:id===null||asOf===null?false:60_000,
    refetchIntervalInBackground:false,
    retry:(count,error)=>count<2&&(!(error instanceof APIRequestError)||error.status===null||error.status>=500),
  })
}
function normalize(value:string|null):string|null{const result=value?.trim()??'';return result===''?null:result}
