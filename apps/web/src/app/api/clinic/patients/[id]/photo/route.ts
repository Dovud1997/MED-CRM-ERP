import { NextRequest } from "next/server";
import { clinicResponse } from "@/lib/server-api";
export const dynamic="force-dynamic";
export async function GET(request:NextRequest,{params}:{params:Promise<{id:string}>}){const{id}=await params;return clinicResponse(request,`/patients/${id}/photo`)}
export async function POST(request:NextRequest,{params}:{params:Promise<{id:string}>}){const{id}=await params;return clinicResponse(request,`/patients/${id}/photo`,{method:"POST",body:await request.text()})}
