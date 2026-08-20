import { NextRequest, NextResponse } from 'next/server';

const api = process.env.INTERNAL_API_URL ?? 'http://api:4000/api/v1';
const organizationId = process.env.DEFAULT_ORGANIZATION_ID ?? '';

export async function POST(request: NextRequest) {
  const body = await request.json() as { login?: string; password?: string };
  const upstream = await fetch(`${api}/auth/login`, { method:'POST', headers:{'content-type':'application/json','user-agent':request.headers.get('user-agent')??''}, body:JSON.stringify({ organizationId, login:body.login, password:body.password }), cache:'no-store' });
  const data = await upstream.json();
  if (!upstream.ok) return NextResponse.json(data,{status:upstream.status});
  const response = NextResponse.json({ success:true, mustChangePassword:Boolean(data.mustChangePassword) });
  const secure=process.env.COOKIE_SECURE==='true';
  response.cookies.set('obk_access',data.accessToken,{httpOnly:true,sameSite:'strict',secure,path:'/',maxAge:15*60});
  response.cookies.set('obk_refresh',data.refreshToken,{httpOnly:true,sameSite:'strict',secure,path:'/api/session',maxAge:30*24*60*60});
  return response;
}
