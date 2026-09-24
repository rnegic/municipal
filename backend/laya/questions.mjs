export const questions = {
  category: {
    type: "choice",
    instructions: "К какой категории относится жалоба жильца многоквартирного дома?",
    criteria: {
      WATER_HEAT: "водоснабжение и отопление: нет горячей или холодной воды, протечка труб, холодные батареи, прорвало стояк",
      ELECTRICITY: "электричество: нет света, искрит проводка, не горит освещение в подъезде, выбивает автомат",
      ELEVATOR: "лифт: лифт не работает, застрял, не открываются двери",
      CLEANING_YARD: "уборка, двор, подъезд, вандализм: мусор, грязь, не убран снег, граффити, сломаны почтовые ящики",
      BUILDING_STRUCTURE: "конструктив здания: протекает крыша, трещины в стенах, осыпается фасад, разрушены ступени или балкон",
      CITY_TERRITORY: "городская территория за пределами двора: ямы на дороге, открытый люк на проезжей части, остановки, уличные фонари",
    },
  },
};

const KEYWORDS = {
  WATER_HEAT: /(?<![а-яё])вод[аыуе]|водоснаб|(?<![а-яё])кран|смесител|труб|стояк|батаре|радиатор|отоплен|канализ|засор|протеч|протека|затопил|полотенцесуш|бойлер|напор|тепл[оа](?![а-яё])|горячк|вентил|унитаз/i,
  ELECTRICITY: /электр|(?<![а-яё])свет(?!офор|л)|ламп|(?<![а-яё])провод|розетк|щит[ко]|щитк|автомат|напряжени|искрит|замыкан|пробк|плафон|кабел|освещ|(?<![а-яё])гар(ь|ью)|домофон/i,
  ELEVATOR: /лифт|подъёмник|подъемник|кабин(?!ет)/i,
  CLEANING_YARD: /мусор|грязн|убор|уборк|не моют|снег|налед|лёд|гололед|сосул|граффити|разрисова|нарисова|качел|надпис|почтов|(детск|игров|контейнерн|спортивн)[а-яё]* площадк|газон|трав[ауы]|крыс|тарака|дезинс|дератиз|воня/i,
  BUILDING_STRUCTURE: /(?<![а-яё])крыш(?!к)|кровл|чердак|трещин|фасад|стен[аыуе]|штукатур|балкон|ступен|перил|окн[оа]|двер[ьи] (в )?подъезд|входн|козыр[её]к|фундамент|отмостк|панельн|плесен/i,
  CITY_TERRITORY: /(?<![а-яё])дорог(?!ие|ой|ая|ую|ого)|проезж|(?<![а-яё])ям[ау]|люк|тротуар|остановк|светофор|уличн|фонар|перекр[её]ст|(?<![а-яё])переход(?!ит|ят|ить)|сквер|(?<![а-яё])парк(?!ов|ет)|улиц|обочин|ливн[её]в|набережн|колод(ец|ц)/i,
};
const KEYWORD_BOOST = 100;
const CALIBRATION_POWER = 0.5;
const NEUTRAL_TEXTS = ["жалоба", "N/A", "текст обращения"];
const LABELS = Object.keys(questions.category.criteria);

const toState = (text) => ({ "жалоба жильца": text });

const normalize = (weights) => {
  const sum = LABELS.reduce((s, l) => s + weights[l], 0);
  return Object.fromEntries(LABELS.map((l) => [l, weights[l] / sum]));
};

const rawProbabilities = async (laya, text) => (await laya.systemOne(toState(text), questions)).answers.category.probabilities;

export const createClassifier = async (laya) => {
  const prior = Object.fromEntries(LABELS.map((l) => [l, 0]));
  for (const text of NEUTRAL_TEXTS) {
    const p = await rawProbabilities(laya, text);
    for (const l of LABELS) prior[l] += p[l] / NEUTRAL_TEXTS.length;
  }
  return async (text) => {
    const p = await rawProbabilities(laya, text);
    const probabilities = normalize(
      Object.fromEntries(LABELS.map((l) => [l, (p[l] / prior[l] ** CALIBRATION_POWER) * (KEYWORDS[l].test(text) ? KEYWORD_BOOST : 1)])),
    );
    const choice = LABELS.reduce((best, l) => (probabilities[l] > probabilities[best] ? l : best));
    return { choice, probabilities };
  };
};

export const loadOptions = () => ({
  modelDir: process.env.LAYA_MODEL_DIR || undefined,
  sessionOptions: { intraOpNumThreads: Number(process.env.LAYA_THREADS ?? 2) },
});
