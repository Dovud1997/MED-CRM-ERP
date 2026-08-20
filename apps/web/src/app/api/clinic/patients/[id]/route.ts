import { NextRequest } from 'next/server';import { clinicResponse } from '@/lib/server-api';
export async function GET(request:NextRequest,{params}:{params:Promise<{id:string}>}){const{id}=await params;return clinicResponse(request,`/patients/${encodeURIComponent(id)}`);}
export async function PATCH(request:NextRequest,{params}:{params:Promise<{id:string}>}){const{id}=await params;return clinicResponse(request,`/patients/${encodeURIComponent(id)}`,{method:'PATCH',body:await request.text()});}
export async function DELETE(request:NextRequest,{params}:{params:Promise<{id:string}>}){const{id}=await params;return clinicResponse(request,`/patients/${encodeURIComponent(id)}`,{method:'DELETE'});}
