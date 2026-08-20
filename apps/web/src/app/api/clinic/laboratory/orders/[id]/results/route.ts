import { NextRequest } from "next/server";
import { clinicResponse } from "@/lib/server-api";
export async function POST(r:NextRequest,{params}:{params:Promise<{id:string}>}){const{id}=await params;return clinicResponse(r,`/laboratory/orders/${id}/results`,{method:"POST",body:await r.text()})}
