import { NextRequest } from "next/server"; import { clinicResponse } from "@/lib/server-api"; export const dynamic="force-dynamic";
export async function POST(r:NextRequest,{params}:{params:Promise<{id:string}>}){const{id}=await params;return clinicResponse(r,`/inpatient/rooms/${id}/beds`,{method:"POST",body:await r.text()})}
