'use client'
import { useQuery, type UseQueryResult } from '@tanstack/react-query'
import { APIRequestError } from '@/lib/api/client'
import { getWeatherContext } from '@/lib/api/weather-context'
import { defaultProjectionDurationSeconds } from '@/lib/queries/projection-intelligence'
import type { WeatherContextResponse } from '@/types/weather-context'
const keys={all:['weather-context'] as const,byRequest:(id:string|null,time:string|null,duration:number)=>[...keys.all,id,time,duration] as const}
export function useWeatherContext(trajectoryID:string|null,asOfTime:string|null,duration=defaultProjectionDurationSeconds):UseQueryResult<WeatherContextResponse,Error>{const id=normalize(trajectoryID);const time=normalize(asOfTime);return useQuery({queryKey:keys.byRequest(id,time,duration),queryFn:({signal})=>{if(id===null||time===null)throw new APIRequestError('Weather Context requires a trajectory and an analytical timestamp.');return getWeatherContext({trajectoryID:id,asOfTime:time,durationSeconds:duration,signal})},enabled:id!==null&&time!==null,staleTime:60_000,refetchInterval:id===null||time===null?false:120_000,refetchIntervalInBackground:false,retry:(count,error)=>count<2&&(!(error instanceof APIRequestError)||error.status===null||error.status>=500)})}
function normalize(value:string|null){const result=value?.trim()??'';return result===''?null:result}
