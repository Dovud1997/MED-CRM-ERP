import {NextRequest} from 'next/server';import{clinicResponse}from'@/lib/server-api';
export async function GET(request:NextRequest,{params}:{params:Promise<{id:string}>}){const{id}=await params;return clinicResponse(request,`/patients/${encodeURIComponent(id)}/clinical`);}
