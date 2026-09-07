/* Спільний довідник моделей для картки і реєстру.
   Тримаємо в одному місці: раніше він дублювався в двох сторінках,
   і після додавання нових моделей реєстр показував сирий ключ. */
var MODELS={
  m3t:'DJI Mavic 3T',
  m3ta:'DJI Mavic 3TA',
  m3p:'DJI Mavic 3 Pro',
  m30t:'DJI Matrice 30T',
  m4t:'DJI Matrice 4T',
  m4e:'DJI Matrice 4E',
  evo:'Autel EVO MAX 4T'
};
var SHORT={
  m3t:'MAVIC 3T',
  m3ta:'MAVIC 3TA',
  m3p:'MAVIC 3 PRO',
  m30t:'MATRICE 30T',
  m4t:'MATRICE 4T',
  m4e:'MATRICE 4E',
  evo:'EVO MAX 4T'
};
/* ── Стани картки. Єдине джерело: картка, стікер і реєстр беруть
   назви звідси, щоб той самий стан усюди звався однаково. ──────── */

/* Висновок — рішення щодо виробу. */
var VERDICT=[
  {k:'ind',t:'Незалежна діагностика',short:'ДІАГНОСТИКА',mk:'sq'},
  {k:'fix',t:'Підлягає ремонту',     short:'РЕМОНТ',     mk:'cir'},
  {k:'out',t:'Не підлягає ремонту',  short:'НЕ ПІДЛЯГАЄ',mk:'x'}
];
/* Підстава — офіційний документ, на основі якого прийнято виріб. */
var BASIS={t:'Офіційна підстава',short:'ОФІЦІЙНО',mk:'tri'};

/* Статус — результат виконаних робіт. */
var STATUS=[['fix','Відновлено'],['out','Списано']];
/* ── Стани картки. Єдине джерело: картка, стікер і реєстр беруть
   назви звідси, щоб той самий стан усюди звався однаково. ──────── */

/* Висновок — рішення щодо виробу. */
var VERDICT=[
  {k:'ind',t:'Незалежна діагностика',short:'ДІАГНОСТИКА',mk:'sq'},
  {k:'fix',t:'Підлягає ремонту',     short:'РЕМОНТ',     mk:'cir'},
  {k:'out',t:'Не підлягає ремонту',  short:'НЕ ПІДЛЯГАЄ',mk:'x'}
];
/* Підстава — офіційний документ, на основі якого прийнято виріб. */
var BASIS={t:'Офіційна підстава',short:'ОФІЦІЙНО',mk:'tri'};

/* Статус — результат виконаних робіт. */
var STATUS=[['fix','Відновлено'],['out','Списано']];
var STLBL={fix:'ВІДНОВЛЕНО',out:'СПИСАНО'};

/* Старі картки зберігали висновок за позицією (vd.0..vd.3) зі списку
   Діагностика / Ремонт / Заміна вузлів / Списання. Розкладаємо в нові
   іменовані ключі. Спільна для картки й реєстру, бо реєстр читає
   checks напряму з бази. */
function migrateChecks(list){
  if(!list||!list.length) return list||[];
  function isOld(k){ return k==='vd.0'||k==='vd.1'||k==='vd.2'||k==='vd.3'; }
  var has=false,i;
  for(i=0;i<list.length;i++) if(isOld(list[i])) has=true;
  /* Навіть без старих ключів треба зняти суперечність: обидва рішення
     могли потрапити з імпорту чи давнішої версії. */
  if(!has) return dropConflict(list.slice());

  var out=[];
  for(i=0;i<list.length;i++) if(!isOld(list[i])) out.push(list[i]);
  /* Ремонт і Заміна вузлів -> підлягає ремонту; Списання -> не підлягає.
     Стара «Діагностика» не означала незалежної, тож не переносимо. */
  if((list.indexOf('vd.1')>=0||list.indexOf('vd.2')>=0)&&out.indexOf('vd.fix')<0)
    out.push('vd.fix');
  if(list.indexOf('vd.3')>=0&&out.indexOf('vd.out')<0)
    out.push('vd.out');
  return dropConflict(out);
}

/* Обидва рішення разом — суперечність; лишаємо суворіше «не підлягає». */
function dropConflict(list){
  if(list.indexOf('vd.fix')>=0&&list.indexOf('vd.out')>=0)
    return list.filter(function(k){ return k!=='vd.fix'; });
  return list;
}

/* Назва висновку за ключем. */
function verdictName(k,short){
  var v=VERDICT.filter(function(x){ return x.k===k; })[0];
  return v?(short?v.short:v.t):k;
}

/* Назва моделі: повна або скорочена. Ніколи не показує сирий ключ —
   якщо моделі немає в довіднику, повертає сам ключ у верхньому регістрі. */
function modelName(k,short){
  if(!k) return '';
  return (short?SHORT[k]:MODELS[k])||String(k).toUpperCase();
}
/* Останні 4 знаки S/N — для компактного показу. */
function snShort(sn){
  sn=String(sn||'');
  return sn.length>4?sn.slice(-4):sn;
}
