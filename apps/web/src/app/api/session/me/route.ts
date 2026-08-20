import { NextRequest, NextResponse } from 'next/server';
const api=process.env.INTERNAL_API_URL??'http://api:4000/api/v1';
export async function GET(request:NextRequest){
  let access=request.cookies.get('obk_access')?.value;
  let upstream=await fetch(`${api}/auth/me`,{headers:{authorization:`Bearer ${access??''}`},cache:'no-store'});
  let rotated:{accessToken:string;refreshToken:string}|null=null;
  if(upstream.status===401){const refresh=request.cookies.get('obk_refresh')?.value;if(refresh){const r=await fetch(`${api}/auth/refresh`,{method:'POST',headers:{'content-type':'application/json'},body:JSON.stringify({refreshToken:refresh}),cache:'no-store'});if(r.ok){rotated=await r.json();access=rotated!.accessToken;upstream=await fetch(`${api}/auth/me`,{headers:{authorization:`Bearer ${access}`},cache:'no-store'});}}}
  if(!upstream.ok)return NextResponse.json({authenticated:false},{status:401});
  const response=NextResponse.json({authenticated:true,user:await upstream.json()});
  if(rotated){const secure=process.env.COOKIE_SECURE==='true';response.cookies.set('obk_access',rotated.accessToken,{httpOnly:true,sameSite:'strict',secure,path:'/',maxAge:15*60});response.cookies.set('obk_refresh',rotated.refreshToken,{httpOnly:true,sameSite:'strict',secure,path:'/api/session',maxAge:30*24*60*60});}
  return response;
}
