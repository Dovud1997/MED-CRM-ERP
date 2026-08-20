import { NextRequest } from "next/server"; import { clinicResponse } from "@/lib/server-api"; export const dynamic="force-dynamic";
export async function DELETE(r:NextRequest,{params}:{params:Promise<{id:string}>}){const{id}=await params;return clinicResponse(r,`/inpatient/beds/${id}`,{method:"DELETE"})}
