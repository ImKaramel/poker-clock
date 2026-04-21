// src/utils/tournamentImages.ts
import defaultTournamentImage from "../assets/grand_opening.jpg";

// Импортируем картинки для каждого турнира
import midnight_phoenix from "../assets/MIDNIGHT Phoenix.jpg";
import deep_stack from "../assets/DEEP STACK.jpg";
import freezout from "../assets/FREEZOUT.jpg";
import midnight_gift from "../assets/MIDNIGHT Gift.jpg";

// Создаем маппинг названий турниров на импортированные картинки
const TOURNAMENT_IMAGES_MAP: Record<string, string> = {
  "MIDNIGHT Phoenix": midnight_phoenix,
  "DEEP STACK": deep_stack,
  "FREEZOUT": freezout,
  "MIDNIGHT Gift": midnight_gift,
};

// Функция для получения картинки по названию турнира
export const getTournamentImage = (tournamentName: string): string => {
  // Проверяем прямое совпадение
  if (TOURNAMENT_IMAGES_MAP[tournamentName]) {
    return TOURNAMENT_IMAGES_MAP[tournamentName];
  }

  // Нормализуем имя для поиска (на случай небольших различий в написании)
  const normalizedName = tournamentName
    .trim()
    .toLowerCase();

  // Ищем совпадение среди ключей с нормализацией
  const matchingKey = Object.keys(TOURNAMENT_IMAGES_MAP).find(key => 
    key.trim().toLowerCase() === normalizedName
  );

  if (matchingKey) {
    return TOURNAMENT_IMAGES_MAP[matchingKey];
  }

  // Если картинка не найдена, возвращаем дефолтную
  console.warn(`Картинка для турнира "${tournamentName}" не найдена, используется дефолтная`);
  return defaultTournamentImage;
};

// Опционально: функция для проверки, есть ли кастомная картинка для турнира
export const hasCustomTournamentImage = (tournamentName: string): boolean => {
  return tournamentName in TOURNAMENT_IMAGES_MAP;
};