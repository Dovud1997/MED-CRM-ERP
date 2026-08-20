"use client";
import{useEffect,useState}from"react";
type Employee={id:string;firstName:string;lastName:string;isActive:boolean};
export function ResponsibleEmployeeSelect({name="responsibleEmployeeId",label="Ответственный сотрудник"}:{name?:string;label?:string}){const[employees,setEmployees]=useState<Employee[]>([]);useEffect(()=>{void fetch("/api/clinic/employees",{cache:"no-store"}).then(async r=>{if(r.ok)setEmployees((await r.json()).items??[])})},[]);return <label>{label}<select name={name}><option value="">— Не назначен —</option>{employees.filter(x=>x.isActive).map(x=><option key={x.id} value={x.id}>{x.lastName} {x.firstName}</option>)}</select></label>}
