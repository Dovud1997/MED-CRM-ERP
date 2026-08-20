import {NextRequest} from 'next/server';import{clinicResponse}from'@/lib/server-api';
export async function POST(request:NextRequest,{params}:{params:Promise<{id:string}>}){const{id}=await params;return clinicResponse(request,`/patients/${encodeURIComponent(id)}/history`,{method:'POST',body:await request.text()});}
