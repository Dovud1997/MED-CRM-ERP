import { NextRequest } from 'next/server';
import { clinicResponse } from '@/lib/server-api';
export async function PATCH(request:NextRequest,{params}:{params:Promise<{id:string}>}){const {id}=await params;return clinicResponse(request,`/appointments/${id}/status`,{method:'PATCH',body:await request.text()});}
