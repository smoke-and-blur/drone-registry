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
var STLBL={fix:'ВІДНОВЛЕНО',out:'СПИСАНО'};

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
