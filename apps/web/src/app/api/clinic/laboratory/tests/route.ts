import { NextRequest } from "next/server";
import { clinicResponse } from "@/lib/server-api";
export const dynamic="force-dynamic";
export async function GET(r:NextRequest){return clinicResponse(r,"/laboratory/tests")}
export async function POST(r:NextRequest){return clinicResponse(r,"/laboratory/tests",{method:"POST",body:await r.text()})}
