import { NextRequest } from "next/server"; import { clinicResponse } from "@/lib/server-api"; export const dynamic="force-dynamic";
export async function GET(r:NextRequest){return clinicResponse(r,"/inpatient/bookings")}
export async function POST(r:NextRequest){return clinicResponse(r,"/inpatient/bookings",{method:"POST",body:await r.text()})}
