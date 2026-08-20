import { NextRequest } from "next/server";
import { clinicResponse } from "@/lib/server-api";
export async function PUT(request:NextRequest,{params}:{params:Promise<{id:string}>}){const{id}=await params;return clinicResponse(request,`/services/${encodeURIComponent(id)}/providers`,{method:"PUT",headers:{"content-type":"application/json"},body:await request.text()});}
