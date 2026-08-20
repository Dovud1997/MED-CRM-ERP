import { NextRequest } from 'next/server';import { clinicResponse } from '@/lib/server-api';
export const dynamic='force-dynamic';
export async function GET(request:NextRequest){return clinicResponse(request,'/branches');}
export async function POST(request:NextRequest){return clinicResponse(request,'/branches',{method:'POST',headers:{'content-type':'application/json'},body:await request.text()});}
