import { NextRequest } from 'next/server';import { clinicResponse } from '@/lib/server-api';
export const dynamic='force-dynamic';
export async function GET(request:NextRequest){return clinicResponse(request,'/employees');}
export async function POST(request:NextRequest){return clinicResponse(request,'/employees',{method:'POST',body:await request.text()});}
