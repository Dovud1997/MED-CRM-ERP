import { NextRequest } from 'next/server';
import { clinicResponse } from '@/lib/server-api';
export async function PUT(request:NextRequest,{params}:{params:Promise<{id:string}>}){const {id}=await params;return clinicResponse(request,`/patients/${id}/clinical-profile`,{method:'PUT',body:await request.text()});}
