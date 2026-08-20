import { NextRequest } from "next/server";
import { clinicResponse } from "@/lib/server-api";
export async function PATCH(r:NextRequest,{params}:{params:Promise<{id:string}>}){const{id}=await params;return clinicResponse(r,`/laboratory/tests/${id}`,{method:"PATCH",body:await r.text()})}
