import { NextRequest } from 'next/server';
import { clinicResponse } from '@/lib/server-api';
export async function POST(request:NextRequest,{params}:{params:Promise<{id:string}>}){const {id}=await params;return clinicResponse(request,`/patients/${id}/vaccinations`,{method:'POST',body:await request.text()});}
